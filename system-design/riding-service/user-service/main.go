package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"userservice/db"
	"userservice/handlers"
	"userservice/middleware"
	"userservice/utils"

	"github.com/ernesto/riding-service/shared/config"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize JWT
	utils.InitJWT(cfg)

	// Initialize database
	if err := db.Init(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	r := mux.NewRouter()

	// Auth routes
	r.HandleFunc("/api/auth/register", handlers.Register).Methods("POST")
	r.HandleFunc("/api/auth/login", handlers.Login).Methods("POST")

	// Protected routes
	protected := r.PathPrefix("/api").Subrouter()
	protected.Use(middleware.JWTAuth)
	protected.HandleFunc("/profile", handlers.GetProfile).Methods("GET")

	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
