package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"queue"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Could not load .env file: %v", err)
	}

	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		log.Fatal("RABBITMQ_URL environment variable is required")
	}

	queueName := os.Getenv("QUEUE_NAME")
	if queueName == "" {
		log.Fatal("QUEUE_NAME environment variable is required")
	}

	config := queue.Config{
		URL: rabbitmqURL,
	}

	q, err := queue.NewRabbitMQ(config)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer q.Close()

	queueConfig := queue.QueueConfig{
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
	}

	err = q.Create(queueName, queueConfig)
	if err != nil {
		log.Fatalf("Failed to create queue: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messages, err := q.Consume(ctx, queueName)
	if err != nil {
		log.Fatalf("Failed to start consuming messages: %v", err)
	}

	log.Printf("Consumer service started. Listening for messages on queue: %s", queueName)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case msg := <-messages:
			log.Printf("Received message: ID=%s, Type=%s, Content=%s, Timestamp=%s",
				msg.ID, msg.Type, msg.Content, msg.Timestamp.Format("2006-01-02 15:04:05"))
		case <-sigChan:
			log.Println("Received shutdown signal, stopping consumer...")
			cancel()
			return
		case <-ctx.Done():
			log.Println("Context cancelled, stopping consumer...")
			return
		}
	}
}
