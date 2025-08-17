package queue

import (
	"context"
	"time"
)

type Message struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
}

type QueueConfig struct {
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
}

type DeliveryMessage struct {
	Message Message
	Ack     func() error
	Nack    func() error
}

type Queue interface {
	Create(config QueueConfig) error
	Publish(ctx context.Context, message Message) error
	Consume(ctx context.Context) (<-chan Message, error)
	ConsumeWithAck(ctx context.Context) (<-chan DeliveryMessage, error)
	Close() error
}

type Config struct {
	URL       string
	Username  string
	Password  string
	Host      string
	Port      string
	QueueName string
}
