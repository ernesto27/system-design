package main

import (
	"context"
	"log"
	"queue"
	"time"
)

func main() {
	config := queue.Config{
		URL: "amqp://admin:password@localhost:5672",
	}

	q, err := queue.NewRabbitMQ(config)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer q.Close()

	queueName := "webhook_messages"
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

	msg := queue.Message{
		ID:        "msg-00e1",
		Content:   "Test webhook message",
		Timestamp: time.Now(),
		Type:      "webhook_event",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = q.Publish(ctx, msg)
	if err != nil {
		log.Fatalf("Failed to publish message: %v", err)
	}

	log.Printf("Successfully published message: %s", msg.ID)
}
