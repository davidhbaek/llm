package claude_test

import (
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/davidhbaek/llm/claude"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func TestCreateMesssage(t *testing.T) {
	imgURL := "https://wallpapercave.com/wp/ENozMYj.jpg"

	err := godotenv.Load("../.env")
	require.NoError(t, err)

	client := claude.NewClient(
		claude.HAIKU,
		claude.NewConfig(
			"https://api.anthropic.com",
			os.Getenv("ANTHROPIC_API_KEY")),
	)

	imgBytes, err := downloadImageFromURL(imgURL)
	require.NoError(t, err)

	contentType := http.DetectContentType(imgBytes)
	imgBase64String := base64.StdEncoding.EncodeToString(imgBytes)

	messages := []claude.Message{
		{Role: "user", Content: []claude.Content{&claude.Text{Type: "text", Text: "hello claude how are you?"}}},
		{Role: "assistant", Content: []claude.Content{&claude.Text{Type: "text", Text: "How are you doing today?"}}},
		{Role: "user", Content: []claude.Content{
			&claude.Text{Type: "text", Text: "can you describe this image"},
			&claude.Image{Type: "image", Source: claude.Source{Type: "base64", MediaType: contentType, Data: imgBase64String}},
		}},
	}
	rsp, err := client.CreateMessage(messages, "")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rsp.StatusCode)
}

func downloadImageFromURL(url string) ([]byte, error) {
	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	return io.ReadAll(rsp.Body)
}
