package openai_test

import (
	"net/http"
	"testing"

	"github.com/davidhbaek/llm/internal/openai"
	"github.com/davidhbaek/llm/internal/wire"

	"github.com/stretchr/testify/require"
)

func TestSendMessage(t *testing.T) {
	client := openai.NewClient("gpt-3.5-turbo")
	tests := []struct {
		Name               string
		ExpectedStatusCode int
		InputMsg           []wire.Message
		SystemPrompt       string
	}{
		{Name: "Hello World", ExpectedStatusCode: http.StatusOK, InputMsg: []wire.Message{{Role: "user", Content: []wire.Content{&wire.Text{Type: "text", Text: "Hello World"}}}}},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			rsp, err := client.SendMessage(test.InputMsg, test.SystemPrompt)
			require.NoError(t, err)
			require.Equal(t, test.ExpectedStatusCode, rsp.StatusCode)

			_, err = client.ReadBody(rsp.Body)
			require.NoError(t, err)
		})
	}
}
