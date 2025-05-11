package main

import (
	"context"
	"encoding/json"
	"errors"
	"firmaelectronica/internal/controllers"
	"firmaelectronica/pkg/auth"
	"firmaelectronica/pkg/db"
	"firmaelectronica/pkg/storage"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

type config struct {
	// Database configuration
	DBHost     string `env:"DB_HOST" envDefault:"localhost"`
	DBPort     int    `env:"DB_PORT" envDefault:"5432"`
	DBName     string `env:"DB_NAME" envDefault:"firma_electronica"`
	DBUser     string `env:"DB_USER" envDefault:"admin"`
	DBPassword string `env:"DB_PASSWORD" envDefault:"secure_password"`
	DBLogLevel string `env:"DB_LOG_LEVEL" envDefault:"info"`

	// JWT configuration
	JWTSecret     string        `env:"JWT_SECRET" envDefault:"your-default-secret-key-change-this-in-production"`
	JWTExpiration time.Duration `env:"JWT_EXPIRATION" envDefault:"24h"`

	// S3 configuration
	S3BucketName       string        `env:"S3_BUCKET_NAME" envDefault:""`
	S3Region           string        `env:"AWS_REGION" envDefault:""`
	S3Timeout          time.Duration `env:"S3_TIMEOUT" envDefault:"30s"`
	AWSAccessKeyID     string        `env:"AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey string        `env:"AWS_SECRET_ACCESS_KEY"`

	// HTTP server configuration
	HTTPPort        int           `env:"HTTP_PORT" envDefault:"8080"`
	ReadTimeout     time.Duration `env:"READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout    time.Duration `env:"WRITE_TIMEOUT" envDefault:"10s"`
	IdleTimeout     time.Duration `env:"IDLE_TIMEOUT" envDefault:"120s"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"10s"`
}

func main() {
	// Parse command-line arguments
	command := "serve"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or could not be loaded: %v", err)
	}

	// Load configuration from environment variables
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to parse environment variables: %v", err)
	}

	// Create database config from environment variables
	dbConfig := db.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		Name:     cfg.DBName,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		LogLevel: cfg.DBLogLevel,
	}

	// Process command
	switch command {
	case "migrate":
		runMigrations(dbConfig)
	case "serve":
		runServer(cfg, dbConfig)
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}

func runMigrations(dbConfig db.Config) {
	// Create database connection
	database, err := db.New(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Get the underlying SQL DB to close it properly
	sqlDB, err := database.DB.DB()
	if err != nil {
		log.Fatalf("Failed to get SQL DB: %v", err)
	}
	defer sqlDB.Close()

	// Run migrations
	if err := database.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Seed test user
	if err := database.SeedTestUser(); err != nil {
		log.Fatalf("Failed to seed test user: %v", err)
	}

	log.Println("Migrations completed successfully")
}

func runServer(cfg config, dbConfig db.Config) {
	// Create database connection
	database, err := db.New(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Get the underlying SQL DB to close it properly
	sqlDB, err := database.DB.DB()
	if err != nil {
		log.Fatalf("Failed to get SQL DB: %v", err)
	}
	defer sqlDB.Close()

	// Create JWT service
	jwtConfig := auth.Config{
		Secret:     cfg.JWTSecret,
		Expiration: cfg.JWTExpiration,
	}
	jwtService := auth.NewService(jwtConfig)

	// Create S3 storage service
	s3Config := storage.S3Config{
		BucketName:      cfg.S3BucketName,
		Region:          cfg.S3Region,
		Timeout:         cfg.S3Timeout,
		AccessKeyID:     cfg.AWSAccessKeyID,
		SecretAccessKey: cfg.AWSSecretAccessKey,
	}
	s3Storage, err := storage.New(s3Config)
	if err != nil {
		log.Fatalf("Failed to create S3 storage service: %v", err)
	}

	// Initialize controllers with dependencies
	controller := controllers.NewController(database, jwtService)

	// Initialize document handler
	documentHandler := controllers.NewDocumentHandler(database, s3Storage)

	// Create a new server mux
	mux := http.NewServeMux()

	// Register routes using latest pattern syntax
	mux.HandleFunc("GET /hello", controller.HelloHandler)
	mux.HandleFunc("POST /api/login", controller.LoginHandler)

	// Protected API routes
	protected := http.NewServeMux()
	protected.HandleFunc("GET /api/protected", func(w http.ResponseWriter, r *http.Request) {
		// Extract the claims from the request context
		claims, ok := r.Context().Value(controllers.UserClaimsKey).(*auth.JWTClaims)
		if !ok {
			http.Error(w, "User not authenticated", http.StatusUnauthorized)
			return
		}

		// Return user ID and message in the response
		response := map[string]string{
			"message": "This is a protected endpoint",
			"user_id": claims.UserID,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Document routes
	protected.HandleFunc("POST /api/documents", documentHandler.Create)

	// Use auth middleware for protected routes
	mux.Handle("GET /api/protected", controller.AuthMiddleware(protected))
	mux.Handle("POST /api/documents", controller.AuthMiddleware(protected))

	// Configure the HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	// Start server in a goroutine so that it doesn't block shutdown handling
	go func() {
		log.Printf("HTTP server running on port %d", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	// Block until we receive a signal
	<-quit
	log.Println("Server shutting down...")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
