package handlers

import (
	"encoding/json"
	"net/http"

	"userservice/db"
	"userservice/models"
)

func GetProfile(w http.ResponseWriter, r *http.Request) {
	// For now, just return a mock response
	// In a real application, you would extract the user ID from the JWT token
	// and fetch the user's profile from the database

	var user models.User
	if err := db.DB.First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
