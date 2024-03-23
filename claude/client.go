package claude

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	config     *Config
	model      string
	httpClient *http.Client
}

func NewClient(config *Config) *Client {
	return &Client{
		config: config,
		model:  "claude-3-haiku-20240307",
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

func (c *Client) CreateMessage(messages []Message, systemPrompt string) ([]byte, error) {
	// resize any images that are sent
	reqBody, err := json.Marshal(requestBody{
		Model:        c.model,
		MaxTokens:    2048,
		SystemPrompt: systemPrompt,
		Messages:     messages,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", c.config.baseURL, "v1/messages"), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-api-key", c.config.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	rsp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	rspBody, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}

	return rspBody, nil
}
