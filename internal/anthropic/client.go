package anthropic

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

type Client struct {
	config     *Config
	model      string
	httpClient *http.Client
}

func NewClient(model string) *Client {
	return &Client{
		config: &Config{
			baseURL: "https://api.anthropic.com",
			apiKey:  os.Getenv("ANTHROPIC_API_KEY"),
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
	reqBody, err := json.Marshal(struct {
		Model        string         `json:"model"`
		MaxTokens    int            `json:"max_tokens"`
		SystemPrompt string         `json:"system"`
		Messages     []wire.Message `json:"messages"`
		Stream       bool           `json:"stream"`
	}{
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

		// The Anthropic API doesn't return a msgType key:value pair for errors
		// Only a JSON body
		if strings.Contains(line, "error") {
			errRsp := struct {
				Type  string `json:"type"`
				Error struct {
					Type    string `json:"type"`
					Message string `json:"message"`
				} `json:"error"`
			}{}

			err := json.Unmarshal([]byte(line), &errRsp)
			if err != nil {
				return "", err
			}

			fmt.Printf("error from Anthropic API: type=%s message=%s", errRsp.Error.Type, errRsp.Error.Message)
			return "", nil
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		msgType, payload := parts[0], parts[1]
		switch msgType {
		case "event":
		case "data":

			sseData := SSEData{}
			err := json.Unmarshal([]byte(payload), &sseData)
			if err != nil {
				return "", err
			}

			switch sseData.Type {
			case "content_block_delta":
				content := ContentBlockDelta{}
				err := json.Unmarshal([]byte(payload), &content)
				if err != nil {
					return "", err
				}
				fmt.Printf("%s", content.Delta.Text)
				text += content.Delta.Text
			}

		}

	}

	fmt.Println()

	return text, nil
}
