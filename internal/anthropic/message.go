package anthropic

type Message struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

type Content interface {
	GetType() string
}

type Text struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

var _ Content = &Text{}

func (t *Text) GetType() string {
	return "text"
}

type Image struct {
	Type   string `json:"type"`
	Source Source `json:"source"`
}

var _ Content = &Image{}

func (I *Image) GetType() string {
	return "image"
}

type Source struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}
