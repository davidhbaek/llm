package claude

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/pprof"
	"runtime/trace"
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
		return fmt.Errorf("loading environment: %w", err)
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

	var docs fileList
	fl.Var(&docs, "d", "list of filepaths to docs (PDFs)")
	fl.Var(&docs, "document", "list of filepaths to docs (PDFs)")

	var isChat bool
	fl.BoolVar(&isChat, "c", false, "Start a live chat that retains conversation history")
	fl.BoolVar(&isChat, "chat", false, "Start a live chat that retains conversation history")

	if err := fl.Parse(args); err != nil {
		return fmt.Errorf("parsing command line arguments: %w", err)
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
	app.docs = docs
	app.isChat = isChat
	app.client = *NewClient(claudeModel, NewConfig("https://api.anthropic.com", os.Getenv("ANTHROPIC_API_KEY")))

	return nil
}

func (app *env) run() error {
	cpuProfile, err := os.Create("cpu.prof")
	if err != nil {
		return fmt.Errorf("creating CPU profile: %w", err)
	}
	defer cpuProfile.Close()

	err = pprof.StartCPUProfile(cpuProfile)
	if err != nil {
		return fmt.Errorf("starting CPU profile: %w", err)
	}
	defer pprof.StopCPUProfile()

	traceFile, err := os.Create("trace.out")
	if err != nil {
		return fmt.Errorf("creating trace file: %w", err)
	}
	defer traceFile.Close()

	err = trace.Start(traceFile)
	if err != nil {
		return fmt.Errorf("starting trace: %w", err)
	}
	defer trace.Stop()

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
		return fmt.Errorf("sending message to Claude: %w", err)
	}

	body := json.NewDecoder(rsp.Body)
	switch rsp.StatusCode {
	case http.StatusOK:
		okRsp := OkResponseBody{}
		err := body.Decode(&okRsp)
		if err != nil {
			return err
		}

		log.Printf("Answer: %s\n", okRsp.Content[0].Text)
	default:
		errRsp := ErrResponseBody{}
		err := body.Decode(&errRsp)
		if err != nil {
			return err
		}

		return fmt.Errorf("HTTP %d error from Claude API: %+v\n", rsp.StatusCode, errRsp.Error)

	}

	// TODO: Log usage/cost?

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

		body := json.NewDecoder(rsp.Body)
		switch rsp.StatusCode {
		case http.StatusOK:
			okRsp := OkResponseBody{}
			err := body.Decode(&okRsp)
			if err != nil {
				return err
			}

			log.Printf("Answer: %s\n", okRsp.Content[0].Text)

			chatHistory = append(chatHistory, Message{Role: "assistant", Content: []Content{&Text{Type: "text", Text: okRsp.Content[0].Text}}})
		default:
			errRsp := ErrResponseBody{}
			err := body.Decode(&errRsp)
			if err != nil {
				return err
			}

			return fmt.Errorf("HTTP %d error from Claude API: %+v\n", rsp.StatusCode, errRsp.Error)

		}

	}
}

func wrapInXMLTags(text, tag string) string {
	return fmt.Sprintf("<%s>%s</%s>", tag, text, tag)
}
