package openai_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/davidhbaek/llm/internal/openai"
	"github.com/davidhbaek/llm/internal/wire"

	"github.com/stretchr/testify/require"
)

func TestSendMessage(t *testing.T) {
	client := openai.NewClient("gpt-4-turbo")
	tests := []struct {
		Name               string
		ExpectedStatusCode int
		InputMsg           []wire.Message
		SystemPrompt       string
	}{
		{Name: "Hello ChatGPT", ExpectedStatusCode: http.StatusOK, InputMsg: []wire.Message{{Role: "user", Content: []wire.Content{&wire.Text{Type: "text", Text: "Hello World"}}}}},
		{Name: "send image with prompt", ExpectedStatusCode: http.StatusOK, InputMsg: []wire.Message{{Role: "user", Content: []wire.Content{
			&wire.Text{Type: "text", Text: "What's in this image?"},
			&wire.OpenAIImage{
				Type: "image_url",
				ImageURL: struct {
					URL string `json:"url"`
				}{
					URL: "https://tetonheritagebuilders.com/wp-content/uploads/2016/12/krafty_photos_Aguzin-2.jpg",
				},
			},
		}}}},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			rsp, err := client.SendMessage(context.Background(), test.InputMsg, test.SystemPrompt)
			require.NoError(t, err)
			require.Equal(t, test.ExpectedStatusCode, rsp.StatusCode)

			_, err = client.ReadBody(rsp.Body)
			require.NoError(t, err)
		})
	}
}
