package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/davidhbaek/llm/internal/wire"
)

type Config struct {
	baseURL string
	apiKey  string
}

type Client struct {
	config     Config
	model      string
	httpClient *http.Client
}

func NewClient(model string) *Client {
	return &Client{
		config: Config{
			baseURL: "https://api.openai.com",
			apiKey:  os.Getenv("OPENAI_API_KEY"),
		},
		model: model,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     1 * time.Minute,
			},
		},
	}
}

func (c *Client) Model() string {
	return c.model
}

func (c *Client) SendMessage(ctx context.Context, messages []wire.Message, systemPrompt string) (*wire.Response, error) {
	// The OpenAI API doesn't have a separate field for system prompts like the Anthropic API does
	if len(systemPrompt) > 0 {
		messages = append(messages, wire.Message{
			Role:    "system",
			Content: []wire.Content{&wire.Text{Type: "text", Text: systemPrompt}},
		})
	}

	reqBody, err := json.Marshal(struct {
		Model    string         `json:"model"`
		Messages []wire.Message `json:"messages"`
		Stream   bool           `json:"stream"`
	}{
		Model:    c.model,
		Messages: messages,
		Stream:   true,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", c.config.baseURL, "v1/chat/completions"), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.apiKey))

	rsp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return &wire.Response{
		StatusCode: rsp.StatusCode,
		Body:       rsp.Body,
	}, nil
}

func (c *Client) ReadBody(body io.Reader) (string, error) {
	scanner := bufio.NewScanner(body)

	var text string
	for scanner.Scan() {
		line := scanner.Text()

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		payload := parts[1]
		if strings.Contains(payload, "[DONE]") {
			break
		}

		// TODO: Add error response handling

		response := struct {
			ID                string `json:"id"`
			Object            string `json:"-"`
			Model             string `json:"model"`
			Created           int    `json:"-"`
			SystemFingerPrint string `json:"-"`
			Choices           []struct {
				Index int `json:"index"`
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}{}

		err := json.Unmarshal([]byte(payload), &response)
		if err != nil {
			return "", fmt.Errorf("unmarshaling response from API: %w", err)
		}

		for _, choice := range response.Choices {
			fmt.Printf("%s", choice.Delta.Content)
			text += choice.Delta.Content
		}

	}

	fmt.Println()

	return text, nil
}
