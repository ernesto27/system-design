// Package response provides standardized JSON response formatting for API endpoints
package response

import (
	"encoding/json"
	"net/http"
	"time"
)

// Response is the base structure for all API responses
type Response struct {
	Success   bool        `json:"success"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo contains detailed error information
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Standard error codes
const (
	ErrBadRequest          = "BAD_REQUEST"
	ErrUnauthorized        = "UNAUTHORIZED"
	ErrForbidden           = "FORBIDDEN"
	ErrNotFound            = "NOT_FOUND"
	ErrMethodNotAllowed    = "METHOD_NOT_ALLOWED"
	ErrConflict            = "CONFLICT"
	ErrValidation          = "VALIDATION_ERROR"
	ErrInternalServerError = "INTERNAL_SERVER_ERROR"
)

// HTTP status code to error code mapping
var statusToErrorCode = map[int]string{
	http.StatusBadRequest:          ErrBadRequest,
	http.StatusUnauthorized:        ErrUnauthorized,
	http.StatusForbidden:           ErrForbidden,
	http.StatusNotFound:            ErrNotFound,
	http.StatusMethodNotAllowed:    ErrMethodNotAllowed,
	http.StatusConflict:            ErrConflict,
	http.StatusInternalServerError: ErrInternalServerError,
}

// Success sends a successful JSON response with provided data
func Success(w http.ResponseWriter, statusCode int, data interface{}) {
	resp := Response{
		Success:   true,
		Timestamp: time.Now(),
		Data:      data,
	}

	writeJSON(w, statusCode, resp)
}

// Error sends an error JSON response
func Error(w http.ResponseWriter, statusCode int, message string, details string) {
	errorCode := statusToErrorCode[statusCode]
	if errorCode == "" {
		errorCode = ErrInternalServerError
	}

	resp := Response{
		Success:   false,
		Timestamp: time.Now(),
		Error: &ErrorInfo{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}

	writeJSON(w, statusCode, resp)
}

// ValidationError sends a validation error response with details
func ValidationError(w http.ResponseWriter, message string, details string) {
	resp := Response{
		Success:   false,
		Timestamp: time.Now(),
		Error: &ErrorInfo{
			Code:    ErrValidation,
			Message: message,
			Details: details,
		},
	}

	writeJSON(w, http.StatusBadRequest, resp)
}

// writeJSON writes the response as JSON to the http.ResponseWriter
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If JSON encoding fails, write a plain text error
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
