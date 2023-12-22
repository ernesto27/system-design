package main

import (
	"feedsystem/db"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func handlePublish(w http.ResponseWriter, r *http.Request, db *db.Mysql) {
	err := db.CreateTweet("hello from golang", 1)
	must(err)
	w.Write([]byte("publish"))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {

	host := "localhost"
	user := "root"
	password := "1111"
	port := "3388"
	database := "feedsystem"

	mydb, err := db.NewMysql(host, user, password, port, database)
	must(err)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/feed", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	r.Post("/publish", func(w http.ResponseWriter, r *http.Request) {
		handlePublish(w, r, mydb)
	})
	http.ListenAndServe(":3000", r)
}
