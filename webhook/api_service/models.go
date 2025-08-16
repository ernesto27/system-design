package main

import (
	"time"
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
