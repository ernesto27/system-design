package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/ernesto/riding-service/shared/config"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(cfg *config.Config) error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	// Initialize SQL DB for goose migrations
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// Run migrations
	if err := runMigrations(sqlDB); err != nil {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	// Initialize GORM
	DB, err = gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})

	return err
}

func runMigrations(db *sql.DB) error {
	// Set migrations location
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	// Run migrations
	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}

	log.Println("Migrations completed successfully")
	return nil
}
