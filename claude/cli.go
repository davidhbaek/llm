package claude

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type env struct {
	client       Client
	userPrompt   string
	systemPrompt string
	images       imagesList
	startChat    bool
	model        string
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

	var images imagesList
	fl.Var(&images, "image", "list of image paths (filenames and URLs)")

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
	app.client = *NewClient(NewConfig("https://api.anthropic.com", os.Getenv("ANTHROPIC_API_KEY")))

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

func (app *env) run() error {
	content := []Content{}
	// Download and convert any images provided
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

	content = append(content, &Text{Type: "text", Text: app.userPrompt})
	messages := []Message{{Role: "user", Content: content}}

	rspBytes, err := app.client.CreateMessage(messages, app.systemPrompt)
	if err != nil {
		return err
	}

	type response struct {
		ID      string                   `json:"id"`
		Type    string                   `json:"type"`
		Role    string                   `json:"role"`
		Content []map[string]interface{} `json:"content"`
	}

	rsp := response{}
	err = json.Unmarshal(rspBytes, &rsp)
	if err != nil {
		return err
	}

	fmt.Println(rsp.Content[0]["text"])
	return nil
}
