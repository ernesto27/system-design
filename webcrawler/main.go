package main

import (
	"fmt"
	"webcrawler/contentparser"
	"webcrawler/htmldownloader"
	"webcrawler/seedurl"
)

func main() {

	seed := seedurl.New()

	for _, u := range seed.Urls {
		fmt.Println(u)

		h := htmldownloader.New(u)
		html, err := h.Download()
		if err != nil {
			panic(err)
		}

		c := contentparser.New(html)
		valid := c.IsValidHTML()
		fmt.Println(valid)

	}
}
