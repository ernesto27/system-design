package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// Config holds all configuration for the application
type Config struct {
	App       AppConfig
	Database  DatabaseConfig
	Cassandra CassandraConfig
	OAuth     OAuthConfig
	JWT       JWTConfig
}

// AppConfig holds application configuration
type AppConfig struct {
	Name        string
	Version     string
	Environment string
	Port        string
	BaseURL     string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// CassandraConfig holds Cassandra configuration
type CassandraConfig struct {
	Host     string
	Port     int
	Keyspace string
	Timeout  time.Duration
}

// OAuthConfig holds OAuth configuration
type OAuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret    string
	ExpiresIn string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found")
	}

	config := &Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "twitter-service"),
			Version:     getEnv("API_VERSION", "1.0.0"),
			Environment: getEnv("APP_ENV", "development"),
			Port:        getEnv("APP_PORT", "8080"),
			BaseURL:     getEnv("BASE_URL", "http://localhost:8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "twitter_user"),
			Password: getEnv("DB_PASSWORD", "twitter_password"),
			DBName:   getEnv("DB_NAME", "twitter_db"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Cassandra: CassandraConfig{
			Host:     getEnv("CASSANDRA_HOST", "localhost"),
			Port:     getEnvAsInt("CASSANDRA_PORT", 9042),
			Keyspace: getEnv("CASSANDRA_KEYSPACE", "twitter_keyspace"),
			Timeout:  getEnvAsDuration("CASSANDRA_TIMEOUT", "10s"),
		},
		OAuth: OAuthConfig{
			GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/google/callback"),
		},
		JWT: JWTConfig{
			Secret:    getEnv("JWT_SECRET", "your-secret-key"),
			ExpiresIn: getEnv("JWT_EXPIRES_IN", "24h"),
		},
	}

	return config, nil
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvAsInt gets an environment variable as int with a fallback value
func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

// getEnvAsDuration gets an environment variable as duration with a fallback value
func getEnvAsDuration(key string, fallback string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	duration, _ := time.ParseDuration(fallback)
	return duration
}
