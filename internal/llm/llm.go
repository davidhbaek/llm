package llm

import "github.com/davidhbaek/llm/internal/wire"

type Client interface {
	SendMessage(messages []wire.Message, systemPrompt string) (*wire.Response, error)
}
