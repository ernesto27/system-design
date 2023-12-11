package linkextractor

import (
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type ImageExtractor struct {
	url  string
	html string
}

func NewImageExtractor(url string, html string) *ImageExtractor {
	return &ImageExtractor{
		url:  url,
		html: html,
	}
}

func (l *ImageExtractor) GetURLs() []string {
	urls := []string{}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(l.html))
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		url, _ := s.Attr("src")
		urlValid := strings.Split(url, " ")[0]
		if urlValid != "" {
			if strings.HasPrefix(urlValid, "http") {
				urls = append(urls, urlValid)
			} else {
				if len(url) > 2 {
					urls = append(urls, l.url+urlValid[1:])
				} else {
					urls = append(urls, urlValid)
				}
			}
		}
	})
	return urls
}
