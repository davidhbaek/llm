package llm

import (
	"bufio"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/davidhbaek/llm/internal/anthropic"
	"github.com/davidhbaek/llm/internal/wire"
)

type env struct {
	client       Client
	userPrompt   string
	systemPrompt string
	images       fileList
	isChat       bool
	docs         fileList
}

type fileList []string

var _ flag.Value = &fileList{}

func (f *fileList) String() string {
	return fmt.Sprintf("%v", *f)
}

func (f *fileList) Set(value string) error {
	*f = append(*f, value)
	return nil
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

const (
	OPUS   = "claude-3-opus-20240229"
	SONNET = "claude-3-sonnet-20240229"
	HAIKU  = "claude-3-haiku-20240307"
	GPT4   = "gpt-4-turbo"
)

func (app *env) fromArgs(args []string) error {
	fl := flag.NewFlagSet("claude", flag.ContinueOnError)

	var prompt string
	fl.StringVar(&prompt, "p", "", "user prompt to Claude")
	fl.StringVar(&prompt, "prompt", "", "user prompt to Claude")

	var system string
	fl.StringVar(&system, "s", "", "system prompt to  Claude")
	fl.StringVar(&system, "system", "", "system prompt to  Claude")

	var inputModel string
	fl.StringVar(&inputModel, "m", "haiku", "the Claude model to use")
	fl.StringVar(&inputModel, "model", "haiku", "the Claude model to use")

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

	models := map[string]string{
		"haiku":  HAIKU,
		"sonnet": SONNET,
		"opus":   OPUS,
		"gpt4":   GPT4,
	}

	model, ok := models[inputModel]
	if !ok {
		return errors.New("input model must be one of [haiku, sonnet, opus, gpt4]")
	}

	app.client = setupClient(model)

	// Get the prompt text if they're coming from a file
	if filepath.Ext(prompt) == ".txt" {
		log.Printf("reading prompt file at path=%s", prompt)
		bytes, err := os.ReadFile(prompt)
		if err != nil {
			return err
		}

		prompt = string(bytes)
	}

	if filepath.Ext(system) == ".txt" {
		log.Printf("reading system file at path=%s", system)
		bytes, err := os.ReadFile(system)
		if err != nil {
			return err
		}

		system = string(bytes)
	}

	app.userPrompt = prompt
	app.systemPrompt = system
	app.images = images
	app.docs = docs
	app.isChat = isChat

	return nil
}

func (app *env) run() error {
	content := []wire.Content{&wire.Text{Type: "text", Text: app.userPrompt}}

	for _, path := range app.images {
		// TODO: Re-factor to make this model agnostic
		// For images, GPT-4 models just need the image URL
		if app.client.Model() == GPT4 {
			content = append(content, &wire.OpenAIImage{
				Type: "image_url",
				ImageURL: struct {
					URL string `json:"url"`
				}{URL: path},
			})
			// While Anthropic requires a base64 encoded string of the image bytes
		} else {
			imgBytes, err := anthropic.DownloadImage(path)
			if err != nil {
				return err
			}

			content = append(content, &wire.AnthropicImage{
				Type: "image",
				Source: struct {
					Type      string `json:"type"`
					MediaType string `json:"media_type"`
					Data      string `json:"data"`
				}{
					Type:      "base64",
					MediaType: http.DetectContentType(imgBytes),
					Data:      base64.StdEncoding.EncodeToString(imgBytes),
				},
			})

		}
	}

	if app.isChat {
		err := app.runChatSession()
		if err != nil {
			return fmt.Errorf("running chat session: %w", err)
		}
	}

	messages := []wire.Message{{Role: "user", Content: content}}
	rsp, err := app.client.SendMessage(messages, app.systemPrompt)
	if err != nil {
		return fmt.Errorf("sending prompt: %w", err)
	}

	_, err = app.client.ReadBody(rsp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	return nil
}

func (app *env) runChatSession() error {
	log.Printf("Beginning chat session with model=%s", app.client.Model())
	chatHistory := []wire.Message{}
	input := bufio.NewReader(os.Stdin)

	for {
		prompt, err := input.ReadString('\n')
		if err != nil {
			return err
		}

		chatHistory = append(chatHistory, wire.Message{Role: "user", Content: []wire.Content{&wire.Text{Type: "text", Text: prompt}}})

		rsp, err := app.client.SendMessage(chatHistory, app.systemPrompt)
		if err != nil {
			return fmt.Errorf("sending chat prompt: %w", err)
		}

		chatRsp, err := app.client.ReadBody(rsp.Body)
		if err != nil {
			return fmt.Errorf("reading chat response body: %w", err)
		}

		chatHistory = append(chatHistory, wire.Message{Role: "assistant", Content: []wire.Content{&wire.Text{Type: "text", Text: chatRsp}}})

	}
}

func setupClient(model string) Client {
	config := NewClientConfig()
	factory, ok := config.Models[model]
	if !ok {
		log.Fatalf("unsupported model: %s", model)
	}

	return factory(model)
}
