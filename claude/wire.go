package claude

import "io"

// Wire types (i.e. types that go across a boundary)
// Normally I/O types like request, response
// Whatever you need to define to send a request to another service

type Content interface {
	GetType() string
}

type Text struct {
	Type string `json:"type"` // text | image
	Text string `json:"text"`
}

func (t *Text) GetType() string {
	return "text"
}

type Image struct {
	Type   string `json:"type"`
	Source Source `json:"source"`
}

func (I *Image) GetType() string {
	return "image"
}

type Source struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type Message struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

type Request struct {
	Model         string    `json:"model"`
	Messages      []Message `json:"messages"`
	MaxTokens     int       `json:"max_tokens"`
	SystemPrompt  string    `json:"system"`
	Temperature   float64   `json:"temperature,omitempty"`
	TopP          float64   `json:"top_p,omitempty"`
	TopK          int       `json:"top_k,omitempty"`
	Stream        bool      `json:"stream,omitempty"`
	StopSequences []string  `json:"stop_sequences,omitempty"`
}

type Response struct {
	StatusCode int       `json:"status_code"`
	Body       io.Reader `json:"body"`
}

type ResponseBody interface {
	GetType() string
}

type OkResponseBody struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Role         string `json:"role"`
	Content      []Text `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        Usage  `json:"usage"`
}

type ErrResponseBody struct {
	Type  string   `json:"type"`
	Error APIError `jsson:"error"`
}

type APIError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}
