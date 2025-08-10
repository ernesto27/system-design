package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type WebhookEvent struct {
	ID        string                 `json:"id"`
	Source    string                 `json:"source"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

type WebhookResponse struct {
	Status string `json:"status"`
}

func HandleWebhook(c echo.Context) error {
	var event WebhookEvent
	
	// Bind and validate request
	if err := c.Bind(&event); err != nil {
		c.Logger().Errorf("Failed to bind request: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
		})
	}

	// Basic validation
	if event.Source == "" || event.Type == "" {
		c.Logger().Warn("Missing required fields: source or type")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing required fields: source and type are required",
		})
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	
	// TODO: Add event to message queue here
	// For now, just log the received event
	c.Logger().Infof("Received webhook event - Source: %s, Type: %s, ID: %s", 
		event.Source, event.Type, event.ID)
	
	response := WebhookResponse{
		Status: "success",
	}
	
	return c.JSON(http.StatusOK, response)
}