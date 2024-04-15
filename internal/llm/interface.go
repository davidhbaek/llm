package llm

import (
	"io"

	"github.com/davidhbaek/llm/internal/anthropic"
	"github.com/davidhbaek/llm/internal/openai"
	"github.com/davidhbaek/llm/internal/wire"
)

type Client interface {
	// Define how to send a prompt to the LLMs API
	SendMessage(messages []wire.Message, systemPrompt string) (*wire.Response, error)
	// Define how to read the response body from the LLM
	ReadBody(body io.Reader) (string, error)
}

// Enforce interface compliance for these implementations of the Client interface
var (
	_ Client = &openai.Client{}
	_ Client = &anthropic.Client{}
)
