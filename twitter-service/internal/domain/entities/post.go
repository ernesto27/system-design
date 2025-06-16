package entities

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
	ID        uuid.UUID `json:"id" cql:"id"`
	UserID    uuid.UUID `json:"user_id" cql:"user_id"`
	Content   string    `json:"content" cql:"content"`
	CreatedAt time.Time `json:"created_at" cql:"created_at"`
	UpdatedAt time.Time `json:"updated_at" cql:"updated_at"`
	IsDeleted bool      `json:"is_deleted" cql:"is_deleted"`
}

type CreatePostRequest struct {
	Content string `json:"content" binding:"required,min=1,max=280"`
}

type PostResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (p *Post) ToResponse() PostResponse {
	return PostResponse{
		ID:        p.ID,
		UserID:    p.UserID,
		Content:   p.Content,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
