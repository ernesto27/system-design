package messagebroker

import (
	"chatmessages/types"
	"context"
	"encoding/json"

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

func SaveMessage(host string, topic string, partition int, channelID string, m types.Message) error {
	b, err := NewProducer("localhost:9092", channelID+"_C", 0)
	// TODO: on kafka error, check how you can retry
	if err != nil {
		return err
	}

	defer b.Conn.Close()

	jsonMessage, err := json.Marshal(m)
	if err != nil {
		return err
	}

	err = b.Write([]byte(jsonMessage))
	if err != nil {
		return err
	}

	return nil
}
