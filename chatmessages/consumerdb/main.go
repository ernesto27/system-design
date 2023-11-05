package main

import (
	"chatmessages/db"
	"chatmessages/messagebroker"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func main() {
	cassandra, err := db.NewCassandra("127.0.0.1", "chatmessages")
	if err != nil {
		panic(err)
	}

	defer cassandra.Session.Close()

	// get latest offset kafka
	offset, err := cassandra.GetConfig(1)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//cassandra.UpdateConfig(0)

	c := messagebroker.NewConsumer("localhost:9092", "chat2", 0, int64(offset))

	messages := make(chan []byte)
	errors := make(chan error)
	go c.ReadMessages(messages, errors)
	for {
		select {
		case msg := <-messages:
			fmt.Println("SAVE ON DB" + string(msg))
			var m db.Message
			err := json.Unmarshal(msg, &m)
			if err != nil {
				fmt.Println(err)
				continue
			}

			if err := cassandra.CreateMessage(m); err != nil {
				fmt.Println(err)
				continue
			}

			offset += 1
			if err := cassandra.UpdateConfig(1, offset); err != nil {
				fmt.Println(err)
				continue
			}

		case err := <-errors:
			// Handle error
			fmt.Println(err)
			if err := c.Reader.Close(); err != nil {
				log.Fatal("failed to close reader:", err)
			}
		}
	}

}
