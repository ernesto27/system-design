package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"
	"webcrawler/contentparser"
	"webcrawler/contentseen"
	"webcrawler/db"
	"webcrawler/htmldownloader"
	"webcrawler/linkextractor"
	"webcrawler/messagequeue"
	"webcrawler/seedurl"
)

func main() {

	//h := htmldownloader.New("https://i2.wp.com/www.scienceofnoise.net/wp-content/uploads/2021/04/wages-of-sin-5d35f17abe55e.jpg")
	// h := htmldownloader.New("https://musicemall.files.wordpress.com/2010/12/arch-enemy-1_1024_768_i2x.jpg")
	// html, err := h.Download()
	// if err != nil {
	// 	panic(err)
	// }

	// //fmt.Println(html)
	// hash := contentseen.New(html).CreateHash()
	// fmt.Println(hash)
	// return

	// // fmt.Println(html)
	// i := linkextractor.NewImageExtractor("https://www.google.com", html)
	// i.GetURLs()

	// return

	db, err := db.NewPostgresql()
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

func job(url string, db *db.Postgres, mq *messagequeue.Rabbit) error {
	h := htmldownloader.New(url)
	html, err := h.Download()
	if err != nil {
		return err
	}

	hash := contentseen.New(html).CreateHash()
	fmt.Println(hash)
	c := contentparser.New(html)
	if c.IsValidHTML() {
		err := db.CreateLink(url, hash, "")
		if err != nil {
			fmt.Println(err)
		} else {
			go func(h string) {
				err := saveFile(h)
				if err != nil {
					fmt.Println(err)
				}
			}(html)
		}

		// image extractor
		go func() {
			imageURLs := linkextractor.NewImageExtractor(url, html).GetURLs()
			for _, iu := range imageURLs {
				hi := htmldownloader.New(url)
				imgContent, err := hi.Download()
				if err != nil {
					fmt.Println(err)
					continue
				}
				hashImg := contentseen.New(imgContent).CreateHash()
				err = db.CreateImage(url, iu, "", hashImg)
				if err != nil {
					fmt.Println(err)
				}
			}
		}()

		links := linkextractor.New(url, html).GetLinks()
		fmt.Println("Links from:", url)
		for _, l := range links {
			// Send message to queue
			err := mq.Producer(l)
			if err != nil {
				fmt.Println(err)
			}

		}
	}

	return nil
}

func saveFile(html string) error {
	saveDirectory := "./pages/"

	randomFilename := generateRandomFilename()
	filePath := filepath.Join(saveDirectory, randomFilename)

	content := []byte(html)

	err := os.WriteFile(filePath, content, 0644)
	if err != nil {
		return err
	}

	return nil
}

func generateRandomFilename() string {
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(10000)
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%d_%d.html", timestamp, randomNumber)
}
