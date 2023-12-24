package types

import "time"

type Request struct {
	UserID int
	Text   string `json:"text"`
}

type JSONResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    []Post `json:"data"`
}

type Post struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}
