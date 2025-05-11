package db

import (
	"time"

	"github.com/google/uuid"
)

// DocumentStatus represents document status enum
type DocumentStatus string

// SignerStatus represents signer status enum
type SignerStatus string

// NotificationType represents notification type enum
type NotificationType string

// NotificationStatus represents notification status enum
type NotificationStatus string

// AccessAction represents access action enum
type AccessAction string

// Enum values
const (
	// Document status values
	DocumentStatusDraft     DocumentStatus = "draft"
	DocumentStatusPending   DocumentStatus = "pending"
	DocumentStatusCompleted DocumentStatus = "completed"
	DocumentStatusCanceled  DocumentStatus = "canceled"

	// Signer status values
	SignerStatusPending  SignerStatus = "pending"
	SignerStatusSigned   SignerStatus = "signed"
	SignerStatusDeclined SignerStatus = "declined"
	SignerStatusExpired  SignerStatus = "expired"

	// Notification type values
	NotificationTypeInvitation         NotificationType = "invitation"
	NotificationTypeReminder           NotificationType = "reminder"
	NotificationTypeConfirmation       NotificationType = "confirmation"
	NotificationTypeSignedConfirmation NotificationType = "signed_confirmation"

	// Notification status values
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusDelivered NotificationStatus = "delivered"
	NotificationStatusOpened    NotificationStatus = "opened"
	NotificationStatusFailed    NotificationStatus = "failed"

	// Access action values
	AccessActionViewed     AccessAction = "viewed"
	AccessActionDownloaded AccessAction = "downloaded"
	AccessActionShared     AccessAction = "shared"
	AccessActionPrinted    AccessAction = "printed"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email        string    `gorm:"type:varchar(255);not null;unique"`
	FirstName    string    `gorm:"type:varchar(100)"`
	LastName     string    `gorm:"type:varchar(100)"`
	PasswordHash string    `gorm:"type:varchar(255)"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`

	// Relationships
	Documents []Document `gorm:"foreignKey:UserID"`
}

// Document represents a document to be signed
type Document struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Title       string         `gorm:"type:varchar(255);not null"`
	Description string         `gorm:"type:text"`
	ContentURL  string         `gorm:"type:varchar(255)"`
	ContentHash string         `gorm:"type:varchar(64)"`
	Status      DocumentStatus `gorm:"type:document_status;default:'draft'"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	ExpiresAt   *time.Time

	// Relationships
	User       User        `gorm:"foreignKey:UserID"`
	Signers    []Signer    `gorm:"foreignKey:DocumentID"`
	Signatures []Signature `gorm:"foreignKey:DocumentID"`
}

// Signer represents a person who needs to sign a document
type Signer struct {
	ID         uint         `gorm:"primaryKey;autoIncrement"`
	DocumentID uuid.UUID    `gorm:"type:uuid;not null"`
	Email      string       `gorm:"type:varchar(255);not null"`
	FirstName  string       `gorm:"type:varchar(100)"`
	LastName   string       `gorm:"type:varchar(100)"`
	Hash       string       `gorm:"type:varchar(100)"`
	Status     SignerStatus `gorm:"type:signer_status;default:'pending'"`
	CreatedAt  time.Time    `gorm:"autoCreateTime"`
	UpdatedAt  time.Time    `gorm:"autoUpdateTime"`

	// Relationships
	Document      Document       `gorm:"foreignKey:DocumentID"`
	Signatures    []Signature    `gorm:"foreignKey:SignerID"`
	Notifications []Notification `gorm:"foreignKey:SignerID"`
}

// Signature represents a signature made on a document
type Signature struct {
	ID            uint      `gorm:"primaryKey;autoIncrement"`
	SignerID      uint      `gorm:"not null"`
	DocumentID    uuid.UUID `gorm:"type:uuid;not null"`
	SignatureData string    `gorm:"type:text"`
	IPAddress     string    `gorm:"type:varchar(45)"`
	UserAgent     string    `gorm:"type:text"`
	SignedAt      time.Time `gorm:"autoCreateTime"`

	// Relationships
	Signer   Signer   `gorm:"foreignKey:SignerID"`
	Document Document `gorm:"foreignKey:DocumentID"`
}

// DocumentAccessLog represents access logs for documents
type DocumentAccessLog struct {
	ID         uint         `gorm:"primaryKey;autoIncrement"`
	DocumentID uuid.UUID    `gorm:"type:uuid;not null"`
	UserID     *uuid.UUID   `gorm:"type:uuid"`
	Email      string       `gorm:"type:varchar(255)"`
	IPAddress  string       `gorm:"type:varchar(45)"`
	Action     AccessAction `gorm:"type:access_action;not null"`
	AccessedAt time.Time    `gorm:"autoCreateTime"`

	// Relationships
	Document Document `gorm:"foreignKey:DocumentID"`
	User     *User    `gorm:"foreignKey:UserID"`
}

// Notification represents notifications sent to signers
type Notification struct {
	ID          uint               `gorm:"primaryKey;autoIncrement"`
	SignerID    uint               `gorm:"not null"`
	DocumentID  uuid.UUID          `gorm:"type:uuid;not null"`
	Type        NotificationType   `gorm:"type:notification_type;not null"`
	Status      NotificationStatus `gorm:"type:notification_status;not null"`
	SentAt      time.Time          `gorm:"autoCreateTime"`
	DeliveredAt *time.Time
	OpenedAt    *time.Time

	// Relationships
	Signer   Signer   `gorm:"foreignKey:SignerID"`
	Document Document `gorm:"foreignKey:DocumentID"`
}

// TableName overrides for proper enum type mapping
func (Document) TableName() string {
	return "documents"
}

func (User) TableName() string {
	return "users"
}

func (Signer) TableName() string {
	return "signers"
}

func (Signature) TableName() string {
	return "signatures"
}

func (DocumentAccessLog) TableName() string {
	return "document_access_logs"
}

func (Notification) TableName() string {
	return "notifications"
}
