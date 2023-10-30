package main

import (
	"chatmessages/db"
)

func main() {
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
