package llm

import (
	"github.com/davidhbaek/llm/internal/anthropic"
	"github.com/davidhbaek/llm/internal/openai"
)

type ClientFactory func(model string) Client

type ClientConfig struct {
	Models map[string]ClientFactory
}

func NewClientConfig() *ClientConfig {
	return &ClientConfig{
		Models: map[string]ClientFactory{
			GPT4: func(model string) Client {
				return openai.NewClient(model)
			},
			OPUS: func(model string) Client {
				return anthropic.NewClient(model)
			},
			SONNET: func(model string) Client {
				return anthropic.NewClient(model)
			},
			HAIKU: func(model string) Client {
				return anthropic.NewClient(model)
			},
		},
	}
}
