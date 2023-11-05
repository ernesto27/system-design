package messagebroker

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Kafka struct {
	Conn   *kafka.Conn
	Reader *kafka.Reader
}

func NewProducer(host string, topic string, partition int) (*Kafka, error) {
	conn, err := kafka.DialLeader(context.Background(), "tcp", host, topic, partition)
	if err != nil {
		return nil, err
	}

	//conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	return &Kafka{Conn: conn}, nil
}

func (k *Kafka) Write(msg []byte) error {
	_, err := k.Conn.WriteMessages(
		kafka.Message{Value: msg},
	)
	if err != nil {
		return err
	}

	return nil
}

func NewConsumer(host string, topic string, partition int, offset int64) *Kafka {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{host},
		Topic:     topic,
		Partition: partition,
		MaxBytes:  10e6, // 10MB
	})
	r.SetOffset(offset)

	return &Kafka{Reader: r}
}

func (k *Kafka) ReadMessages(messages chan<- []byte, errors chan<- error) {
	for {
		m, err := k.Reader.ReadMessage(context.Background())
		if err != nil {
			errors <- err
			break
		}
		messages <- m.Value
	}
}
