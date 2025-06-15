package main

import (
	"twitterservice/internal/api/routes"
	"twitterservice/internal/config"
	"twitterservice/internal/domain/repositories"
	"twitterservice/internal/infrastructure/database"
	"twitterservice/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load configuration")
	}

	// Set up logging
	logrus.SetFormatter(&logrus.JSONFormatter{})
	if cfg.App.Environment == "development" {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	// Set Gin mode
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to database
	if err := database.Connect(cfg); err != nil {
		logrus.WithError(err).Fatal("Failed to connect to database")
	}

	// Run migrations
	if err := database.AutoMigrate(); err != nil {
		logrus.WithError(err).Fatal("Failed to run database migrations")
	}

	// Initialize repositories
	userRepo := repositories.NewUserRepository()

	// Initialize services
	authService := services.NewAuthService(userRepo, cfg)

	// Setup routes
	router := routes.SetupRoutes(authService, cfg)

	// Start server
	logrus.WithFields(logrus.Fields{
		"port":        cfg.App.Port,
		"environment": cfg.App.Environment,
		"service":     cfg.App.Name,
		"version":     cfg.App.Version,
	}).Info("Starting Twitter Service server")

	if err := router.Run(":" + cfg.App.Port); err != nil {
		logrus.WithError(err).Fatal("Failed to start server")
	}
}
