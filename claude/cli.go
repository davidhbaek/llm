package claude

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/davecgh/go-spew/spew"
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
	content := []Content{}

	// Load up any images or docs provided to the LLM
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

	var text string
	for _, path := range app.docs {
		file, err := pdf.Open(path)
		if err != nil {
			return err
		}

		for i := 1; i <= file.NumPage(); i++ {
			textSlice := file.Page(i).Content().Text
			for _, t := range textSlice {
				text += t.S + "\n"
			}
		}

	}
	// Live chat session
	// We'll have the user prompt come from stdin instead of a CLI argument
	if app.isChat {
		err := app.runChatSession(text)
		if err != nil {
			return err
		}
	}

	spew.Dump(app)

	// One off prompting
	// Add the prompt that the user initally provided
	content = append(content, &Text{Type: "text", Text: app.userPrompt})

	messages := []Message{{Role: "user", Content: content}}
	systemPrompt := fmt.Sprintf(wrapInXMLTags(text, "document"), app.systemPrompt)

	rsp, err := app.client.CreateMessage(messages, systemPrompt)
	if err != nil {
		return err
	}

	fmt.Println(string(rsp))

	answer, err := parseResponse(rsp)
	if err != nil {
		return err
	}

	fmt.Println("Answer: ", answer)

	return nil
}

// runChatSession can accept a document text to be used during the chat session
func (app *env) runChatSession(docText string) error {
	fmt.Println("Welcome to the chat session")

	chatHistory := []Message{}

	reader := bufio.NewReader(os.Stdin)

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		chatHistory = append(chatHistory, Message{Role: "user", Content: []Content{&Text{Type: "text", Text: strings.TrimSpace(input)}}})
		systemPrompt := fmt.Sprintf(wrapInXMLTags(docText, "document"), app.systemPrompt)

		rsp, err := app.client.CreateMessage(chatHistory, systemPrompt)
		if err != nil {
			return err
		}

		answer, err := parseResponse(rsp)
		if err != nil {
			return err
		}

		fmt.Println("Answer:", answer)

		chatHistory = append(chatHistory, Message{Role: "assistant", Content: []Content{&Text{Type: "text", Text: answer}}})

	}
}

func parseResponse(rspBytes []byte) (string, error) {
	rsp := struct {
		Type string `json:"type"`
	}{}
	err := json.Unmarshal(rspBytes, &rsp)
	if err != nil {
		return "", err
	}

	switch rsp.Type {
	case "error":
		errRsp := responseError{}
		err := json.Unmarshal(rspBytes, &errRsp)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("error from Claude API: %s", errRsp.Error.Message), nil

	case "message":
		okRsp := responseOK{}
		err := json.Unmarshal(rspBytes, &okRsp)
		if err != nil {
			return "", err
		}

		return okRsp.Content[0]["text"].(string), nil
	default:
		return "", nil
	}
}

func wrapInXMLTags(text, tag string) string {
	return fmt.Sprintf("<%s>%s</%s>", tag, text, tag)
}
