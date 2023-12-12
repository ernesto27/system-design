package urlfilter

import (
	"net/http"
	"time"
)

type URLFilter struct {
	url string
}

func New(url string) *URLFilter {
	return &URLFilter{
		url: url,
	}
}

func (u *URLFilter) IsValid(url string) bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	response, err := client.Get(url)
	if err != nil {
		return true
	}
	defer response.Body.Close()

	return response.StatusCode != 200
}
