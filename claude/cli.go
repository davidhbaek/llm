package claude

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type env struct {
	client       Client
	userPrompt   string
	systemPrompt string
	startChat    bool
}

func (app *env) fromArgs(args []string) error {
	err := godotenv.Load("../../.env")
	if err != nil {
		return err
	}

	fl := flag.NewFlagSet("claude", flag.ContinueOnError)

	userPrompt := fl.String("prompt", "", "the user prompt to send to Claude")
	systemPrompt := fl.String("system", "", "the system prompt to send to Claude")
	startChat := fl.Bool("chat", false, "chat")
	app.client = *NewClient(NewConfig("https://api.anthropic.com", os.Getenv("ANTHROPIC_API_KEY")))

	if err := fl.Parse(args); err != nil {
		return err
	}

	app.userPrompt = *userPrompt
	app.systemPrompt = *systemPrompt
	app.startChat = *startChat

	return nil
}

func CLI(args []string) int {
	app := env{}
	err := app.fromArgs(args)
	if err != nil {
		return 2
	}

	if err := app.run(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		return 1

	}
	return 0
}

func (app *env) run() error {
	messages := []Message{
		{Role: "user", Content: []Content{&Text{Type: "text", Text: app.userPrompt}}},
	}

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
