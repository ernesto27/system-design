package middleware

import (
	"net/http"
	"strings"

	"twitterservice/internal/services"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware creates an authentication middleware
func AuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Check if it's a Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Bearer token required",
			})
			c.Abort()
			return
		}

		// Validate JWT token
		user, err := authService.ValidateJWT(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		// Set user in context
		c.Set("user", user)
		c.Set("user_id", user.ID.String())

		c.Next()
	}
}

// OptionalAuthMiddleware creates an optional authentication middleware
func OptionalAuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			// Validate JWT token
			user, err := authService.ValidateJWT(authHeader)
			if err == nil {
				// Set user in context if token is valid
				c.Set("user", user)
				c.Set("user_id", user.ID.String())
			}
		}

		c.Next()
	}
}
