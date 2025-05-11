package controllers

import (
	"bytes"
	"encoding/json"
	"firmaelectronica/pkg/auth"
	"firmaelectronica/pkg/db"
	"firmaelectronica/pkg/response"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

var TestConfig = db.Config{
	Host:     "localhost",
	Port:     5433,
	Name:     "firma_electronica_test",
	User:     "postgres",
	Password: "postgres",
	LogLevel: "error",
}

var TestDB *db.DB

var TestJWTService *auth.Service

func setupTestDB() error {
	var err error
	TestDB, err = db.New(TestConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to test database: %v", err)
	}

	if err := TestDB.AutoMigrate(); err != nil {
		return fmt.Errorf("failed to migrate test database: %v", err)
	}

	jwtConfig := auth.Config{
		Secret:     "test-secret",
		Expiration: 24 * time.Hour,
	}
	TestJWTService = auth.NewService(jwtConfig)

	return nil
}

func clearTestDB() error {
	tables := []string{"notifications", "document_access_logs", "signatures", "signers", "documents", "users"}

	for _, table := range tables {
		if err := TestDB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)).Error; err != nil {
			return fmt.Errorf("failed to truncate table %s: %v", table, err)
		}
	}
	return nil
}

func createTestUser(email, password, firstName, lastName string) (*db.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	user := &db.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		FirstName:    firstName,
		LastName:     lastName,
	}

	result := TestDB.Create(user)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create test user: %v", result.Error)
	}

	return user, nil
}

func TestMain(m *testing.M) {
	if err := setupTestDB(); err != nil {
		log.Fatalf("Failed to set up test database: %v", err)
	}

	exitCode := m.Run()

	if sqlDB, err := TestDB.DB.DB(); err == nil {
		sqlDB.Close()
	}

	os.Exit(exitCode)
}

func TestHelloHandler(t *testing.T) {
	controller := &Controller{
		DB:         nil,
		JWTService: nil,
	}

	req, err := http.NewRequest("GET", "/api/hello", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(controller.HelloHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var responseObj response.Response
	if err := json.Unmarshal(rr.Body.Bytes(), &responseObj); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !responseObj.Success {
		t.Errorf("Response success flag is false, expected true")
	}

	if time.Since(responseObj.Timestamp) > time.Minute {
		t.Errorf("Response timestamp is too old: %v", responseObj.Timestamp)
	}

	data, ok := responseObj.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Response data is not a map: %v", responseObj.Data)
	}

	message, ok := data["message"].(string)
	if !ok {
		t.Fatalf("Response message is not a string: %v", data["message"])
	}

	if message != "Hello, World!" {
		t.Errorf("Response message is incorrect: got %v want %v", message, "Hello, World!")
	}
}

func TestLoginHandler(t *testing.T) {
	if err := clearTestDB(); err != nil {
		t.Fatalf("Failed to clear test database: %v", err)
	}

	testUser, err := createTestUser("test@example.com", "testpassword", "Test", "User")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	controller := &Controller{
		DB:         TestDB,
		JWTService: TestJWTService,
	}

	tests := []struct {
		name           string
		method         string
		body           interface{}
		setupFunc      func() error
		cleanupFunc    func() error
		expectedStatus int
		validateBody   func(*testing.T, []byte)
	}{
		{
			name:   "Successful Login",
			method: http.MethodPost,
			body: LoginRequest{
				Email:    "test@example.com",
				Password: "testpassword",
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var resp response.Response
				err := json.Unmarshal(body, &resp)
				assert.NoError(t, err)
				assert.True(t, resp.Success)

				data, ok := resp.Data.(map[string]interface{})
				assert.True(t, ok)
				token, ok := data["token"].(string)
				assert.True(t, ok)
				assert.NotEmpty(t, token)

				claims, err := TestJWTService.ValidateToken(token)
				assert.NoError(t, err)
				assert.Equal(t, testUser.ID.String(), claims.UserID)
				assert.Equal(t, testUser.Email, claims.Email)
			},
		},
		{
			name:           "Invalid Method",
			method:         http.MethodGet,
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
			validateBody: func(t *testing.T, body []byte) {
				var resp response.Response
				err := json.Unmarshal(body, &resp)
				assert.NoError(t, err)
				assert.False(t, resp.Success)
				assert.NotNil(t, resp.Error)
				assert.Equal(t, response.ErrMethodNotAllowed, resp.Error.Code)
			},
		},
		{
			name:           "Invalid Request Body",
			method:         http.MethodPost,
			body:           "invalid-json",
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body []byte) {
				var resp response.Response
				err := json.Unmarshal(body, &resp)
				assert.NoError(t, err)
				assert.False(t, resp.Success)
				assert.NotNil(t, resp.Error)
				assert.Equal(t, response.ErrBadRequest, resp.Error.Code)
			},
		},
		{
			name:           "Empty Request Body",
			method:         http.MethodPost,
			body:           LoginRequest{},
			expectedStatus: http.StatusUnauthorized,
			validateBody: func(t *testing.T, body []byte) {
				var resp response.Response
				err := json.Unmarshal(body, &resp)
				assert.NoError(t, err)
				assert.False(t, resp.Success)
				assert.NotNil(t, resp.Error)
				assert.Equal(t, response.ErrUnauthorized, resp.Error.Code)
				assert.Equal(t, "Invalid credentials", resp.Error.Message)
			},
		},
		{
			name:   "User Not Found",
			method: http.MethodPost,
			body: LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password",
			},
			expectedStatus: http.StatusUnauthorized,
			validateBody: func(t *testing.T, body []byte) {
				var resp response.Response
				err := json.Unmarshal(body, &resp)
				assert.NoError(t, err)
				assert.False(t, resp.Success)
				assert.NotNil(t, resp.Error)
				assert.Equal(t, response.ErrUnauthorized, resp.Error.Code)
				assert.Equal(t, "Invalid credentials", resp.Error.Message)
			},
		},
		{
			name:   "Invalid Password",
			method: http.MethodPost,
			body: LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			validateBody: func(t *testing.T, body []byte) {
				var resp response.Response
				err := json.Unmarshal(body, &resp)
				assert.NoError(t, err)
				assert.False(t, resp.Success)
				assert.NotNil(t, resp.Error)
				assert.Equal(t, response.ErrUnauthorized, resp.Error.Code)
				assert.Equal(t, "Invalid credentials", resp.Error.Message)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				if err := tt.setupFunc(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			var reqBody []byte
			var err error

			switch v := tt.body.(type) {
			case string:
				reqBody = []byte(v)
			case nil:
				reqBody = nil
			default:
				reqBody, err = json.Marshal(v)
				assert.NoError(t, err)
			}

			req, err := http.NewRequest(tt.method, "/api/login", bytes.NewBuffer(reqBody))
			assert.NoError(t, err)
			if reqBody != nil {
				req.Header.Set("Content-Type", "application/json")
			}

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(controller.LoginHandler)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.Bytes())
			}

			if tt.cleanupFunc != nil {
				if err := tt.cleanupFunc(); err != nil {
					t.Fatalf("Cleanup failed: %v", err)
				}
			}
		})
	}
}
