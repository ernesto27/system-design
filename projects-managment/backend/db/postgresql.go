package db

import (
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

type Postgres struct {
	Db *sql.DB
}

func NewPostgres(host, user, password, port, database, sslmode string) (*Postgres, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, database, sslmode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {

		db.Close()
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	return &Postgres{
		Db: db,
	}, nil
}

func RunMigrations(db *sql.DB, embedMigrations embed.FS) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}

	return nil
}
