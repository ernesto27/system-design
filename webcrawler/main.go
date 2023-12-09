package main

import (
	"fmt"
	"log"
	"webcrawler/contentparser"
	"webcrawler/contentseen"
	"webcrawler/db"
	"webcrawler/htmldownloader"
	"webcrawler/linkextractor"
	"webcrawler/messagequeue"
	"webcrawler/seedurl"
)

func main() {

	db, err := db.New()
	if err != nil {
		panic(err)
	}
	db.Init()

	mq, err := messagequeue.New("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}
	defer mq.Conn.Close()

	//return

	// mq.Producer("Hello World!")
	// return

	seed := seedurl.New()

	for _, u := range seed.Urls {

		h := htmldownloader.New(u)
		html, err := h.Download()
		if err != nil {
			panic(err)
		}

		hash := contentseen.New(html).CreateHash()
		fmt.Println(hash)

		c := contentparser.New(html)
		valid := c.IsValidHTML()
		fmt.Println(valid)

		links := linkextractor.New(u, html).GetLinks()

		fmt.Println("Links from:", u)
		for _, l := range links {
			fmt.Println(l)
			err := db.CreateLink(l)
			if err != nil {
				fmt.Println(err)
			} else {
				// Send message to queue
				err := mq.Producer(l)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}

	messages := make(chan []byte)
	errors := make(chan error)
	go mq.Consumer(messages, errors)

	for {
		select {
		case msg := <-messages:
			log.Printf("Received a message: %s", msg)
			h := htmldownloader.New(string(msg))
			html, err := h.Download()
			if err != nil {
				panic(err)
			}

			hash := contentseen.New(html).CreateHash()
			fmt.Println(hash)

			c := contentparser.New(html)
			valid := c.IsValidHTML()
			fmt.Println(valid)

			links := linkextractor.New(string(msg), html).GetLinks()

			fmt.Println("Links from:", string(msg))
			for _, l := range links {
				fmt.Println(l)
				err := db.CreateLink(l)
				if err != nil {
					fmt.Println(err)
				} else {
					// Send message to queue
					err := mq.Producer(l)
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		case err := <-errors:
			log.Println(err)
		}
	}

}
