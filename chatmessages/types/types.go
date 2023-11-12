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

const TypeMessage = "message"
const TypeUpdateStatus = "updateStatus"

type Request struct {
	Content   string `json:"content"`
	MessageTo string `json:"messageTo"`
	ChannelID string `json:"channelID"`
	CreatedAt string `json:"createdAt"`
	Type      string `json:"type"`
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
