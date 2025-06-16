package routes

import (
	"net/http"

	"twitterservice/internal/api/handlers"
	"twitterservice/internal/api/middleware"
	"twitterservice/internal/config"
	"twitterservice/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all routes for the application
func SetupRoutes(
	authService *services.AuthService,
	postService services.PostService,
	cfg *config.Config,
) *gin.Engine {
	// Create Gin router
	r := gin.Default()

	// Add basic middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// CORS middleware for development
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Create handlers
	authHandler := handlers.NewAuthHandler(authService)
	postHandler := handlers.NewPostHandler(postService)

	// Serve static files (for test HTML)
	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/templates/*")

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":     "Welcome to Twitter Service API",
			"version":     cfg.App.Version,
			"environment": cfg.App.Environment,
			"endpoints": map[string]string{
				"version":      "/api/version",
				"health":       "/health",
				"auth_test":    "/test",
				"google_login": "/auth/google/login",
			},
		})
	})

	// API version endpoint
	r.GET("/api/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version":     cfg.App.Version,
			"service":     cfg.App.Name,
			"environment": cfg.App.Environment,
			"status":      "active",
		})
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": cfg.App.Name,
			"version": cfg.App.Version,
		})
	})

	// Test page for Google OAuth
	r.GET("/test", func(c *gin.Context) {
		c.HTML(http.StatusOK, "auth_test.html", gin.H{
			"title":   "Twitter Service Auth Test",
			"baseURL": cfg.App.BaseURL,
		})
	})

	// Authentication routes
	authGroup := r.Group("/auth")
	{
		// Google OAuth routes
		authGroup.GET("/google/login", authHandler.GoogleLogin)
		authGroup.GET("/google/callback", authHandler.GoogleCallback)

		// Development/Testing endpoint
		if cfg.App.Environment == "development" {
			authGroup.POST("/test-login", authHandler.TestLogin)
		}

		// Protected routes (require authentication)
		protected := authGroup.Group("/")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			protected.GET("/profile", authHandler.GetProfile)
			protected.POST("/logout", authHandler.Logout)
		}
	}

	// API routes
	apiGroup := r.Group("/api/v1")
	{
		// Public routes
		apiGroup.GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  "ok",
				"version": cfg.App.Version,
			})
		})

		// Public posts routes
		apiGroup.GET("/posts/:id", postHandler.GetPost) // Get specific post (public)

		// Protected API routes
		protected := apiGroup.Group("/")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			// User routes
			userGroup := protected.Group("/users")
			{
				userGroup.GET("/me", authHandler.GetProfile)
				// Add more user endpoints here
			}

			// Posts routes
			postsGroup := protected.Group("/posts")
			{
				postsGroup.POST("", postHandler.CreatePost)                // Create post
				postsGroup.PUT("/:id", postHandler.UpdatePost)             // Update post
				postsGroup.DELETE("/:id", postHandler.DeletePost)          // Delete post
				postsGroup.GET("/my", postHandler.GetMyPosts)              // Get my posts
				postsGroup.GET("/user/:user_id", postHandler.GetUserPosts) // Get user posts
			}

			// Feed routes (future)
			// feedGroup := protected.Group("/feed")
			// {
			//     // Add feed endpoints here
			// }
		}
	}

	return r
}
