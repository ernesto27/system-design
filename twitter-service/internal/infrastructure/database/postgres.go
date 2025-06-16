package database

import (
	"fmt"
	"strings"
	"time"

	"twitterservice/internal/config"
	"twitterservice/internal/domain/entities"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global database instance
var DB *gorm.DB

// Connect connects to the PostgreSQL database
func Connect(cfg *config.Config) error {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)

	// Configure GORM logger
	gormLogger := logger.Default
	if cfg.App.Environment == "production" {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db
	logrus.Info("Connected to PostgreSQL database")

	return nil
}

// AutoMigrate runs database migrations
func AutoMigrate() error {
	if DB == nil {
		return fmt.Errorf("database not connected")
	}

	// Enable UUID extension
	if err := DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		logrus.Warn("Failed to create uuid-ossp extension (might already exist)")
	}

	// Run auto migrations for tables
	if err := DB.AutoMigrate(&entities.User{}, &entities.Follow{}); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create database functions and triggers
	if err := createDatabaseFunctions(); err != nil {
		return fmt.Errorf("failed to create database functions: %w", err)
	}

	// Create indexes
	if err := createIndexes(); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	// Create triggers
	if err := createTriggers(); err != nil {
		return fmt.Errorf("failed to create triggers: %w", err)
	}

	// Add constraints
	if err := addConstraints(); err != nil {
		return fmt.Errorf("failed to add constraints: %w", err)
	}

	// Seed test data in development
	cfg, _ := config.Load()
	if cfg != nil && cfg.App.Environment == "development" {
		if err := seedTestData(); err != nil {
			logrus.WithError(err).Warn("Failed to seed test data")
		}
	}

	logrus.Info("Database migrations completed")
	return nil
}

// createDatabaseFunctions creates necessary database functions
func createDatabaseFunctions() error {
	// Function to update updated_at timestamp
	updateTimestampFunction := `
		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ language 'plpgsql';
	`

	if err := DB.Exec(updateTimestampFunction).Error; err != nil {
		return fmt.Errorf("failed to create update_updated_at_column function: %w", err)
	}

	// Function to update user follow counts when a follow relationship is created
	followCountInsertFunction := `
		CREATE OR REPLACE FUNCTION update_follow_counts_on_insert()
		RETURNS TRIGGER AS $$
		BEGIN
			-- Increment follower count for the user being followed
			UPDATE users SET follower_count = follower_count + 1 WHERE id = NEW.following_id;
			
			-- Increment following count for the user doing the following
			UPDATE users SET following_count = following_count + 1 WHERE id = NEW.follower_id;
			
			RETURN NEW;
		END;
		$$ language 'plpgsql';
	`

	if err := DB.Exec(followCountInsertFunction).Error; err != nil {
		return fmt.Errorf("failed to create update_follow_counts_on_insert function: %w", err)
	}

	// Function to update user follow counts when a follow relationship is deleted
	followCountDeleteFunction := `
		CREATE OR REPLACE FUNCTION update_follow_counts_on_delete()
		RETURNS TRIGGER AS $$
		BEGIN
			-- Decrement follower count for the user being unfollowed
			UPDATE users SET follower_count = follower_count - 1 WHERE id = OLD.following_id;
			
			-- Decrement following count for the user doing the unfollowing
			UPDATE users SET following_count = following_count - 1 WHERE id = OLD.follower_id;
			
			RETURN OLD;
		END;
		$$ language 'plpgsql';
	`

	if err := DB.Exec(followCountDeleteFunction).Error; err != nil {
		return fmt.Errorf("failed to create update_follow_counts_on_delete function: %w", err)
	}

	return nil
}

