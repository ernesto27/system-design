package handlers

import (
	"chatmessages/db"
	"chatmessages/types"
	"encoding/json"
	"fmt"
	"net/http"
)

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
