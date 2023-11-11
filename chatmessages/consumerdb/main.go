package main

import (
	"chatmessages/db"
	"chatmessages/messagebroker"
	"chatmessages/types"
	"encoding/json"
	"fmt"
	"log"
)

func main() {
	cassandra, err := db.NewCassandra("127.0.0.1", "chatmessages")
	if err != nil {
		panic(err)
	}
	defer cassandra.Session.Close()

	channels, err := cassandra.GetChannels()
	if err != nil {
		panic(err)
	}

	fmt.Println(channels)

	for _, channel := range channels {
		fmt.Println(channel.ID.String())
		fmt.Println(channel.Offset)
		c := messagebroker.NewConsumer("localhost:9092", channel.ID.String()+"_P", 0, channel.Offset)

		messages := make(chan []byte)
		errors := make(chan error)
		go c.ReadMessages(messages, errors)

		for {
			select {
			case msg := <-messages:
				fmt.Println("save on db" + string(msg))
				var m types.Message
				err := json.Unmarshal(msg, &m)
				if err != nil {
					fmt.Println(err)
					continue
				}

				if err := cassandra.CreateMessage(m); err != nil {
					fmt.Println(err)
					continue
				}

				err = cassandra.UpdateChannelOffset(channel.ID)
				if err != nil {
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

}
