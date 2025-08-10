package queue

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn      *amqp.Connection
	ch        *amqp.Channel
	queueName string
}

func NewRabbitMQ(config Config) (*RabbitMQ, error) {
	conn, err := amqp.Dial(config.URL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &RabbitMQ{
		conn:      conn,
		ch:        ch,
		queueName: config.QueueName,
	}, nil
}

func (r *RabbitMQ) Create(queueName string, config QueueConfig) error {
	_, err := r.ch.QueueDeclare(
		queueName,
		config.Durable,
		config.AutoDelete,
		config.Exclusive,
		config.NoWait,
		nil,
	)
	return err
}

func (r *RabbitMQ) Publish(ctx context.Context, message Message) error {
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return r.ch.PublishWithContext(ctx,
		"",             // exchange
		r.queueName,    // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}

func (r *RabbitMQ) Consume(ctx context.Context, queueName string) (<-chan Message, error) {
	msgs, err := r.ch.Consume(
		queueName,
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return nil, err
	}

	messageCh := make(chan Message)
	go func() {
		defer close(messageCh)
		for {
			select {
			case d, ok := <-msgs:
				if !ok {
					return
				}
				var msg Message
				if err := json.Unmarshal(d.Body, &msg); err == nil {
					select {
					case messageCh <- msg:
					case <-ctx.Done():
						return
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return messageCh, nil
}

func (r *RabbitMQ) Close() error {
	if r.ch != nil {
		r.ch.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}