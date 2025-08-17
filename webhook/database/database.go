package database

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

type WebhookEvent struct {
	ID        uuid.UUID              `json:"id" db:"id"`
	EventID   *string                `json:"event_id" db:"event_id"`
	Source    string                 `json:"source" db:"source"`
	Type      string                 `json:"type" db:"type"`
	Data      map[string]interface{} `json:"data" db:"data"`
	Timestamp time.Time              `json:"timestamp" db:"timestamp"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}

func Connect(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func RunMigrations(db *sql.DB) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}

	log.Println("Migrations completed successfully")
	return nil
}

func SaveWebhookEvent(db *sql.DB, event WebhookEvent) error {
	// Convert Data map to JSON for PostgreSQL storage
	dataJSON, err := json.Marshal(event.Data)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO webhook_events (id, event_id, source, type, data, timestamp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = db.Exec(query,
		event.ID,
		event.EventID,
		event.Source,
		event.Type,
		dataJSON, // Use JSON bytes instead of map
		event.Timestamp,
		event.CreatedAt,
	)

	return err
}
