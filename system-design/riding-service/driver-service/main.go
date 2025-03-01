package main

import (
	"log"
	"net/http"

	"driverservice/db"
	"driverservice/handlers"
	"driverservice/middleware"

	"github.com/ernesto/riding-service/shared/config"

	"github.com/gorilla/mux"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := db.Init(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	r := mux.NewRouter()

	// Auth routes
	r.HandleFunc("/api/drivers/register", handlers.RegisterDriver).Methods("POST")
	// r.HandleFunc("/api/drivers/login", handlers.LoginDriver).Methods("POST")

	// Protected routes
	protected := r.PathPrefix("/api/drivers").Subrouter()
	protected.Use(middleware.JWTAuth)
	// protected.HandleFunc("/profile", handlers.GetDriverProfile).Methods("GET")
	protected.HandleFunc("/status", handlers.UpdateDriverStatus).Methods("PUT")
	protected.HandleFunc("/location", handlers.UpdateLocation).Methods("PUT")
	// protected.HandleFunc("/vehicle", handlers.UpdateVehicleInfo).Methods("PUT")

	log.Println("Driver Service starting on port 8081...")
	log.Fatal(http.ListenAndServe(":8081", r))
}
