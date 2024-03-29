package claude

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

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

func NewClient(config *Config) *Client {
	return &Client{
		config: config,
		model:  OPUS,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

func (c *Client) CreateMessage(messages []Message, systemPrompt string) ([]byte, error) {
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
		return nil, fmt.Errorf("sending POST request: %w", err)
	}

	req.Header.Set("x-api-key", c.config.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	rsp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	body := bufio.NewReader(rsp.Body)
	buffer := bytes.Buffer{}
	chunk := make([]byte, 4096)
	for {
		n, err := body.Read(chunk)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("reading buffer: %w", err)
		}

		buffer.Write(chunk[:n])

	}

	return buffer.Bytes(), nil
}
