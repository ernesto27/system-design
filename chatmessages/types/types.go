package types

import "github.com/gorilla/websocket"

type Request struct {
	Message   string `json:"message"`
	MessageTo string `json:"messageTo"`
	ChannelID string `json:"channelID"`
	CreatedAt string `json:"createdAt"`
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type Client struct {
	Conn *websocket.Conn
}
