package claude

import (
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
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     1 * time.Minute,
			},
		},
	}
}

func (c *Client) CreateMessage(messages []Message, systemPrompt string) (*Response, error) {
	// response := http.Response{}
	reqBody, err := json.Marshal(Request{
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

	body := bytes.Buffer{}
	chunk := make([]byte, 4096)
	for {
		n, err := rsp.Body.Read(chunk)

		if err == io.EOF {
			err = nil
			// An instance of this general case is that a Reader returning
			// a non-zero number of bytes at the end of the input stream may
			// return either err == EOF or err == nil. The next Read should
			// return 0, EOF.
			if n <= 0 {
				break
			}
		}

		if err != nil {
			return nil, err
		}

		body.Write(chunk[:n])

	}

	return &Response{
		StatusCode: rsp.StatusCode,
		Body:       &body,
	}, nil
}
