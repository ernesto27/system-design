package messagequeue

import (
	"context"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Rabbit struct {
	Conn *amqp.Connection
}

func New(host string) (*Rabbit, error) {
	// conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	conn, err := amqp.Dial(host)
	if err != nil {
		return nil, err
	}

	return &Rabbit{
		Conn: conn,
	}, nil
}

func (r *Rabbit) Producer(message string) error {
	// defer conn.Close()

	ch, err := r.Conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"links", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body := message
	err = ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	if err != nil {
		return err
	}

	log.Printf(" [x] Sent %s\n", body)
	return nil

}

func (r *Rabbit) Consumer(messages chan<- []byte, errors chan<- error) {
	ch, err := r.Conn.Channel()
	if err != nil {
		errors <- err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"links", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		errors <- err
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		errors <- err
	}

	var forever chan struct{}
	go func() {
		for d := range msgs {
			messages <- d.Body
		}
	}()

	<-forever
}
