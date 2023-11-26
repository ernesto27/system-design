package htmldownloader

import (
	"fmt"
	"io"
	"net/http"
)

type HTMLDownloader struct {
	url string
}

func New(url string) *HTMLDownloader {
	return &HTMLDownloader{url: url}
}

func (h *HTMLDownloader) Download() (string, error) {
	// Make a GET request to the URL
	response, err := http.Get(h.url)
	if err != nil {
		fmt.Println("Error making GET request:", err)
		return "", err
	}
	defer response.Body.Close()

	// Read the HTML content from the response body
	htmlBytes, err := io.ReadAll(io.Reader(response.Body))
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return "", err
	}

	// Convert the byte slice to a string
	htmlString := string(htmlBytes)

	return htmlString, nil
}
