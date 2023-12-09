package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type SQLite struct {
	db *sql.DB
}

func New() (*SQLite, error) {
	db, err := sql.Open("sqlite3", "./webcrawler.db")
	if err != nil {
		return nil, err
	}
	return &SQLite{db: db}, nil
}

func (s *SQLite) Init() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS links (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			link TEXT NOT NULL UNIQUE,
			created_at DATE
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

func (s *SQLite) CreateLink(link string) error {
	_, err := s.db.Exec(`
		INSERT INTO links (link, created_at) VALUES (?, datetime('now'))
	`, link)
	if err != nil {
		return err
	}

	return nil
}
