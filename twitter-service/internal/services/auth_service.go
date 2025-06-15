package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"twitterservice/internal/config"
	"twitterservice/internal/domain/entities"
	"twitterservice/internal/domain/repositories"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// AuthService handles authentication operations
type AuthService struct {
	userRepo    repositories.UserRepository
	config      *config.Config
	oauthConfig *oauth2.Config
}

// GoogleUserInfo represents user info from Google
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// LoginResponse represents the response after successful login
type LoginResponse struct {
	User        *entities.User `json:"user"`
	AccessToken string         `json:"access_token"`
	TokenType   string         `json:"token_type"`
	ExpiresIn   int64          `json:"expires_in"`
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo repositories.UserRepository, cfg *config.Config) *AuthService {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.OAuth.GoogleClientID,
		ClientSecret: cfg.OAuth.GoogleClientSecret,
		RedirectURL:  cfg.OAuth.GoogleRedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &AuthService{
		userRepo:    userRepo,
		config:      cfg,
		oauthConfig: oauthConfig,
	}
}

// GetGoogleLoginURL returns the Google OAuth login URL
func (s *AuthService) GetGoogleLoginURL(state string) string {
	return s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// HandleGoogleCallback handles the Google OAuth callback
func (s *AuthService) HandleGoogleCallback(ctx context.Context, code string) (*LoginResponse, error) {
	// Exchange code for token
	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Get user info from Google
	userInfo, err := s.getGoogleUserInfo(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Check if user exists
	user, err := s.userRepo.GetUserByGoogleID(ctx, userInfo.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by Google ID: %w", err)
	}

	// If user doesn't exist, create new user
	if user == nil {
		user, err = s.createUserFromGoogleInfo(ctx, userInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		// Update user info from Google
		user.DisplayName = userInfo.Name
		user.AvatarURL = userInfo.Picture
		if err := s.userRepo.UpdateUser(ctx, user); err != nil {
			logrus.WithError(err).Warn("Failed to update user info from Google")
		}
	}

	// Generate JWT token
	accessToken, expiresIn, err := s.generateJWT(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	return &LoginResponse{
		User:        user,
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
	}, nil
}

// ValidateJWT validates a JWT token and returns the user
func (s *AuthService) ValidateJWT(tokenString string) (*entities.User, error) {
	// Remove "Bearer " prefix if present
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWT.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Get user from database
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in token: %w", err)
	}

	user, err := s.userRepo.GetUserByID(context.Background(), userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// getGoogleUserInfo fetches user info from Google API
func (s *AuthService) getGoogleUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error) {
	client := s.oauthConfig.Client(ctx, token)
	response, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(response.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// createUserFromGoogleInfo creates a new user from Google user info
func (s *AuthService) createUserFromGoogleInfo(ctx context.Context, userInfo *GoogleUserInfo) (*entities.User, error) {
	// Generate username from email
	username := s.generateUsernameFromEmail(userInfo.Email)

	// Ensure username is unique
	existingUser, _ := s.userRepo.GetUserByUsername(ctx, username)
	if existingUser != nil {
		username = fmt.Sprintf("%s_%s", username, uuid.New().String()[:8])
	}

	user := &entities.User{
		GoogleID:    userInfo.ID,
		Email:       userInfo.Email,
		Username:    username,
		DisplayName: userInfo.Name,
		AvatarURL:   userInfo.Picture,
		IsActive:    true,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// generateUsernameFromEmail generates a username from email
func (s *AuthService) generateUsernameFromEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) > 0 {
		// Remove dots and special characters
		username := strings.ReplaceAll(parts[0], ".", "")
		username = strings.ReplaceAll(username, "+", "")
		return strings.ToLower(username)
	}
	return "user"
}

// generateJWT generates a JWT token for a user
func (s *AuthService) generateJWT(user *entities.User) (string, int64, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &JWTClaims{
		UserID:   user.ID.String(),
		Email:    user.Email,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    s.config.App.Name,
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.JWT.Secret))
	if err != nil {
		return "", 0, err
	}

	return tokenString, int64(24 * 60 * 60), nil // 24 hours in seconds
}
