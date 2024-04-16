package llm

import (
	"context"
	"io"

	"github.com/davidhbaek/llm/internal/anthropic"
	"github.com/davidhbaek/llm/internal/openai"
	"github.com/davidhbaek/llm/internal/wire"
)

type Client interface {
	// Define how to send a prompt to the LLMs API
	SendMessage(ctx context.Context, messages []wire.Message, systemPrompt string) (*wire.Response, error)
	// Define how to read the response body from the LLM
	ReadBody(body io.Reader) (string, error)
	// Return the underlying LLM being prompted
	Model() string
}

// Enforce interface compliance
var (
	_ Client = &openai.Client{}
	_ Client = &anthropic.Client{}
)
