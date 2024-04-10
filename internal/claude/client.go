package claude

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Model names for the Claude API
// https://docs.anthropic.com/claude/docs/models-overview#model-comparison
const (
	OPUS   = "claude-3-opus-20240229"
	SONNET = "claude-3-sonnet-20240229"
	HAIKU  = "claude-3-haiku-20240307"
)

type Client struct {
	config     *Config
	model      string
	httpClient *http.Client
}

func NewClient(model string, config *Config) *Client {
	return &Client{
		config: config,
		model:  model,
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

func (c *Client) SendMessage(messages []Message, systemPrompt string) (*Response, error) {
	reqBody, err := json.Marshal(Request{
		Model:        c.model,
		MaxTokens:    2048,
		SystemPrompt: systemPrompt,
		Messages:     messages,
		Stream:       true,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", c.config.baseURL, "v1/messages"), bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("sending POST request: %w", err)
	}

	req.Header.Set("x-api-key", c.config.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Accept", "text/event-stream")

	rsp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return &Response{
		StatusCode: rsp.StatusCode,
		Body:       rsp.Body,
	}, nil
}
