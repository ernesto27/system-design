package main

import (
	"log"
	"os"
	"webhook/database"
	"webhook/handlers"

	"queue"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load .env file if running outside Docker
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	// Initialize database connection
	dbURL := os.Getenv("DATABASE_URL")
	db, err := database.Connect(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize RabbitMQ queue
	config := queue.Config{
		URL:       os.Getenv("RABBITMQ_URL"),
		QueueName: os.Getenv("QUEUE_NAME"),
	}

	rabbitMQ, err := queue.NewRabbitMQ(config)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitMQ.Close()

	// Create webhook queue
	queueConfig := queue.QueueConfig{
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
	}

	if err := rabbitMQ.Create(os.Getenv("QUEUE_NAME"), queueConfig); err != nil {
		log.Fatalf("Failed to create queue: %v", err)
	}

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Routes
	e.POST("/webhook", func(c echo.Context) error {
		return handlers.HandleWebhook(c, rabbitMQ)
	})

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}
