package claude

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

type requestBody struct {
	Model        string    `json:"model"`
	MaxTokens    int       `json:"max_tokens"`
	SystemPrompt string    `json:"system"`
	Messages     []Message `json:"messages"`
}
