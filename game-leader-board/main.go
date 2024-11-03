package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/joho/godotenv"
)

type LeaderboardEntry struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Score string `json:"score"`
}

var ddbClient *dynamodb.Client

func handleAddUpdateScore(w http.ResponseWriter, r *http.Request) {
	var entry ScoreRequest
	err := json.NewDecoder(r.Body).Decode(&entry)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	if entry.UserID == 0 || entry.GameID == "" || entry.CurrentScore == "" {
		fmt.Println("Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = myDB.PutItem(entry)

	if err != nil {
		fmt.Println("PUT ITEM ERROR", err)
		http.Error(w, "Failed to add/update score", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Score added/updated successfully"))
}

func handleGetLeaderboard(w http.ResponseWriter, r *http.Request) {
	leaderboardName := r.URL.Query().Get("leaderboard_name")
	if leaderboardName == "" {
		http.Error(w, "leaderboard_name is required", http.StatusBadRequest)
		return
	}

	entries, err := myDB.GetLeaderboard(leaderboardName)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to get leaderboard", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

var myDB *DynamoDB

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	myDB = NewDynamoDB(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), os.Getenv("AWS_REGION"))

	router := http.NewServeMux()

	router.HandleFunc("POST /score", handleAddUpdateScore)
	router.HandleFunc("GET /leaderboard", handleGetLeaderboard)

	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	log.Printf("Starting server on port %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
