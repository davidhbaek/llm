// Package wire holds types that represent anything that goes across a boundary
// Think I/O operations
package wire

import "io"

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
	Source struct {
		URL string `json:"url"`
	} `json:"image_url"`
}

var _ Content = &Image{}

func (I *Image) GetType() string {
	return "image"
}

type AnthropicImage struct {
	Type   string `json:"type"`
	Source struct {
		Type      string `json:"type"`
		MediaType string `json:"media_type"`
		Data      string `json:"data"`
	} `json:"source"`
}

var _ Content = &Image{}

func (I *AnthropicImage) GetType() string {
	return "image"
}

type Response struct {
	StatusCode int
	Body       io.Reader
}
