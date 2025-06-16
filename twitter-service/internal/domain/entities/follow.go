package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Follow represents a following relationship between users
type Follow struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	FollowerID  uuid.UUID `json:"follower_id" gorm:"type:uuid;not null;index"`
	FollowingID uuid.UUID `json:"following_id" gorm:"type:uuid;not null;index"`

	// Relations
	Follower  User `json:"follower" gorm:"foreignKey:FollowerID;references:ID"`
	Following User `json:"following" gorm:"foreignKey:FollowingID;references:ID"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate hook to generate UUID
func (f *Follow) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for Follow
func (Follow) TableName() string {
	return "follows"
}
