package db

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgresql() (*Postgres, error) {
	connectionString := "user=postgres password=1111 dbname=webcrawler sslmode=disable"
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}
	return &Postgres{db: db}, nil
}

func (p *Postgres) Init() error {
	_, err := p.db.Exec(`
		CREATE TABLE IF NOT EXISTS links (
			id SERIAL PRIMARY KEY,
			link TEXT NOT NULL UNIQUE,
			hash TEXT NOT NULL UNIQUE,
			html TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) CreateLink(link string, hash string, html string) error {
	_, err := p.db.Exec(`
		INSERT INTO links (link, hash, html, created_at) VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
	`, link, hash, html)
	if err != nil {
		return err
	}

	return nil
}
