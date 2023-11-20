package handlers

import (
	"chatmessages/db"
	"chatmessages/messagebroker"
	"chatmessages/types"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gocql/gocql"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var clients = make(map[*websocket.Conn]types.Client)
var resetTimerChan = make(chan *websocket.Conn)

func hearbeatStatusTimeout(clients map[*websocket.Conn]types.Client, conn *websocket.Conn, db *db.Cassandra) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c := clients[conn]
			c.UpdateSeconds(c.Seconds + 1)

			if c.Seconds > 30 {
				c.UpdateSeconds(0)
				fmt.Println("user is offline")
				// TODO prevent update if status does not change
				err := db.UpdateUserStatus(c.User.ID, types.StatusOffline)
				if err != nil {
					fmt.Println(err)
					continue
				}

				// Send message to queue
				sendUpdateStatusToTopics(c.User, types.StatusOffline)
			}

			clients[conn] = c

		case <-resetTimerChan:
			// Reset the timer
			c := clients[conn]
			c.UpdateSeconds(0)
			clients[conn] = c
		}
	}
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request, db *db.Cassandra) {
	fmt.Println("WebSocketHandler")

	// websocker upgrade
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	clients[conn] = types.Client{Conn: conn, Seconds: 0}

	user, err := db.GetUserByApiKey(r.Header.Get("Api-Token"))
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

	go hearbeatStatusTimeout(clients, conn, db)

	fmt.Println("user uuid: ", user.ID)
	fmt.Println("user channels: ", user.Channels)

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

					switch m.Type {
					case types.CreateMessage:
						// Check message date after last message read by user
						layout := "2006-01-02T15:04:05.99Z"
						lastCreated, err := time.Parse(layout, r.Header.Get("Last-Created-At"))
						if err != nil {
							fmt.Println(err)
							continue
						}

						if err != nil {
							fmt.Println(err)
							continue
						}

						if m.CreatedAt.UTC().Before(lastCreated) {
							continue
						}

						fmt.Println("send message to client", string(msg))
						clients[conn].Conn.WriteJSON(string(msg))
					case types.DeleteMessage:
						fmt.Println("send message to delete to clients", string(msg))
						clients[conn].Conn.WriteJSON(string(msg))
					case types.UpdateMessage:
						fmt.Println("send message to update to clients", string(msg))
						clients[conn].Conn.WriteJSON(string(msg))

					case "":
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

		switch request.Type {
		case types.CreateMessage:
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
				Type:        types.CreateMessage,
			}

			id, createdAt, err := db.CreateMessage(m)
			if err != nil {
				fmt.Println(err)
				// send message to client
				continue
			}

			m.ID = id
			m.CreatedAt = createdAt

			err = messagebroker.SaveMessage("localhost:9092", request.ChannelID+"_C", 0, request.ChannelID, m)
			if err != nil {
				fmt.Println(err)
				continue
			}

		case types.UpdateStatus:
			fmt.Println(clients[conn].Seconds)
			err := db.UpdateUserStatus(user.ID, types.StatusOnline)

			resetTimerChan <- conn

			if err != nil {
				fmt.Println(err)
				continue
			}

			sendUpdateStatusToTopics(user, types.StatusOnline)

		case types.DeleteMessage:
			fmt.Println("delete message")
			request.MessageFrom = user.ID.String()
			m, err := types.NewMessage(request)

			if err != nil {
				fmt.Println(err)
				continue
			}

			m.Type = types.DeleteMessage
			err = db.DeleteMessage(m)
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Println("delete message, send to topic")
			err = messagebroker.SaveMessage("localhost:9092", request.ChannelID+"_C", 0, request.ChannelID, m)
			if err != nil {
				fmt.Println(err)
				continue
			}

		case types.UpdateMessage:
			fmt.Println("update message")
			request.MessageFrom = user.ID.String()
			m, err := types.NewMessage(request)

			if err != nil {
				fmt.Println(err)
				continue
			}

			m.Type = types.UpdateMessage
			err = db.UpdateMessage(m)
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Println("update message, send to topic")
			err = messagebroker.SaveMessage("localhost:9092", request.ChannelID+"_C", 0, request.ChannelID, m)
			if err != nil {
				fmt.Println(err)
				continue
			}

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

func HandleMessagesOneToOne(w http.ResponseWriter, r *http.Request, db *db.Cassandra) {
	// validate token
	_, err := db.GetUserByApiKey(r.Header.Get("Api-Token"))
	if err != nil {
		fmt.Println("Error:", err)
		jsonData, _ := json.Marshal(types.Response{Status: "error", Message: "Invalid Api-Token"})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(jsonData)
		return
	}

	// Validate channelID associated with user

	// Validate query params
	queryParams := r.URL.Query()
	channelID := queryParams.Get("channelID")
	createdAt := queryParams.Get("createdAt")

	// Get messages from DB
	messages, err := db.GetMessagesOneToOne(channelID, createdAt)
	if err != nil {
		fmt.Println("Error:", err)
		jsonData, _ := json.Marshal(types.Response{Status: "error", Message: "Error getting messages"})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonData)
		return
	}

	// Send messages to client
	jsonData, _ := json.Marshal(messages)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