// createIndexes creates database indexes for better performance
func createIndexes() error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_google_id ON users(google_id)",
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
		"CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)",
		"CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_follows_follower_id ON follows(follower_id)",
		"CREATE INDEX IF NOT EXISTS idx_follows_following_id ON follows(following_id)",
		"CREATE INDEX IF NOT EXISTS idx_follows_created_at ON follows(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_follows_follower_following ON follows(follower_id, following_id)",
	}

	for _, indexSQL := range indexes {
		if err := DB.Exec(indexSQL).Error; err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// createTriggers creates database triggers
func createTriggers() error {
	triggers := []string{
		`CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
		 FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()`,
		`CREATE TRIGGER update_follows_updated_at BEFORE UPDATE ON follows
		 FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()`,
		`CREATE TRIGGER trigger_follow_counts_insert 
		 AFTER INSERT ON follows
		 FOR EACH ROW EXECUTE FUNCTION update_follow_counts_on_insert()`,
		`CREATE TRIGGER trigger_follow_counts_delete 
		 AFTER DELETE ON follows
		 FOR EACH ROW EXECUTE FUNCTION update_follow_counts_on_delete()`,
	}

	for _, triggerSQL := range triggers {
		// Drop trigger if exists first, then create
		triggerName := ""
		if triggerSQL[14:] != "" {
			parts := strings.Fields(triggerSQL)
			if len(parts) > 2 {
				triggerName = parts[2]
			}
		}

		if triggerName != "" {
			dropSQL := fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON users", triggerName)
			if strings.Contains(triggerSQL, "follows") {
				dropSQL = fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON follows", triggerName)
			}
			DB.Exec(dropSQL) // Ignore errors here
		}

		if err := DB.Exec(triggerSQL).Error; err != nil {
			logrus.WithError(err).Warnf("Failed to create trigger (might already exist): %s", triggerSQL[:50])
		}
	}

	return nil
}

// addConstraints adds database constraints
func addConstraints() error {
	constraints := []string{
		`ALTER TABLE follows ADD CONSTRAINT IF NOT EXISTS check_no_self_follow 
		 CHECK (follower_id != following_id)`,
		`ALTER TABLE follows ADD CONSTRAINT IF NOT EXISTS unique_follow_relationship 
		 UNIQUE (follower_id, following_id)`,
	}

	for _, constraintSQL := range constraints {
		if err := DB.Exec(constraintSQL).Error; err != nil {
			logrus.WithError(err).Warnf("Failed to add constraint (might already exist): %s", constraintSQL[:50])
		}
	}

	return nil
}

// seedTestData seeds the database with test data for development
func seedTestData() error {
	// Check if test data already exists
	var count int64
	if err := DB.Model(&entities.User{}).Where("google_id LIKE ?", "test_google_id_%").Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check existing test data: %w", err)
	}

	if count > 0 {
		logrus.Info("Test data already exists, skipping seeding")
		return nil
	}

	// Create test users
	testUsers := []entities.User{
		{
			GoogleID:    "test_google_id_1",
			Email:       "alice@example.com",
			Username:    "alice_dev",
			DisplayName: "Alice Developer",
			AvatarURL:   "https://example.com/avatar1.jpg",
			Bio:         "Full-stack developer passionate about Go and React",
			Location:    "San Francisco, CA",
			Website:     "https://alice-dev.com",
			IsVerified:  false,
			IsPrivate:   false,
			IsActive:    true,
		},
		{
			GoogleID:    "test_google_id_2",
			Email:       "bob@example.com",
			Username:    "bob_design",
			DisplayName: "Bob Designer",
			AvatarURL:   "https://example.com/avatar2.jpg",
			Bio:         "UI/UX Designer creating beautiful digital experiences",
			Location:    "New York, NY",
			Website:     "https://bobdesign.io",
			IsVerified:  true,
			IsPrivate:   false,
			IsActive:    true,
		},
		{
			GoogleID:    "test_google_id_3",
			Email:       "charlie@example.com",
			Username:    "charlie_tech",
			DisplayName: "Charlie Tech",
			AvatarURL:   "https://example.com/avatar3.jpg",
			Bio:         "Tech entrepreneur and startup advisor",
			Location:    "Austin, TX",
			Website:     "https://charlietech.com",
			IsVerified:  false,
			IsPrivate:   true,
			IsActive:    true,
		},
		{
			GoogleID:    "test_google_id_4",
			Email:       "diana@example.com",
			Username:    "diana_data",
			DisplayName: "Diana Data",
			AvatarURL:   "https://example.com/avatar4.jpg",
			Bio:         "Data scientist specializing in machine learning",
			Location:    "Seattle, WA",
			Website:     "https://diana-data.com",
			IsVerified:  false,
			IsPrivate:   false,
			IsActive:    true,
		},
		{
			GoogleID:    "test_google_id_5",
			Email:       "eve@example.com",
			Username:    "eve_product",
			DisplayName: "Eve Product",
			AvatarURL:   "https://example.com/avatar5.jpg",
			Bio:         "Product manager building the future of tech",
			Location:    "Los Angeles, CA",
			Website:     "https://eveproduct.com",
			IsVerified:  true,
			IsPrivate:   false,
			IsActive:    true,
		},
	}

	// Insert test users
	for i := range testUsers {
		if err := DB.Create(&testUsers[i]).Error; err != nil {
			return fmt.Errorf("failed to create test user %s: %w", testUsers[i].Username, err)
		}
	}

	logrus.Info("Test data seeded successfully")
	return nil
}

// Close closes the database connection
func Close() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}
