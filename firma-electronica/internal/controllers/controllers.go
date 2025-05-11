package controllers

import (
	"context"
	"encoding/json"
	"firmaelectronica/pkg/auth"
	"firmaelectronica/pkg/db"
	"firmaelectronica/pkg/response"
	"log"
	"net/http"
	"strings"
)

// contextKey is a custom type to prevent context key collisions
type contextKey string

// Context keys
const (
	UserClaimsKey contextKey = "userClaims"
)

// Controller holds dependencies for all handlers
type Controller struct {
	DB         *db.DB
	JWTService *auth.Service
}

// LoginRequest represents login request body
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents login response with JWT token
type LoginResponse struct {
	Token string `json:"token"`
}

// NewController creates a new controller instance
func NewController(database *db.DB, jwtService *auth.Service) *Controller {
	return &Controller{
		DB:         database,
		JWTService: jwtService,
	}
}

// HelloHandler handles the hello endpoint
func (c *Controller) HelloHandler(w http.ResponseWriter, r *http.Request) {
	response.OK(w, map[string]string{"message": "Hello, World!"})
}

// LoginHandler authenticates a user and returns a JWT token
func (c *Controller) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}

	// Parse request body
	var loginReq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		log.Printf("Error decoding login request: %v", err)
		response.BadRequest(w, "Invalid request body", err.Error())
		return
	}

	// Find user by email
	var user db.User
	if err := c.DB.Where("email = ?", loginReq.Email).First(&user).Error; err != nil {
		log.Printf("User not found: %v", err)
		response.Unauthorized(w, "Invalid credentials")
		return
	}

	// Verify password using bcrypt
	if err := c.DB.VerifyPassword(loginReq.Password, user.PasswordHash); err != nil {
		log.Printf("Password verification failed: %v", err)
		response.Unauthorized(w, "Invalid credentials")
		return
	}

	// Generate JWT token for user
	tokenString, err := c.JWTService.GenerateToken(&user)
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		response.InternalServerError(w, err)
		return
	}

	// Return token in response
	response.OK(w, LoginResponse{Token: tokenString})
}

// AuthMiddleware validates the JWT token and adds user info to the request context
func (c *Controller) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.Unauthorized(w, "Authorization header required")
			return
		}

		// Remove "Bearer " prefix if present
		tokenString := authHeader
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = authHeader[7:]
		}

		// Validate the token
		claims, err := c.JWTService.ValidateToken(tokenString)
		if err != nil {
			log.Printf("Invalid token: %v", err)
			response.Unauthorized(w, "Invalid or expired token")
			return
		}

		// Token is valid, add claims to request context
		ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
		r = r.WithContext(ctx)

		log.Printf("Authenticated user: %s (%s)", claims.Email, claims.UserID)
		next.ServeHTTP(w, r)
	})
}
