package main

import (
	"chatmessages/db"
	"chatmessages/messagebroker"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Request struct {
	Message string `json:"message"`
}

type Response struct {
	Message string `json:"message"`
}

type Client struct {
	Conn *websocket.Conn
}

var clients = make(map[*websocket.Conn]Client)

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// TODO AUTH VALIDATION

	fmt.Println("WebSocketHandler")

	// websocker upgrade
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	clients[conn] = Client{Conn: conn}

	// setup producers
	b, err := messagebroker.NewProducer("localhost:9092", "chat2", 0)
	if err != nil {
		panic(err)
	}

	// setup consumers
	c := messagebroker.NewConsumer("localhost:9092", "chat2", 0, 0)
	// listen to messages
	go func() {
		messages := make(chan []byte)
		errors := make(chan error)
		go c.ReadMessages(messages, errors)
		for {
			select {
			case msg := <-messages:
				fmt.Println("CONSUMER READ" + string(msg))
				// send to websocker client
			case err := <-errors:
				// Handle error
				fmt.Println(err)
				if err := c.Reader.Close(); err != nil {
					log.Fatal("failed to close reader:", err)
				}
			}
		}
	}()

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
		}

		request := Request{}
		err = json.Unmarshal(p, &request)
		if err != nil {
			fmt.Println(err)
		}

		uuid, err := gocql.RandomUUID()
		if err != nil {
			continue
		}
		m := db.Message{
			MessageFrom: uuid,
			MessageTo:   uuid,
			Content:     string(request.Message),
		}

		jsonMessage, err := json.Marshal(m)
		if err != nil {
			fmt.Println(err)
			continue
		}

		err = b.Write([]byte(jsonMessage))
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(request.Message))
		// Add messages to queue topic

	}

}

func main() {
	// create websocket server

	r := mux.NewRouter()
	r.HandleFunc("/ws", WebSocketHandler)

	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}

	// Consumer the get messages from topic and save on DB

	// topic := "newtopic"
	// partition := 0
	// p, err := messagebroker.NewProducer("localhost:9092", topic, partition)
	// if err != nil {
	// 	panic(err)
	// }

	// err = p.Write([]byte("6666"))

	// if err != nil {
	// 	log.Fatal("failed to write messages:", err)
	// }

	// if err := p.Conn.Close(); err != nil {
	// 	log.Fatal("failed to close writer:", err)
	// }

	// consumer := messagebroker.NewConsumer("localhost:9092", topic, partition)
	// messages := make(chan []byte)
	// errors := make(chan error)
	// go consumer.ReadMessages(messages, errors)
	// for {
	// 	select {
	// 	case msg := <-messages:
	// 		fmt.Println("from loop" + string(msg))
	// 	case err := <-errors:
	// 		// Handle error
	// 		fmt.Println(err)
	// 		if err := consumer.Reader.Close(); err != nil {
	// 			log.Fatal("failed to close reader:", err)
	// 		}
	// 	}
	// }

	//os.Exit(0)

	// cluster := gocql.NewCluster("127.0.0.1")
	// cluster.Keyspace = "chatmessages"

	// session, err := cluster.CreateSession()
	// if err != nil {
	// 	panic(err)
	// }
	// defer session.Close()

	// c, err := db.NewCassandra("127.0.0.1", "chatmessages")
	// if err != nil {
	// 	panic(err)
	// }

	// defer c.Session.Close()

	// // uuid, err := gocql.ParseUUID("e63093e5-497c-407b-a391-676ba6d5db2f")
	// // if err != nil {
	// // 	panic(err)
	// // }

	// // m := db.Message{
	// // 	MessageFrom: uuid,
	// // 	MessageTo:   uuid,
	// // 	Content:     "Hello World100",
	// // }

	// // if err := c.CreateMessage(m); err != nil {
	// // 	panic(err)
	// // }

	// // m, err := c.GetMessages()
	// // if err != nil {
	// // 	panic(err)
	// // }
	// // fmt.Println(m)

	// // err = c.CreateUser(db.User{
	// // 	Username: "ernesto",
	// // 	Password: "password",
	// // })

	// // if err != nil {
	// // 	panic(err)
	// // }

	// err = c.LoginUser(db.User{
	// 	Username: "ernesto",
	// 	Password: "password",
	// })
	// if err != nil {
	// 	panic(err)
	// }

}
