package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestHandleWebhook(t *testing.T) {
	e := echo.New()

	t.Run("Valid webhook request", func(t *testing.T) {
		json := `{
			"id": "evt_123",
			"source": "shopify",
			"type": "order.created",
			"data": {
				"order_id": "12345",
				"customer": "john@example.com"
			}
		}`

		req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(json))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := HandleWebhook(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"status":"success"`)
	})

	t.Run("Missing source field", func(t *testing.T) {
		json := `{
			"id": "evt_124",
			"type": "order.created",
			"data": {}
		}`

		req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(json))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := HandleWebhook(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Missing required fields")
	})

	t.Run("Missing type field", func(t *testing.T) {
		json := `{
			"id": "evt_125",
			"source": "shopify",
			"data": {}
		}`

		req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(json))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := HandleWebhook(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Missing required fields")
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		json := `{invalid json}`

		req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(json))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := HandleWebhook(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid request payload")
	})
}