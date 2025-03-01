package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"userservice/db"
	"userservice/models"
	"userservice/testutils"

	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) {
	db.DB = testutils.SetupTestDB(t)
}

func TestRegister(t *testing.T) {
	setupTestDB(t)

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "Valid registration",
			payload: map[string]interface{}{
				"email":    "test@example.com",
				"password": "password123",
				"name":     "Test User",
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name: "Missing password",
			payload: map[string]interface{}{
				"email": "test@example.com",
				"name":  "Test User",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "Invalid email",
			payload: map[string]interface{}{
				"email":    "invalid-email",
				"password": "password123",
				"name":     "Test User",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payloadBytes, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(payloadBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			Register(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if !tt.expectedError {
				var response models.User
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.payload["email"], response.Email)
				assert.Equal(t, tt.payload["name"], response.Name)
				assert.Empty(t, response.Password) // Password should not be in response
			}
		})
	}
}

func TestLogin(t *testing.T) {
	setupTestDB(t)

	// Create a test user
	testUser := models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}
	testUser.HashPassword()
	db.DB.Create(&testUser)

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		expectedToken  bool
	}{
		{
			name: "Valid login",
			payload: map[string]interface{}{
				"email":    "test@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusOK,
			expectedToken:  true,
		},
		{
			name: "Invalid password",
			payload: map[string]interface{}{
				"email":    "test@example.com",
				"password": "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedToken:  false,
		},
		{
			name: "User not found",
			payload: map[string]interface{}{
				"email":    "nonexistent@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedToken:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payloadBytes, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(payloadBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			Login(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedToken {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response["token"])
			}
		})
	}
}
