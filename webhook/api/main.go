package main

import (
	"log"
	"webhook/handlers"

	"queue"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Initialize RabbitMQ queue
	queueName := "webhook_events"
	config := queue.Config{
		URL:       "amqp://admin:password@localhost:5672",
		QueueName: queueName,
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

	if err := rabbitMQ.Create(queueName, queueConfig); err != nil {
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
