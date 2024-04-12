package openai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/davidhbaek/llm/internal/llm"
)

type Client struct {
	model      string
	httpClient *http.Client
}

func NewClient(model string) *Client {
	return &Client{
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

var _ llm.Client = &Client{}

func (c *Client) SendMessage(messages []llm.Message, systemPrompt string) (*llm.Response, error) {
	reqBody, err := json.Marshal(struct {
		Model    string        `json:"model"`
		Messages []llm.Message `json:"messages"`
		Stream   bool          `json:"stream"`
	}{
		Model:    "gpt-3.5-turbo",
		Messages: messages,
		Stream:   true,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", "https://api.openai.com", "/v1/chat/completions"), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("OPENAI_API_KEY")))

	rsp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return &llm.Response{
		StatusCode: rsp.StatusCode,
		Body:       rsp.Body,
	}, nil
}

func ReadBody(body io.Reader) (string, error) {
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
