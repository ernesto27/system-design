package linkextractor

import (
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type LinkExtractor struct {
	website string
	html    string
}

func New(website string, html string) *LinkExtractor {
	return &LinkExtractor{
		website: website,
		html:    html,
	}
}

func (l *LinkExtractor) GetLinks() []string {
	links := []string{}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(l.html))
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, _ := s.Attr("href")
		if strings.HasPrefix(link, "http") {
			links = append(links, link)
		} else {
			if len(link) > 2 {
				links = append(links, l.website+link[1:])
			} else {
				links = append(links, l.website)
			}
		}
	})
	return links
}

func (l *LinkExtractor) GetMetaData() (string, string, error) {
	metaDescription := ""
	metaKeywords := ""

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(l.html))
	if err != nil {
		return metaDescription, metaKeywords, err
	}

	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		if name, _ := s.Attr("name"); strings.ToLower(name) == "description" {
			metaDescription, _ = s.Attr("content")
		} else if name, _ := s.Attr("name"); strings.ToLower(name) == "keywords" {
			metaKeywords, _ = s.Attr("content")
		}
	})

	return metaDescription, metaKeywords, nil
}
