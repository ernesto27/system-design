package main

import (
	"chatmessages/db"
	"chatmessages/messagebroker"
	"fmt"
	"log"
	"os"
)

func main() {
	topic := "newtopic"
	partition := 0
	p, err := messagebroker.NewProducer("localhost:9092", topic, partition)
	if err != nil {
		panic(err)
	}

	err = p.Write([]byte("6666"))

	if err != nil {
		log.Fatal("failed to write messages:", err)
	}

	if err := p.Conn.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}

	consumer := messagebroker.NewConsumer("localhost:9092", topic, partition)
	messages := make(chan []byte)
	errors := make(chan error)
	go consumer.ReadMessages(messages, errors)
	for {
		select {
		case msg := <-messages:
			fmt.Println("from loop" + string(msg))
		case err := <-errors:
			// Handle error
			fmt.Println(err)
			if err := consumer.Reader.Close(); err != nil {
				log.Fatal("failed to close reader:", err)
			}
		}
	}

	os.Exit(0)

	// cluster := gocql.NewCluster("127.0.0.1")
	// cluster.Keyspace = "chatmessages"

	// session, err := cluster.CreateSession()
	// if err != nil {
	// 	panic(err)
	// }
	// defer session.Close()

	c, err := db.NewCassandra("127.0.0.1", "chatmessages")
	if err != nil {
		panic(err)
	}

	defer c.Session.Close()

	// uuid, err := gocql.ParseUUID("e63093e5-497c-407b-a391-676ba6d5db2f")
	// if err != nil {
	// 	panic(err)
	// }

	// m := db.Message{
	// 	MessageFrom: uuid,
	// 	MessageTo:   uuid,
	// 	Content:     "Hello World100",
	// }

	// if err := c.CreateMessage(m); err != nil {
	// 	panic(err)
	// }

	// m, err := c.GetMessages()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(m)

	// err = c.CreateUser(db.User{
	// 	Username: "ernesto",
	// 	Password: "password",
	// })

	// if err != nil {
	// 	panic(err)
	// }

	err = c.LoginUser(db.User{
		Username: "ernesto",
		Password: "password",
	})
	if err != nil {
		panic(err)
	}

}
