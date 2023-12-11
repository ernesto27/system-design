package main

import (
	"bytes"
	"compress/gzip"
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

	seed := seedurl.New()

	for _, u := range seed.Urls {
		go func(u string) {
			err := job(u, db, mq)
			if err != nil {
				fmt.Println(err)
			}
		}(u)
	}

	messages := make(chan []byte)
	errors := make(chan error)
	go mq.Consumer(messages, errors)

	for {
		select {
		case msg := <-messages:
			log.Printf("Received a message: %s", msg)
			err := job(string(msg), db, mq)
			if err != nil {
				fmt.Println(err)
			}
		case err := <-errors:
			log.Println(err)
		}
	}
}

func job(url string, db *db.SQLite, mq *messagequeue.Rabbit) error {
	h := htmldownloader.New(url)
	html, err := h.Download()
	if err != nil {
		return err
	}

	hash := contentseen.New(html).CreateHash()
	fmt.Println(hash)
	c := contentparser.New(html)
	if c.IsValidHTML() {
		err := db.CreateLink(url, hash, html)
		if err != nil {
			fmt.Println(err)
		}

		links := linkextractor.New(url, html).GetLinks()
		fmt.Println("Links from:", url)
		for _, l := range links {
			fmt.Println(l)
			// Send message to queue
			err := mq.Producer(l)
			if err != nil {
				fmt.Println(err)
			}

		}
	}

	return nil
}

func compressHTML(inputHTML string) (string, error) {
	var compressedBuffer bytes.Buffer

	// Create a gzip writer
	writer := gzip.NewWriter(&compressedBuffer)

	// Write the HTML content to the gzip writer
	_, err := writer.Write([]byte(inputHTML))
	if err != nil {
		return "", err
	}

	// Close the gzip writer to flush any remaining data
	err = writer.Close()
	if err != nil {
		return "", err
	}

	// Return the compressed HTML as a base64-encoded string
	return compressedBuffer.String(), nil
}
