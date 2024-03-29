package db

import (
	"chatmessages/types"
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

func (c *Cassandra) CreateMessage(m types.Message) (gocql.UUID, time.Time, error) {
	uuid := gocql.TimeUUID()
	now := time.Now().UTC()
	err := c.Session.Query("INSERT INTO chatmessages.messages (id, message_from, message_to, channel_id, content, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		uuid,
		m.MessageFrom,
		m.MessageTo,
		m.ChannelID,
		m.Content,
		now).Exec()

	if err != nil {
		return gocql.UUID{}, now, err
	}

	return uuid, now, nil
}

func (c *Cassandra) GetMessages() ([]types.Message, error) {
	messages := []types.Message{}

	scanner := c.Session.Query("SELECT id, message_from, message_to, content, record_id, created_at FROM chatmessages.messages").Iter().Scanner()

	var id gocql.UUID
	var messageFrom gocql.UUID
	var messageTo gocql.UUID
	var content string
	var channelID gocql.UUID
	var createdAt time.Time

	for scanner.Next() {
		err := scanner.Scan(&id, &messageFrom, &messageTo, &content, &channelID, &createdAt)
		if err != nil {
			return nil, err
		}

		messages = append(messages, types.Message{
			ID:          id,
			MessageFrom: messageFrom,
			MessageTo:   messageTo,
			Content:     content,
			CreatedAt:   createdAt,
		})
	}

	err := scanner.Err()
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (c *Cassandra) DeleteMessage(m types.Message) error {
	err := c.Session.Query("DELETE FROM chatmessages.messages WHERE id = ? AND channel_id = ? AND created_at = ? AND message_from = ?", m.ID, m.ChannelID, m.CreatedAt, m.MessageFrom).Exec()
	return err

}

func (c *Cassandra) UpdateMessage(m types.Message) error {
	err := c.Session.Query(`UPDATE chatmessages.messages 
							  SET content = ? 
							  WHERE id = ? 
							  AND channel_id = ? 
							  AND created_at = ? 
							  AND message_from = ?`,
		m.Content, m.ID, m.ChannelID, m.CreatedAt, m.MessageFrom).Exec()
	return err
}

func (c *Cassandra) CreateUser(u types.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = c.Session.Query("INSERT INTO chatmessages.users (id, username, password, api_token, contacts, created_at) VALUES (uuid(), ?, ?, ?, ?, toTimeStamp(now()))",
		u.Username, hashedPassword, createRandomString(), u.Contacts).Exec()

	return err
}

func (c *Cassandra) LoginUser(u types.User) error {
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

func (c *Cassandra) GetUserByApiKey(apiKey string) (types.User, error) {
	var user types.User
	err := c.Session.Query("SELECT id, username, api_token, contacts, channels FROM chatmessages.users WHERE api_token = ? LIMIT 1", apiKey).Scan(&user.ID, &user.Username, &user.ApiToken, &user.Contacts, &user.Channels)
	if err != nil {
		return types.User{}, err
	}

	return user, nil
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

func (c *Cassandra) GetMessagesOneToOne(channelID string, createdAt string) ([]types.Message, error) {
	m := []types.Message{}

	pt, err := types.ParseTime(createdAt)
	if err != nil {
		return nil, err
	}

	scanner := c.Session.Query("SELECT id, message_from, message_to, content, created_at, channel_id FROM chatmessages.messages where channel_id = ?  AND created_at < ? ORDER BY created_at DESC LIMIT 200", channelID, pt).Iter().Scanner()

	var id gocql.UUID
	var messageFrom gocql.UUID
	var messageTo gocql.UUID
	var content string
	var ct time.Time
	var channel gocql.UUID

	for scanner.Next() {
		err := scanner.Scan(&id, &messageFrom, &messageTo, &content, &ct, &channel)
		if err != nil {
			return nil, err
		}

		m = append(m, types.Message{
			ID:          id,
			MessageFrom: messageFrom,
			MessageTo:   messageTo,
			Content:     content,
			CreatedAt:   ct,
			ChannelID:   channel,
		})
	}

	return m, nil
}

func (c *Cassandra) GetChannels() ([]types.Channel, error) {
	channels := []types.Channel{}

	scanner := c.Session.Query("SELECT id, name, offset FROM chatmessages.channels").Iter().Scanner()

	var id gocql.UUID
	var name string
	var offset int64

	for scanner.Next() {
		err := scanner.Scan(&id, &name, &offset)
		if err != nil {
			return nil, err
		}

		channels = append(channels, types.Channel{
			ID:     id,
			Name:   name,
			Offset: offset,
		})
	}

	return channels, nil
}

func (c *Cassandra) UpdateChannelOffset(id gocql.UUID) error {
	var offset int
	err := c.Session.Query("SELECT offset FROM chatmessages.channels WHERE id = ? LIMIT 1", id).Scan(&offset)
	if err != nil {
		return err
	}

	err = c.Session.Query("UPDATE chatmessages.channels SET offset = ? WHERE id = ?", offset+1, id).Exec()
	return err
}

func (c *Cassandra) UpdateUserStatus(id gocql.UUID, status types.Status) error {
	err := c.Session.Query("UPDATE chatmessages.users SET status = ? WHERE id = ?", status, id).Exec()
	return err
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
