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
		go func(channel types.Channel) {
			fmt.Println(channel.ID.String())
			fmt.Println(channel.Offset)

			// TODO
			// Check count of messages on topic and compare with offset
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

					uuid, now, err := cassandra.CreateMessage(m)
					if err != nil {
						fmt.Println(err)
						continue
					}

					err = cassandra.UpdateChannelOffset(channel.ID)
					if err != nil {
						fmt.Println(err)
						continue
					}

					// Send message queue consumer
					b, err := messagebroker.NewProducer("localhost:9092", channel.ID.String()+"_C", 0)
					if err != nil {
						fmt.Println(err)
						continue
					}
					m.ID = uuid
					m.CreatedAt = now

					jsonMessage, err := json.Marshal(m)
					if err != nil {
						fmt.Println(err)
						continue
					}

					err = b.Write([]byte(jsonMessage))
					if err != nil {
						fmt.Println(err)
					}
					b.Conn.Close()

				case err := <-errors:
					// Handle error
					fmt.Println(err)
					if err := c.Reader.Close(); err != nil {
						log.Fatal("failed to close reader:", err)
					}
				}
			}
		}(channel)
	}

	// todo fix this
	select {}

}
