package llm

import (
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/davidhbaek/llm/internal/anthropic"
	"github.com/davidhbaek/llm/internal/openai"
	"github.com/davidhbaek/llm/internal/wire"
)

type env struct {
	client       Client
	model        string
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
	GPT    = "gpt-4-turbo"
)

func (app *env) fromArgs(args []string) error {
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
		"gpt":    GPT,
	}

	modelName, ok := modelMap[model]
	if !ok {
		return errors.New("model must be one of [haiku, sonnet, opus, gpt]")
	}

	if model == "gpt" {
		app.client = openai.NewClient(modelName)
	} else {
		app.client = anthropic.NewClient(modelName)
	}

	app.model = modelName

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
		if app.model != GPT {
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

		} else {
			content = append(content, &wire.OpenAIImage{
				Type: "image_url",
				ImageURL: struct {
					URL string `json:"url"`
				}{
					URL: path,
				},
			})
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
