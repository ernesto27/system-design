package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	GoogleID    string    `json:"google_id" gorm:"unique;not null"`
	Email       string    `json:"email" gorm:"unique;not null"`
	Username    string    `json:"username" gorm:"unique"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	Bio         string    `json:"bio"`
	Location    string    `json:"location"`
	Website     string    `json:"website"`

	// Counts
	FollowerCount  int `json:"follower_count" gorm:"default:0"`
	FollowingCount int `json:"following_count" gorm:"default:0"`
	PostCount      int `json:"post_count" gorm:"default:0"`

	// Status
	IsVerified bool `json:"is_verified" gorm:"default:false"`
	IsPrivate  bool `json:"is_private" gorm:"default:false"`
	IsActive   bool `json:"is_active" gorm:"default:true"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate hook to generate UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for User
func (User) TableName() string {
	return "users"
}
