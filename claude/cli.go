package claude

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"rsc.io/pdf"
)

type env struct {
	client       Client
	userPrompt   string
	systemPrompt string
	images       fileList
	isChat       bool
	model        string
	docs         fileList
	content      []Content
}

func CLI(args []string) int {
	app := env{}
	err := app.fromArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parsing args: %v\n", err)
		return 2
	}

	if err := app.run(); err != nil {
		fmt.Fprintf(os.Stderr, "runtime error: %v\n", err)
		return 1

	}
	return 0
}

func (app *env) fromArgs(args []string) error {
	err := godotenv.Load()
	if err != nil {
		return err
	}

	fl := flag.NewFlagSet("claude", flag.ContinueOnError)

	var prompt string
	fl.StringVar(&prompt, "p", "", "user prompt to Claude")
	fl.StringVar(&prompt, "prompt", "", "user prompt to Claude")

	var system string
	fl.StringVar(&system, "s", "", "system prompt to  Claude")
	fl.StringVar(&system, "system", "", "system prompt to  Claude")

	var model string
	fl.StringVar(&model, "m", "haiku", "the Claude model to use")
	fl.StringVar(&model, "model", "haiku", "the Claude model to use")

	var images fileList
	fl.Var(&images, "i", "list of image paths (filenames and URLs)")
	fl.Var(&images, "image", "list of image paths (filenames and URLs)")

	// TODO: Accept a list of documents
	var docs fileList
	fl.Var(&docs, "d", "list of filepaths to docs (PDFs)")
	fl.Var(&docs, "document", "list of filepaths to docs (PDFs)")

	var isChat bool
	fl.BoolVar(&isChat, "c", false, "Start a live chat that retains conversation history")
	fl.BoolVar(&isChat, "chat", false, "Start a live chat that retains conversation history")

	if err := fl.Parse(args); err != nil {
		return err
	}

	modelMap := map[string]string{
		"haiku":  HAIKU,
		"sonnet": SONNET,
		"opus":   OPUS,
	}

	claudeModel, ok := modelMap[model]
	if !ok {
		return errors.New("model must be one of [haiku, sonnet, opus]")
	}

	app.userPrompt = prompt
	app.systemPrompt = system
	app.images = images
	app.model = claudeModel
	app.docs = docs
	app.isChat = isChat
	app.client = *NewClient(NewConfig("https://api.anthropic.com", os.Getenv("ANTHROPIC_API_KEY")))

	return nil
}

func (app *env) run() error {
	// Load up any PDFs
	docs := make([]Text, len(app.docs))
	wg := sync.WaitGroup{}
	for i, path := range app.docs {
		log.Println("ingesting this doc:", path)
		wg.Add(1)
		go func(idx int, path string) error {
			var text string
			file, err := pdf.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open file at path=%s: %w", path, err)
			}

			for i := 1; i <= file.NumPage(); i++ {
				textSlice := file.Page(i).Content().Text
				for _, t := range textSlice {
					text += t.S + "\n"
				}
			}

			docs[idx] = Text{
				Type: "text",
				Text: text,
			}
			wg.Done()
			return nil
		}(i, path)

	}

	wg.Wait()

	var docsPrompt string
	for _, doc := range docs {
		d := fmt.Sprintf("%s\n", wrapInXMLTags(doc.Text, "document"))
		docsPrompt += d
	}

	// Load up any images
	content := []Content{}
	for _, path := range app.images {
		imgBytes, err := downloadImage(path)
		if err != nil {
			return err
		}

		content = append(content, &Image{
			Type: "image",
			Source: Source{
				Type:      "base64",
				MediaType: http.DetectContentType(imgBytes),
				Data:      base64.StdEncoding.EncodeToString(imgBytes),
			},
		})

	}

	// Pass any contextual knowledge from the docs to our chatbot session
	if app.isChat {
		err := app.runChatSession(docsPrompt)
		if err != nil {
			return err
		}
	}

	// One off prompting
	// Add the prompt that the user initally provided
	content = append(content, &Text{Type: "text", Text: app.userPrompt})

	messages := []Message{{Role: "user", Content: content}}
	systemPrompt := fmt.Sprintf(wrapInXMLTags(docsPrompt, "documents"), app.systemPrompt)

	rsp, err := app.client.CreateMessage(messages, systemPrompt)
	if err != nil {
		return err
	}

	answer, err := parseResponse(rsp)
	if err != nil {
		return err
	}

	log.Println("Answer: ", answer)

	return nil
}

func (app *env) runChatSession(docsPrompt string) error {
	fmt.Println("Welcome to the chat session")

	chatHistory := []Message{}

	reader := bufio.NewReader(os.Stdin)

	systemPrompt := fmt.Sprintf(wrapInXMLTags(docsPrompt, "documents"), app.systemPrompt)

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		chatHistory = append(chatHistory, Message{Role: "user", Content: []Content{&Text{Type: "text", Text: strings.TrimSpace(input)}}})

		rsp, err := app.client.CreateMessage(chatHistory, systemPrompt)
		if err != nil {
			return err
		}

		answer, err := parseResponse(rsp)
		if err != nil {
			return err
		}

		log.Println("Answer:", answer)

		chatHistory = append(chatHistory, Message{Role: "assistant", Content: []Content{&Text{Type: "text", Text: answer}}})

	}
}

func parseResponse(rspBytes []byte) (string, error) {
	// First parse the 'type' field so we know how to decode the rest of the response
	var rspType string
	rspTypeDecoder := json.NewDecoder(bytes.NewReader(rspBytes))
	rspTypeDecoder.UseNumber()
	for rspTypeDecoder.More() {
		token, err := rspTypeDecoder.Token()
		if err != nil {
			return "", err
		}

		if key, ok := token.(string); ok && key == "type" {
			err := rspTypeDecoder.Decode(&rspType)
			if err != nil {
				return "", err
			}
			break
		}
	}

	reader := bytes.NewReader(rspBytes)
	rspBodyDecoder := json.NewDecoder(reader)
	switch rspType {
	case "error":

		errRsp := struct {
			Type  string `json:"type"`
			Error struct {
				Type    string `json:"type"`
				Message string `json:"message"`
			} `json:"error"`
		}{}

		err := rspBodyDecoder.Decode(&errRsp)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("error from Claude API: %s", errRsp.Error.Message), nil

	case "message":

		okRsp := struct {
			ID           string                   `json:"id"`
			Type         string                   `json:"type"`
			Role         string                   `json:"role"`
			Content      []map[string]interface{} `json:"content"`
			Model        string                   `json:"model"`
			StopReason   string                   `json:"stop_reason"`
			StopSequence string                   `json:"stop_sequence"`
			Usage        map[string]int           `json:"usage"`
		}{}
		err := rspBodyDecoder.Decode(&okRsp)
		if err != nil {
			return "", err
		}

		return okRsp.Content[0]["text"].(string), nil
	default:
		return "", fmt.Errorf("unsupported response type: %s", rspType)
	}
}

func wrapInXMLTags(text, tag string) string {
	return fmt.Sprintf("<%s>%s</%s>", tag, text, tag)
}
