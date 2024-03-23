package claude

import (
	"encoding/base64"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type env struct {
	client       Client
	userPrompt   string
	systemPrompt string
	images       imagesList
	startChat    bool
}

func (app *env) fromArgs(args []string) error {
	err := godotenv.Load()
	if err != nil {
		return err
	}

	fl := flag.NewFlagSet("claude", flag.ContinueOnError)

	userPrompt := fl.String("prompt", "", "the user prompt to send to Claude")
	systemPrompt := fl.String("system", "", "the system prompt to send to Claude")

	images := imagesList{}
	fl.Var(&images, "image", "list of image paths (filenames and URLs)")

	startChat := fl.Bool("chat", false, "chat")
	app.client = *NewClient(NewConfig("https://api.anthropic.com", os.Getenv("ANTHROPIC_API_KEY")))

	if err := fl.Parse(args); err != nil {
		return err
	}

	app.userPrompt = *userPrompt
	app.systemPrompt = *systemPrompt
	app.startChat = *startChat
	app.images = images

	return nil
}

func CLI(args []string) int {
	app := env{}
	err := app.fromArgs(args)
	if err != nil {
		fmt.Printf("parsing args: %s", err.Error())
		return 2
	}

	if err := app.run(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		return 1

	}
	return 0
}

func (app *env) run() error {
	// Prepare the request to Claude
	// Download the images from any filepath/URL provided
	// Convert the image to a base64 string
	// Append any user prompts and send the request payload to Claude
	content := []Content{}
	for _, path := range app.images {
		if strings.HasPrefix(path, "https://") {
			imgBytes, err := downloadImageFromURL(path)
			if err != nil {
				return err
			}

			contentType := http.DetectContentType(imgBytes)
			imgBase64String := base64.StdEncoding.EncodeToString(imgBytes)

			content = append(content, &Image{
				Type: "image",
				Source: Source{
					Type: "base64", MediaType: contentType, Data: imgBase64String,
				},
			})

		} else {
			// The image file is locally stored
		}
	}

	// Claude works better if the image comes before
	content = append(content, &Text{Type: "text", Text: app.userPrompt})

	messages := []Message{{Role: "user", Content: content}}
	rspBytes, err := app.client.CreateMessage(messages, app.systemPrompt)
	if err != nil {
		return err
	}

	fmt.Println(string(rspBytes))
	// type response struct {
	// 	ID      string                   `json:"id"`
	// 	Type    string                   `json:"type"`
	// 	Role    string                   `json:"role"`
	// 	Content []map[string]interface{} `json:"content"`
	// }

	// rsp := response{}
	// err = json.Unmarshal(rspBytes, &rsp)
	// if err != nil {
	// 	return err
	// }

	// spew.Dump(rsp)
	return nil
}
