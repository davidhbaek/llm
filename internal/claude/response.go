package claude

import "io"

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

var _ ResponseBody = &OkResponseBody{}

func (r *OkResponseBody) GetType() string {
	return "message"
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type ErrResponseBody struct {
	Type  string   `json:"type"`
	Error APIError `jsson:"error"`
}

var _ ResponseBody = &ErrResponseBody{}

func (r *ErrResponseBody) GetType() string {
	return "error"
}

type APIError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
