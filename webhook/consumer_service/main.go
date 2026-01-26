package main

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"os"
	"os/signal"
	"queue"
	"syscall"
	"time"

	"database"

	"github.com/google/uuid"
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

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Initialize database connection
	db, err := database.Connect(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize RabbitMQ connection

	config := queue.Config{
		URL:       rabbitmqURL,
		QueueName: queueName,
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

	err = q.Create(queueConfig)
	if err != nil {
		log.Fatalf("Failed to create queue: %v", err)
	}

	// Create separate DLQ connection
	dlqConfig := queue.Config{
		URL:       rabbitmqURL,
		QueueName: queueName + "_dlq",
	}
	dlqQueue, err := queue.NewRabbitMQ(dlqConfig)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ for DLQ: %v", err)
	}
	defer dlqQueue.Close()

	err = dlqQueue.Create(queueConfig)
	if err != nil {
		log.Fatalf("Failed to create dead letter queue: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messages, err := q.ConsumeWithAck(ctx)
	if err != nil {
		log.Fatalf("Failed to start consuming messages: %v", err)
	}

	log.Printf("Consumer service started. Listening for messages on queue: %s", queueName)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case delivery := <-messages:
			// Process each message in its own goroutine to avoid blocking
			go func(delivery queue.DeliveryMessage) {
				log.Printf("Received message: ID=%s, Type=%s, Content=%s, Timestamp=%s",
					delivery.Message.ID, delivery.Message.Type, delivery.Message.Content,
					delivery.Message.Timestamp.Format("2006-01-02 15:04:05"))

				// Convert queue.Message to database.WebhookEvent
				var data map[string]any
				if err := json.Unmarshal([]byte(delivery.Message.Content), &data); err != nil {
					// If content is not JSON, store as is
					data = map[string]any{"content": delivery.Message.Content}
				}

				// Convert data to JSON for database storage
				_, err := json.Marshal(data)
				if err != nil {
					log.Printf("Failed to marshal data to JSON: %v", err)
					// Send malformed message to DLQ immediately
					sendToDLQ(dlqQueue, ctx, delivery.Message, "marshal_error")
					delivery.Ack()
					return
				}

				event := database.WebhookEvent{
					ID:        uuid.New(),
					EventID:   &delivery.Message.ID,
					Source:    "queue",
					Type:      delivery.Message.Type,
					Data:      data,
					Timestamp: delivery.Message.Timestamp,
					CreatedAt: time.Now(),
				}

				// Retry logic with exponential backoff
				maxRetries := 3
				baseDelay := 1 * time.Second
				success := false

				for attempt := 1; attempt <= maxRetries; attempt++ {
					if err := database.SaveWebhookEvent(db, event); err != nil {
						log.Printf("Attempt %d/%d failed to save message %s to database: %v",
							attempt, maxRetries, delivery.Message.ID, err)

						if attempt < maxRetries {
							// Exponential backoff: 1s, 2s, 4s
							delay := time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt-1)))
							log.Printf("Message %s: retrying in %v...", delivery.Message.ID, delay)
							time.Sleep(delay) // Only blocks this goroutine, not others
						}
					} else {
						log.Printf("Message %s saved to database successfully on attempt %d",
							delivery.Message.ID, attempt)
						success = true
						break
					}
				}

				if success {
					// Ack the message after successful DB save
					if ackErr := delivery.Ack(); ackErr != nil {
						log.Printf("Failed to ack message %s: %v", delivery.Message.ID, ackErr)
					}
				} else {
					// Max retries exceeded, send to dead letter queue
					log.Printf("Max retries exceeded for message %s, sending to DLQ", delivery.Message.ID)
					sendToDLQ(dlqQueue, ctx, delivery.Message, "max_retries_exceeded")
					delivery.Ack() // Remove from main queue
				}
			}(delivery)

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

func sendToDLQ(dlqQueue *queue.RabbitMQ, ctx context.Context, originalMsg queue.Message, reason string) {
	// Add metadata about failure reason
	metadata := map[string]any{
		"original_message": originalMsg,
		"failure_reason":   reason,
		"failed_at":        time.Now().Format(time.RFC3339),
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		log.Printf("Failed to marshal metadata for DLQ: %v", err)
		return
	}

	dlqMessage := queue.Message{
		ID:        "dlq_" + originalMsg.ID,
		Content:   string(metadataJSON),
		Type:      "failed_" + originalMsg.Type,
		Timestamp: time.Now(),
	}

	if err := dlqQueue.Publish(ctx, dlqMessage); err != nil {
		log.Printf("Failed to send message to DLQ: %v", err)
	} else {
		log.Printf("Message %s sent to DLQ", originalMsg.ID)
	}
}
