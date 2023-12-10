package htmldownloader

import (
	"io"
	"net/http"
	"time"
)

type HTMLDownloader struct {
	url string
}

func New(url string) *HTMLDownloader {
	return &HTMLDownloader{url: url}
}

func (h *HTMLDownloader) Download() (string, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	response, err := client.Get(h.url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	// Read the HTML content from the response body
	htmlBytes, err := io.ReadAll(io.Reader(response.Body))
	if err != nil {
		return "", err
	}

	// Convert the byte slice to a string
	htmlString := string(htmlBytes)

	return htmlString, nil
}
