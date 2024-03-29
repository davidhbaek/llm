package claude

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
