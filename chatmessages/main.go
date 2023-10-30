package main

import (
	"fmt"

	"github.com/gocql/gocql"
)

func main() {
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "chatmessages"

	session, err := cluster.CreateSession()
	if err != nil {
		panic(err)
	}
	defer session.Close()

	var result string
	if err := session.Query("select username from chatmessages.users").Scan(&result); err != nil {
		panic(err)
	}
	fmt.Println(result)
}
