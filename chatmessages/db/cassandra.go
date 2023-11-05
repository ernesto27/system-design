package db

import (
	"math/rand"
	"time"

	"github.com/gocql/gocql"
	"golang.org/x/crypto/bcrypt"
)

type Cassandra struct {
	host     string
	keyspace string
	Session  *gocql.Session
}

type Message struct {
	ID          gocql.UUID
	MessageFrom gocql.UUID
	MessageTo   gocql.UUID
	Content     string
	CreatedAt   time.Time
	RecordID    gocql.UUID
}

type User struct {
	ID        gocql.UUID
	Username  string
	Password  string
	ApiToken  string
	CreatedAt time.Time
}

func NewCassandra(host string, keyspace string) (*Cassandra, error) {
	c := &Cassandra{host: host, keyspace: keyspace}
	cluster := gocql.NewCluster(c.host)
	cluster.Keyspace = c.keyspace

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	c.Session = session
	return c, nil
}

func (c *Cassandra) CreateMessage(m Message) error {
	err := c.Session.Query("INSERT INTO chatmessages.messages (id, message_from, message_to, content, record_id, created_at) VALUES (uuid(), ?, ?, ?, now(), toTimeStamp(now()))",
		m.MessageFrom, m.MessageTo, m.Content).Exec()
	return err
}

func (c *Cassandra) GetMessages() ([]Message, error) {
	messages := []Message{}

	scanner := c.Session.Query("SELECT id, message_from, message_to, content, record_id, created_at FROM chatmessages.messages").Iter().Scanner()

	var id gocql.UUID
	var messageFrom gocql.UUID
	var messageTo gocql.UUID
	var content string
	var recordID gocql.UUID
	var createdAt time.Time

	for scanner.Next() {
		err := scanner.Scan(&id, &messageFrom, &messageTo, &content, &recordID, &createdAt)
		if err != nil {
			return nil, err
		}

		messages = append(messages, Message{
			ID:          id,
			MessageFrom: messageFrom,
			MessageTo:   messageTo,
			Content:     content,
			RecordID:    recordID,
			CreatedAt:   createdAt,
		})
	}

	err := scanner.Err()
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (c *Cassandra) CreateUser(u User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = c.Session.Query("INSERT INTO chatmessages.users (id, username, password, api_token, created_at) VALUES (uuid(), ?, ?, ?, toTimeStamp(now()))",
		u.Username, hashedPassword, createRandomString()).Exec()
	return err
}

func (c *Cassandra) LoginUser(u User) error {
	var password string
	err := c.Session.Query("SELECT password FROM chatmessages.users WHERE username = ? LIMIT 1", u.Username).Scan(&password)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(u.Password))
	if err != nil {
		return err
	}

	return nil
}

func (c *Cassandra) UpdateConfig(id int, offset int) error {
	err := c.Session.Query("UPDATE chatmessages.config SET consumer_offset = ? WHERE id = ?", offset, id).Exec()
	return err
}

func (c *Cassandra) GetConfig(id int) (int, error) {
	var offset int
	err := c.Session.Query("SELECT consumer_offset FROM chatmessages.config WHERE id = ? LIMIT 1", id).Scan(&offset)
	if err != nil {
		return 0, err
	}

	return offset, nil
}

func createRandomString() string {
	length := 32
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}

	return string(result)
}
