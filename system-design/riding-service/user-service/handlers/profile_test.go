package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"userservice/db"
	"userservice/models"
	"userservice/testutils"
	"userservice/utils"

	"github.com/stretchr/testify/assert"
)

func TestGetProfile(t *testing.T) {
	db.DB = testutils.SetupTestDB(t)

	user := models.User{
		ID:       1,
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	if err := db.DB.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	token, err := utils.GenerateJWT(user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{
			name:           "Success with valid token",
			token:          token,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Fail with no token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req := httptest.NewRequest("GET", "/api/profile", nil)
			w := httptest.NewRecorder()

			req.Header.Set("Authorization", "Bearer "+tt.token)

			GetProfile(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

		})
	}
}
