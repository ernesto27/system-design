package types

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/gorilla/websocket"
)

type Status int

const (
	StatusOffline Status = iota
	StatusOnline
)

const TypeNewMessage = "newMessage"
const TypeDeleteMessage = "deleteMessage"
const TypeUpdateStatus = "updateStatus"

type Request struct {
	ID          string `json:"id"`
	Content     string `json:"content"`
	MessageFrom string `json:"messageFrom"`
	MessageTo   string `json:"messageTo"`
	ChannelID   string `json:"channelID"`
	CreatedAt   string `json:"createdAt"`
	Type        string `json:"type"`
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type Client struct {
	Conn    *websocket.Conn
	User    User
	Seconds int
}

func (c *Client) UpdateSeconds(value int) {
	c.Seconds = value
}

type UserStatus struct {
	UserID gocql.UUID `json:"userID"`
	Status Status     `json:"status"`
}

type Message struct {
	ID          gocql.UUID `json:"id"`
	MessageFrom gocql.UUID `json:"messageFrom"`
	MessageTo   gocql.UUID `json:"messageTo"`
	Content     string     `json:"content"`
	CreatedAt   time.Time  `json:"createdAt"`
	ChannelID   gocql.UUID `json:"channelID"`
	Type        string     `json:"type"`
}

func NewMessage(r Request) (Message, error) {
	id, err := gocql.ParseUUID(r.ID)
	if err != nil {
		return Message{}, err
	}

	channelID, err := gocql.ParseUUID(r.ChannelID)
	if err != nil {
		return Message{}, err
	}

	createdAt, err := ParseTime(r.CreatedAt)
	if err != nil {
		return Message{}, err
	}

	mf, err := gocql.ParseUUID(r.MessageFrom)
	if err != nil {
		return Message{}, err
	}

	m := Message{
		ID:          id,
		MessageFrom: mf,
		ChannelID:   channelID,
		CreatedAt:   createdAt,
	}

	return m, nil
}

type User struct {
	ID        gocql.UUID
	Username  string
	Password  string
	ApiToken  string
	Contacts  []gocql.UUID
	Channels  []gocql.UUID
	CreatedAt time.Time
}

func (u *User) GetTopicName(contactUUID gocql.UUID) string {
	t := u.ID.String() + "_" + contactUUID.String()
	if contactUUID.String() < u.ID.String() {
		t = contactUUID.String() + "_" + u.ID.String()
	}

	return t
}

func (u *User) ValidateContact(contact string) bool {
	for _, c := range u.Contacts {
		if c.String() == contact {
			return true
		}
	}

	return false
}

func (u *User) ValidateChannel(channel string) bool {
	for _, c := range u.Channels {
		if c.String() == channel {
			return true
		}
	}

	return false
}

type Channel struct {
	ID     gocql.UUID `json:"id"`
	Name   string     `json:"name"`
	Offset int64      `json:"offset"`
}

func ParseTime(createdAt string) (time.Time, error) {
	var parsedTime time.Time
	if createdAt == "" {
		parsedTime = time.Now()
	} else {
		layout := "2006-01-02T15:04:05.99Z"
		var err error
		parsedTime, err = time.Parse(layout, createdAt)
		if err != nil {
			return parsedTime, err
		}
	}

	return parsedTime, nil
}
