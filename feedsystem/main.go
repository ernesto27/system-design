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

	nf := newsfeed.New(userID, cache, db)
	userPosts, err := nf.GetPostsCache()
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

	// tweets, err := db.GetTweetsFollowing(22)
	// if err != nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	w.Write([]byte("error getting tweets"))
	// 	return
	// }

	// d, err := json.Marshal(tweets)
	// if err != nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	w.Write([]byte("error getting tweets"))
	// 	return
	// }
	// w.Header().Set("Content-Type", "application/json")
	// w.Write(d)
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

	// nf := newsfeed.New(1, cache, mydb)
	// // ids, err := nf.GetFollowersIDs()
	// // if err != nil {
	// // 	panic(err)
	// // }
	// // fmt.Println(ids)
	// err = nf.SaveCache("88")
	// must(err)

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
