package claude

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"
)

const maxSize = 5 * 1024 * 1024 // 5 MB (the max Claude allows per image)

type fileList []string

var _ flag.Value = &fileList{}

func (f *fileList) String() string {
	return fmt.Sprintf("%v", *f)
}

func (f *fileList) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func DownloadImage(path string) ([]byte, error) {
	buffer := bytes.Buffer{}

	// Download the image from the internet
	if strings.HasPrefix(path, "https://") {
		rsp, err := http.Get(path)
		if err != nil {
			return nil, err
		}
		defer rsp.Body.Close()

		_, err = io.Copy(&buffer, rsp.Body)
		if err != nil {
			return nil, err
		}

		return buffer.Bytes(), nil

		// Download the image from a local file patha
	} else {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}

		fileInfo, err := os.Stat(path)
		if err != nil {
			return nil, err
		}

		// We immediately send the bytes back if we don't need to resize the image
		if fileInfo.Size() <= maxSize {
			_, err := io.Copy(&buffer, file)
			if err != nil {
				return nil, fmt.Errorf("copying image bytes: %w", err)
			}
			return buffer.Bytes(), nil
		}

		log.Printf("re-sizing photo at path=%s", path)

		targetSize := float64(maxSize) / float64(fileInfo.Size())

		switch filepath.Ext(path) {
		case "jpg":
			img, err := jpeg.Decode(file)
			if err != nil {
				return nil, err
			}

			err = jpeg.Encode(&buffer, resizeImg(img, targetSize), &jpeg.Options{Quality: jpeg.DefaultQuality})
			if err != nil {
				fmt.Println("Error creating output file:", err)
				return nil, err
			}

		case "png":
			img, err := png.Decode(file)
			if err != nil {
				return nil, err
			}

			encoder := png.Encoder{CompressionLevel: png.BestCompression}
			err = encoder.Encode(&buffer, resizeImg(img, targetSize))
			if err != nil {
				fmt.Println("Error creating output file:", err)
				return nil, err
			}

		}

	}

	return buffer.Bytes(), nil
}

func resizeImg(img image.Image, size float64) image.Image {
	width := uint(float64(img.Bounds().Dx()) * size)
	height := uint(float64(img.Bounds().Dy()) * size)

	return resize.Resize(width, height, img, resize.Lanczos3)
}
