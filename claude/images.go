package claude

import (
	"bytes"
	"flag"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"
)

const maxSize = 5 * 1024 * 1024 // 5 MB in bytes

type fileList []string

var _ flag.Value = &fileList{}

func (f *fileList) String() string {
	return fmt.Sprintf("%v", *f)
}

func (f *fileList) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func downloadImage(path string) ([]byte, error) {
	data := []byte{}
	buffer := bytes.Buffer{}

	if strings.HasPrefix(path, "https://") {
		rsp, err := http.Get(path)
		if err != nil {
			return nil, err
		}
		defer rsp.Body.Close()

		data, err = io.ReadAll(rsp.Body)
		if err != nil {
			return nil, err
		}
	} else {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}

		fileInfo, err := os.Stat(path)
		if err != nil {
			return nil, err
		}

		if fileInfo.Size() <= maxSize {
			data, err = io.ReadAll(file)
			if err != nil {
				return nil, err
			}
		}

		switch filepath.Ext(path) {
		case "jpg":
			img, err := jpeg.Decode(file)
			if err != nil {
				return nil, err
			}

			targetSize := float64(maxSize) / float64(fileInfo.Size())
			targetWidth := uint(float64(img.Bounds().Dx()) * targetSize)
			targetHeight := uint(float64(img.Bounds().Dy()) * targetSize)

			resizedImg := resize.Resize(targetWidth, targetHeight, img, resize.Lanczos3)

			err = jpeg.Encode(&buffer, resizedImg, &jpeg.Options{Quality: jpeg.DefaultQuality})
			if err != nil {
				fmt.Println("Error creating output file:", err)
				return nil, err
			}

			data, err = io.ReadAll(&buffer)
			if err != nil {
				return nil, err
			}

		case "png":
			img, err := png.Decode(file)
			if err != nil {
				return nil, err
			}

			targetSize := float64(maxSize) / float64(fileInfo.Size())
			targetWidth := uint(float64(img.Bounds().Dx()) * targetSize)
			targetHeight := uint(float64(img.Bounds().Dy()) * targetSize)

			resizedImg := resize.Resize(targetWidth, targetHeight, img, resize.Lanczos3)

			encoder := png.Encoder{CompressionLevel: png.BestCompression}
			err = encoder.Encode(&buffer, resizedImg)
			if err != nil {
				fmt.Println("Error creating output file:", err)
				return nil, err
			}

			data, err = io.ReadAll(&buffer)
			if err != nil {
				return nil, err
			}

		}

	}

	return data, nil
}
