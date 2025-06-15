package handlers

import (
	"net/http"

	"twitterservice/internal/domain/entities"
	"twitterservice/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// GoogleLogin redirects to Google OAuth
// @Summary Redirect to Google OAuth login
// @Description Redirects user to Google OAuth consent screen
// @Tags authentication
// @Success 302 {string} string "Redirect to Google"
// @Router /auth/google/login [get]
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	// Generate state parameter (in production, store this securely)
	state := uuid.New().String()

	// Get Google OAuth URL
	url := h.authService.GetGoogleLoginURL(state)

	logrus.WithFields(logrus.Fields{
		"state": state,
		"url":   url,
	}).Info("Redirecting to Google OAuth")

	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback handles Google OAuth callback
// @Summary Handle Google OAuth callback
// @Description Processes Google OAuth callback and returns JWT token
// @Tags authentication
// @Param code query string true "OAuth authorization code"
// @Param state query string true "OAuth state parameter"
// @Success 200 {object} services.LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/google/callback [get]
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Authorization code not provided",
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"code":  code[:10] + "...", // Log only first 10 chars for security
		"state": state,
	}).Info("Processing Google OAuth callback")

	// Handle the callback
	response, err := h.authService.HandleGoogleCallback(c.Request.Context(), code)
	if err != nil {
		logrus.WithError(err).Error("Failed to handle Google callback")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to authenticate with Google",
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id": response.User.ID,
		"email":   response.User.Email,
	}).Info("User authenticated successfully")

	c.JSON(http.StatusOK, response)
}

// GetProfile returns the current user's profile
// @Summary Get current user profile
// @Description Returns the profile of the authenticated user
// @Tags authentication
// @Security BearerAuth
// @Success 200 {object} entities.User
// @Failure 401 {object} map[string]string
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	userEntity := user.(*entities.User)
	c.JSON(http.StatusOK, userEntity)
}

// Logout handles user logout
// @Summary Logout user
// @Description Logs out the current user (client should discard token)
// @Tags authentication
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// In a stateless JWT system, we just return success
	// The client should discard the token
	// For a more secure approach, you could maintain a blacklist of tokens

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}
