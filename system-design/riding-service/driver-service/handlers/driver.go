package handlers

import (
	"driverservice/db"
	"driverservice/models"
	"driverservice/utils"
	"encoding/json"
	"net/http"
)

func RegisterDriver(w http.ResponseWriter, r *http.Request) {
	var driver models.Driver
	if err := json.NewDecoder(r.Body).Decode(&driver); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := db.DB.Create(&driver).Error; err != nil {
		http.Error(w, "Failed to register driver", http.StatusInternalServerError)
		return
	}

	token, err := utils.GenerateToken(driver.ID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func UpdateDriverStatus(w http.ResponseWriter, r *http.Request) {
	var status struct {
		IsAvailable bool `json:"is_available"`
	}

	if err := json.NewDecoder(r.Body).Decode(&status); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	driverID := r.Context().Value("userID").(uint)

	if err := db.DB.Model(&models.Driver{}).Where("id = ?", driverID).
		Update("is_available", status.IsAvailable).Error; err != nil {
		http.Error(w, "Failed to update status", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func UpdateLocation(w http.ResponseWriter, r *http.Request) {
	var location struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}

	if err := json.NewDecoder(r.Body).Decode(&location); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	driverID := r.Context().Value("userID").(uint)

	if err := db.DB.Model(&models.Driver{}).Where("id = ?", driverID).
		Updates(map[string]interface{}{
			"current_lat": location.Latitude,
			"current_lng": location.Longitude,
		}).Error; err != nil {
		http.Error(w, "Failed to update location", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
