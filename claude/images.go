package claude

import (
	"flag"
	"fmt"
	"io"
	"net/http"
)

type imagesList []string

var _ flag.Value = &imagesList{}

func (i *imagesList) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *imagesList) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func downloadImageFromURL(url string) ([]byte, error) {
	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	return io.ReadAll(rsp.Body)
}
