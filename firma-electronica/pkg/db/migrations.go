package db

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AutoMigrate automatically migrates all models
func (db *DB) AutoMigrate() error {
	log.Println("Running auto migrations...")

	// Register custom types if needed (for PostgreSQL enums)
	if err := db.registerEnumTypes(); err != nil {
		return fmt.Errorf("failed to register enum types: %w", err)
	}

	// Run migrations for all models
	if err := db.DB.AutoMigrate(
		&User{},
		&Document{},
		&Signer{},
		&Signature{},
		&DocumentAccessLog{},
		&Notification{},
	); err != nil {
		return fmt.Errorf("failed to run auto migrations: %w", err)
	}

	log.Println("Auto migrations completed successfully")
	return nil
}

// SeedTestUser creates a default test user if it doesn't exist
func (db *DB) SeedTestUser() error {
	log.Println("Seeding test user...")

	// Check if test user already exists
	var count int64
	if err := db.DB.Model(&User{}).Where("email = ?", "test@example.com").Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check if test user exists: %w", err)
	}

	// If test user doesn't exist, create it
	if count == 0 {
		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("1111"), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		// Create test user
		testUser := User{
			ID:           uuid.New(),
			Email:        "test@example.com",
			FirstName:    "Test",
			LastName:     "User",
			PasswordHash: string(hashedPassword),
		}

		if err := db.DB.Create(&testUser).Error; err != nil {
			return fmt.Errorf("failed to create test user: %w", err)
		}

		log.Printf("Test user created successfully with ID: %s", testUser.ID)
	} else {
		log.Println("Test user already exists, skipping creation")
	}

	return nil
}

// RegisterEnumTypes creates custom enum types if they don't exist
func (db *DB) registerEnumTypes() error {
	// Check if document_status type exists, create if not
	var exists bool
	err := db.DB.Raw("SELECT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'document_status')").Scan(&exists).Error
	if err != nil {
		return err
	}

	if !exists {
		if err := db.DB.Exec(`CREATE TYPE document_status AS ENUM ('draft', 'pending', 'completed', 'canceled')`).Error; err != nil {
			return err
		}
	}

	// Check if signer_status type exists, create if not
	err = db.DB.Raw("SELECT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'signer_status')").Scan(&exists).Error
	if err != nil {
		return err
	}

	if !exists {
		if err := db.DB.Exec(`CREATE TYPE signer_status AS ENUM ('pending', 'signed', 'declined', 'expired')`).Error; err != nil {
			return err
		}
	}

	// Check if notification_type type exists, create if not
	err = db.DB.Raw("SELECT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'notification_type')").Scan(&exists).Error
	if err != nil {
		return err
	}

	if !exists {
		if err := db.DB.Exec(`CREATE TYPE notification_type AS ENUM ('invitation', 'reminder', 'confirmation', 'signed_confirmation')`).Error; err != nil {
			return err
		}
	}

	// Check if notification_status type exists, create if not
	err = db.DB.Raw("SELECT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'notification_status')").Scan(&exists).Error
	if err != nil {
		return err
	}

	if !exists {
		if err := db.DB.Exec(`CREATE TYPE notification_status AS ENUM ('sent', 'delivered', 'opened', 'failed')`).Error; err != nil {
			return err
		}
	}

	// Check if access_action type exists, create if not
	err = db.DB.Raw("SELECT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'access_action')").Scan(&exists).Error
	if err != nil {
		return err
	}

	if !exists {
		if err := db.DB.Exec(`CREATE TYPE access_action AS ENUM ('viewed', 'downloaded', 'shared', 'printed')`).Error; err != nil {
			return err
		}
	}

	return nil
}
