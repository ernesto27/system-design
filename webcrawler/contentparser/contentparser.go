package contentparser

import (
	"strings"

	"golang.org/x/net/html"
)

type ContentParser struct {
	html string
}

func New(html string) *ContentParser {
	return &ContentParser{html: html}
}

func (c *ContentParser) IsValidHTML() bool {
	// TODO: improve this
	_, err := html.Parse(strings.NewReader(c.html))
	return err == nil
}
