package main

import (
	"chatmessages/db"
	"chatmessages/handlers"
	"chatmessages/messagebroker"
	"chatmessages/types"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var clients = make(map[*websocket.Conn]types.Client)

func hearbeatStatusTimeout(clients map[*websocket.Conn]types.Client, conn *websocket.Conn) {
	for range time.Tick(time.Second) {
		c := clients[conn]
		c.UpdateSeconds(c.Seconds + 1)

		if c.Seconds > 30 {
			c.UpdateSeconds(0)
			fmt.Println("user is offline")
			// TODO prevent update if status does not change
			err := cassandra.UpdateUserStatus(c.User.ID, types.StatusOffline)
			if err != nil {
				fmt.Println(err)
				continue
			}

			// Send message to queue
			sendUpdateStatusToTopics(c.User, types.StatusOffline)

		}

		clients[conn] = c
	}
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("WebSocketHandler")

	// websocker upgrade
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	clients[conn] = types.Client{Conn: conn, Seconds: 0}

	user, err := cassandra.GetUserByApiKey(r.Header.Get("Api-Token"))
	if err != nil {
		fmt.Println(err)
		clients[conn].Conn.WriteJSON(types.
			Response{Message: "Invalid Api-Token"})
		clients[conn].Conn.Close()
		return
	}

	t := clients[conn]
	t.User = user
	clients[conn] = t

	go hearbeatStatusTimeout(clients, conn)

	fmt.Println("USER UUID ", user.ID)
	fmt.Println("USER CHANNELS", user.Channels)

	for _, channel := range user.Channels {
		// setup consumers
		fmt.Println("consumer listening to " + channel.String())
		c := messagebroker.NewConsumer("localhost:9092", channel.String()+"_C", 0, 0)
		// listen to messages
		go func(c *messagebroker.Kafka, channel gocql.UUID) {
			messages := make(chan []byte)
			errors := make(chan error)
			go c.ReadMessages(messages, errors)
			for {
				select {
				case msg := <-messages:
					m := types.Message{}
					err = json.Unmarshal(msg, &m)
					if err != nil {
						fmt.Println(err)
						continue
					}

					if m.Content != "" {
						// Check message date after last message read by user
						layout := "2006-01-02T15:04:05.99Z"
						lastCreated, err := time.Parse(layout, r.Header.Get("Last-Created-At"))
						if err != nil {
							fmt.Println(err)
							continue
						}

						m := types.Message{}
						err = json.Unmarshal(msg, &m)
						if err != nil {
							fmt.Println(err)
							continue
						}

						if m.CreatedAt.UTC().Before(lastCreated) {
							continue
						}

						fmt.Println("send message to client", string(msg))
						clients[conn].Conn.WriteJSON(string(msg))
					} else {
						// check status message
						s := types.UserStatus{}
						err := json.Unmarshal(msg, &s)
						if err != nil {
							fmt.Println(err)
							continue
						}

						fmt.Println("send status to client", string(msg))
						clients[conn].Conn.WriteJSON(string(msg))
					}

				case err := <-errors:
					// Handle error
					fmt.Println(err)
					if err := c.Reader.Close(); err != nil {
						fmt.Println("failed to close reader:", err)
					}
				}
			}
		}(c, channel)
	}

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			continue
		}

		request := types.Request{}
		err = json.Unmarshal(p, &request)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if request.Type == types.TypeMessage {
			// if !user.ValidateContact(request.MessageTo) {
			// 	fmt.Println("invalid contact")
			// 	continue
			// }

			// if !user.ValidateChannel(request.ChannelID) {
			// 	fmt.Println("invalid channel")
			// 	continue
			// }

			mTo, err := gocql.ParseUUID(request.MessageTo)
			if err != nil {
				fmt.Println(err)
				continue
			}

			channelID, err := gocql.ParseUUID(request.ChannelID)
			if err != nil {
				fmt.Println(err)
				continue
			}

			m := types.Message{
				MessageFrom: user.ID,
				MessageTo:   mTo,
				Content:     string(request.Content),
				ChannelID:   channelID,
			}

			jsonMessage, err := json.Marshal(m)
			if err != nil {
				fmt.Println(err)
				continue
			}

			b, err := messagebroker.NewProducer("localhost:9092", request.ChannelID+"_P", 0)
			if err != nil {
				fmt.Println(err)
				continue
			}

			err = b.Write([]byte(jsonMessage))
			if err != nil {
				fmt.Println(err)
			}
			b.Conn.Close()
		} else if request.Type == types.TypeUpdateStatus {
			fmt.Println(clients[conn].Seconds)
			err := cassandra.UpdateUserStatus(user.ID, types.StatusOnline)

			c := clients[conn]
			t.UpdateSeconds(0)
			clients[conn] = c

			if err != nil {
				fmt.Println(err)
				continue
			}

			sendUpdateStatusToTopics(user, types.StatusOnline)
		}

	}
}

func sendUpdateStatusToTopics(user types.User, status types.Status) {
	// Send message to topics user channels
	for _, channel := range user.Channels {
		fmt.Println("send status message to topic " + channel.String())
		b, err := messagebroker.NewProducer("localhost:9092", channel.String()+"_C", 0)
		if err != nil {
			fmt.Println(err)
			continue
		}

		s := types.UserStatus{
			UserID: user.ID,
			Status: status,
		}

		jsonMessage, err := json.Marshal(s)
		if err != nil {
			fmt.Println(err)
			continue
		}

		err = b.Write([]byte(jsonMessage))
		if err != nil {
			fmt.Println(err)
		}
		b.Conn.Close()
	}
}

var cassandra *db.Cassandra

func main() {

	var err error
	cassandra, err = db.NewCassandra("127.0.0.1", "chatmessages")
	if err != nil {
		panic(err)
	}
	defer cassandra.Session.Close()

	// uuid, err := gocql.ParseUUID("c13b2d17-e60e-4f60-9a39-d922eef257cd")
	// if err != nil {
	// 	panic(err)
	// }

	// Input date and time string
	// dateTimeStr := "2023-11-11T20:17:10.307Z"
	// layout := "2006-01-02T15:04:05.99Z"
	// parsedTime, err := time.Parse(layout, dateTimeStr)
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println("Parsed time:", parsedTime)
	// fmt.Println(time.Now().Add(1000 * time.Hour))
	// if parsedTime.Before(time.Now().Add(-1000 * time.Hour)) {
	// 	fmt.Println("Parsed time is before current time")
	// } else {
	// 	fmt.Println("Parsed time is after current time")
	// }
	// os.Exit(1)

	// m, err := cassandra.GetMessagesOneToOne(uuid, parsedTime)
	// //m, err := cassandra.GetMessagesOneToOne(uuid, time.Now())
	// if err != nil {
	// 	panic(err)
	// }

	//os.Exit(1)

	// create websocket server
	r := mux.NewRouter()
	r.HandleFunc("/ws", WebSocketHandler)

	r.HandleFunc("/messages-one-to-one", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleMessagesOneToOne(w, r, cassandra)
	})

	// r.HandleFunc("users/update-status", func(w http.ResponseWriter, r *http.Request) {
	// 	handlers.HandleUpdateStatus(w, r, cassandra)
	// })

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
