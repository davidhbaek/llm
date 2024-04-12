package openai_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/davidhbaek/llm/internal/llm"
	"github.com/davidhbaek/llm/internal/openai"

	"github.com/stretchr/testify/require"
)

func TestSendMessage(t *testing.T) {
	// imgURL := "https://wallpapercave.com/wp/ENozMYj.jpg"

	client := openai.NewClient("gpt-3.5-turbo")

	// imgBytes, err := downloadImageFromURL(imgURL)
	// require.NoError(t, err)

	tests := []struct {
		Name               string
		ExpectedStatusCode int
		InputMsg           []llm.Message
		SystemPrompt       string
	}{
		{Name: "Hello World", ExpectedStatusCode: http.StatusOK, InputMsg: []llm.Message{{Role: "user", Content: []llm.Content{&llm.Text{Type: "text", Text: "Hello World"}}}}},
		// {Name: "Empty input should return bad request", ExpectedStatusCode: http.StatusBadRequest, InputMsg: []llm.Message{{}}}, // empty prompt
		// {Name: "Successful prompt about an image", ExpectedStatusCode: http.StatusOK, InputMsg: []llm.Message{
		// 	{Role: "user", Content: []llm.Content{
		// 		&llm.Text{Type: "text", Text: "can you describe this image"},
		// 	}},
		// }},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			rsp, err := client.SendMessage(test.InputMsg, test.SystemPrompt)
			require.NoError(t, err)
			require.Equal(t, test.ExpectedStatusCode, rsp.StatusCode)

			_, err = openai.ReadBody(rsp.Body)
			require.NoError(t, err)
		})
	}
}

func downloadImageFromURL(url string) ([]byte, error) {
	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	return io.ReadAll(rsp.Body)
}
