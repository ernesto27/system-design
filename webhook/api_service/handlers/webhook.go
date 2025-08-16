package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"queue"

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

func HandleWebhook(c echo.Context, q queue.Queue) error {
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
	// Marshal event.Data to JSON string for Content field
	dataBytes, err := json.Marshal(event.Data)
	if err != nil {
		c.Logger().Errorf("Failed to marshal event data: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to process webhook event data",
		})
	}

	// Create queue message from webhook event
	message := queue.Message{
		ID:        event.ID,
		Content:   string(dataBytes),
		Timestamp: event.Timestamp,
		Type:      event.Type,
	}

	// Publish event to RabbitMQ
	ctx := context.Background()
	if err := q.Publish(ctx, message); err != nil {
		c.Logger().Errorf("Failed to publish message to queue: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to process webhook event",
		})
	}

	c.Logger().Infof("Successfully queued webhook event - Source: %s, Type: %s, ID: %s",
		event.Source, event.Type, event.ID)

	response := WebhookResponse{
		Status: "success",
	}

	return c.JSON(http.StatusOK, response)
}
