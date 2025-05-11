// Helper functions for testing and using the response package
package response

import (
	"log"
	"net/http"
)

// OK sends a 200 OK response with data
func OK(w http.ResponseWriter, data interface{}) {
	Success(w, http.StatusOK, data)
}

// Created sends a 201 Created response with data
func Created(w http.ResponseWriter, data interface{}) {
	Success(w, http.StatusCreated, data)
}

// NoContent sends a 204 No Content response
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// BadRequest sends a 400 Bad Request error response
func BadRequest(w http.ResponseWriter, message string, details string) {
	Error(w, http.StatusBadRequest, message, details)
}

// Unauthorized sends a 401 Unauthorized error response
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, message, "")
}

// Forbidden sends a 403 Forbidden error response
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, message, "")
}

// NotFound sends a 404 Not Found error response
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, message, "")
}

// MethodNotAllowed sends a 405 Method Not Allowed error response
func MethodNotAllowed(w http.ResponseWriter) {
	Error(w, http.StatusMethodNotAllowed, "Method not allowed", "")
}

// InternalServerError sends a 500 Internal Server Error response
func InternalServerError(w http.ResponseWriter, err error) {
	message := "Internal server error"
	details := ""

	if err != nil {
		log.Printf("Internal server error: %v", err)
		// In development environments, you might want to include error details
		// details = err.Error()
	}

	Error(w, http.StatusInternalServerError, message, details)
}
