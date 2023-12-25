package main

import (
	"encoding/json"
	"feedsystem/cache"
	"feedsystem/db"
	"feedsystem/newsfeed"
	"feedsystem/postservice"
	"feedsystem/types"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func handlePublish(w http.ResponseWriter, r *http.Request, db *db.Mysql, cache *cache.Redis) {
	var requestPayload types.Request
	err := json.NewDecoder(r.Body).Decode(&requestPayload)

	jsonResponse := types.JSONResponse{}
	if err != nil {
		jsonResponse.Status = "error"
		jsonResponse.Message = "Invalid body payload"
		responseJSON(w, jsonResponse)
		return
	}

	userID := 1
	requestPayload.UserID = userID
	postID, success := postservice.Create(db, cache, requestPayload)
	if !success {
		jsonResponse.Status = "error"
		jsonResponse.Message = "error creating tweet"
		responseJSON(w, jsonResponse)
		return
	}

	go func() {
		nf := newsfeed.New(userID, cache, db)
		err = nf.SaveCache(fmt.Sprint(postID))
		if err != nil {
			fmt.Println(err)
		}
	}()

	jsonResponse.Status = "success"
	jsonResponse.Message = "tweet created"
	responseJSON(w, jsonResponse)
}

func handleFeed(w http.ResponseWriter, r *http.Request, db *db.Mysql, cache *cache.Redis) {
	userID := 4
	jsonResponse := types.JSONResponse{}

	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}

	p, err := strconv.Atoi(page)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonResponse.Message = "Invalid page"
		responseJSON(w, jsonResponse)
		return
	}

	nf := newsfeed.New(userID, cache, db)
	userPosts, err := nf.GetPostsCache(p)
	if err != nil {
		jsonResponse.Status = "error"
		jsonResponse.Message = "error getting posts"
		responseJSON(w, jsonResponse)
		return
	}

	jsonResponse.Status = "success"
	jsonResponse.Message = "success getting posts"
	jsonResponse.Data = userPosts
	responseJSON(w, jsonResponse)

}

func responseJSON(w http.ResponseWriter, data interface{}) {
	d, err := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server errror"))
		return
	}

	w.Write(d)
}

func main() {

	host := "localhost"
	user := "root"
	password := "1111"
	port := "3388"
	database := "feedsystem"

	mydb, err := db.NewMysql(host, user, password, port, database)
	must(err)

	cache := cache.NewRedis("localhost", "6381")

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/feed", func(w http.ResponseWriter, r *http.Request) {
		handleFeed(w, r, mydb, cache)
	})
	r.Post("/publish", func(w http.ResponseWriter, r *http.Request) {
		handlePublish(w, r, mydb, cache)
	})
	http.ListenAndServe(":3000", r)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
