package claude_test

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/davidhbaek/llm/claude"
	"github.com/stretchr/testify/require"
)

func TestCreateMesssage(t *testing.T) {
	imgURL := "https://wallpapercave.com/wp/ENozMYj.jpg"

	client := claude.NewClient(
		claude.HAIKU,
		claude.NewConfig(
			"https://api.anthropic.com",
			os.Getenv("ANTHROPIC_API_KEY")),
	)

	imgBytes, err := downloadImageFromURL(imgURL)
	require.NoError(t, err)

	fmt.Println(base64.StdEncoding.EncodeToString(imgBytes))

	tests := []struct {
		Name               string
		ExpectedStatusCode int
		InputMsg           []claude.Message
		SystemPrompt       string
	}{
		{Name: "Hello Claude", ExpectedStatusCode: http.StatusOK, InputMsg: []claude.Message{{Role: "user", Content: []claude.Content{&claude.Text{Type: "text", Text: "Hello Claude"}}}}},
		{Name: "Empty input should return bad request", ExpectedStatusCode: http.StatusBadRequest, InputMsg: []claude.Message{{}}}, // empty prompt
		{Name: "Successful prompt about an image", ExpectedStatusCode: http.StatusOK, InputMsg: []claude.Message{
			{Role: "user", Content: []claude.Content{
				&claude.Text{Type: "text", Text: "can you describe this image"},
				&claude.Image{Type: "image", Source: claude.Source{Type: "base64", MediaType: http.DetectContentType(imgBytes), Data: base64.StdEncoding.EncodeToString(imgBytes)}},
			}},
		}},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			rsp, err := client.CreateMessage(test.InputMsg, test.SystemPrompt)
			require.NoError(t, err)
			require.Equal(t, test.ExpectedStatusCode, rsp.StatusCode)
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
