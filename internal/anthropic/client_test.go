package anthropic_test

import (
	"net/http"
	"testing"

	"github.com/davidhbaek/llm/internal/anthropic"
	"github.com/davidhbaek/llm/internal/wire"
	"github.com/stretchr/testify/require"
)

func TestSendMessage(t *testing.T) {
	client := anthropic.NewClient("claude-3-haiku-20240307")

	tests := []struct {
		Name               string
		ExpectedStatusCode int
		InputMsg           []wire.Message
		SystemPrompt       string
	}{
		{Name: "Hello Claude", ExpectedStatusCode: http.StatusOK, InputMsg: []wire.Message{{Role: "user", Content: []wire.Content{&wire.Text{Type: "text", Text: "Hello Claude"}}}}},
		{Name: "Empty input should return bad request", ExpectedStatusCode: http.StatusBadRequest, InputMsg: []wire.Message{{}}}, // empty prompt

	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			rsp, err := client.SendMessage(test.InputMsg, test.SystemPrompt)
			require.NoError(t, err)
			require.Equal(t, test.ExpectedStatusCode, rsp.StatusCode)
		})
	}
}
