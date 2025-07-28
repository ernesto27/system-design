package main

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (u User) MarshalJSON() ([]byte, error) {
	type Alias User
	return json.Marshal(&struct {
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		*Alias
	}{
		CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: u.UpdatedAt.Format("2006-01-02 15:04:05"),
		Alias:     (*Alias)(&u),
	})
}

type Problem struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Title       string         `gorm:"not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Difficulty  string         `gorm:"not null" json:"difficulty"`
	TestCases   datatypes.JSON `gorm:"type:jsonb" json:"test_cases"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

func (p Problem) MarshalJSON() ([]byte, error) {
	type Alias Problem
	return json.Marshal(&struct {
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		*Alias
	}{
		CreatedAt: p.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: p.UpdatedAt.Format("2006-01-02 15:04:05"),
		Alias:     (*Alias)(&p),
	})
}

type Submission struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	ProblemID uint      `gorm:"not null" json:"problem_id"`
	Code      string    `gorm:"type:text;not null" json:"code"`
	Language  string    `gorm:"not null" json:"language"`
	Status    string    `gorm:"not null" json:"status"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (s Submission) MarshalJSON() ([]byte, error) {
	type Alias Submission
	return json.Marshal(&struct {
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		*Alias
	}{
		CreatedAt: s.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: s.UpdatedAt.Format("2006-01-02 15:04:05"),
		Alias:     (*Alias)(&s),
	})
}

type SubmissionRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Language string `json:"language" binding:"required"`
}

type TestCase struct {
	Status string `json:"status"`
}

type SubmissionResponse struct {
	Status    string     `json:"status"`
	TestCases []TestCase `json:"test_cases"`
}

type CodeBase struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProblemID uint      `gorm:"not null;index" json:"problem_id"`
	Language  string    `gorm:"not null;index" json:"language"`
	Template  string    `gorm:"type:text;not null" json:"template"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	
	// Relationships
	Problem Problem `gorm:"foreignKey:ProblemID" json:"problem,omitempty"`
}

func (c CodeBase) MarshalJSON() ([]byte, error) {
	type Alias CodeBase
	return json.Marshal(&struct {
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		*Alias
	}{
		CreatedAt: c.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: c.UpdatedAt.Format("2006-01-02 15:04:05"),
		Alias:     (*Alias)(&c),
	})
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &Problem{}, &Submission{}, &CodeBase{})
}
