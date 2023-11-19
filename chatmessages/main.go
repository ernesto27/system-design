package main

import (
	"chatmessages/db"
	"chatmessages/handlers"
	"net/http"

	"github.com/gorilla/mux"
)

var cassandra *db.Cassandra

func main() {

	var err error
	cassandra, err = db.NewCassandra("127.0.0.1", "chatmessages")
	if err != nil {
		panic(err)
	}
	defer cassandra.Session.Close()

	// DELETE

	// id, err := gocql.ParseUUID("4cc3fd00-8195-11ee-9685-a8b13baddc45")
	// if err != nil {
	// 	panic(err)
	// }

	// channelID, err := gocql.ParseUUID("c13b2d17-e60e-4f60-9a39-d922eef257cd")
	// if err != nil {
	// 	panic(err)
	// }

	// createdAt, err := db.ParseTime("2023-11-12T19:54:33.308Z")
	// if err != nil {
	// 	panic(err)
	// }

	// m := types.Message{
	// 	ID:        id,
	// 	ChannelID: channelID,
	// 	CreatedAt: createdAt,
	// }

	// m, err := types.NewMessage(types.Request{
	// 	ID:        "0db9839e-8195-11ee-9684-a8b13baddc45",
	// 	ChannelID: "c13b2d17-e60e-4f60-9a39-d922eef257cd",
	// 	CreatedAt: "2023-11-12T19:52:47.543Z",
	// })

	// if err != nil {
	// 	panic(err)
	// }

	// err = cassandra.DeleteMessage(m)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("deleted")

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
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handlers.WebSocketHandler(w, r, cassandra)
	})

	r.HandleFunc("/messages-one-to-one", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleMessagesOneToOne(w, r, cassandra)
	})

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
