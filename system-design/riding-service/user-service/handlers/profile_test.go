package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"userservice/db"
	"userservice/models"

	"github.com/stretchr/testify/assert"
)

func TestGetProfile(t *testing.T) {
	setupTestDB(t)

	tests := []struct {
		name           string
		setupUser      bool
		expectedStatus int
	}{
		{
			name:           "Existing user",
			setupUser:      true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No user found",
			setupUser:      false,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupUser {
				testUser := models.User{
					Email:    "test@example.com",
					Password: "password123",
					Name:     "Test User",
				}
				db.DB.Create(&testUser)
			}

			req := httptest.NewRequest("GET", "/api/profile", nil)
			w := httptest.NewRecorder()

			GetProfile(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.setupUser {
				var response UserResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "test@example.com", response.Email)
				assert.Equal(t, "Test User", response.Name)
			}
		})
	}
}
