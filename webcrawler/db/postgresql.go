package db

import (
	"database/sql"
	"webcrawler/types"

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
			url TEXT NOT NULL UNIQUE,
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

func (p *Postgres) CreateLink(link types.Link) error {
	_, err := p.db.Exec(`
		INSERT INTO links (url, hash, keywords, description, html, created_at) VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
	`, link.Url, link.Hash, link.Keywords, link.Description, "")
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) CreateImage(url string, urlImage string, path string, hash string) error {
	_, err := p.db.Exec(`
			INSERT INTO images (url, url_image, path, hash, created_at) VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
	`, url, urlImage, path, hash)
	if err != nil {
		return err
	}

	return nil
}
