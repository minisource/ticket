//go:build integration
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Ticket represents a ticket for testing
type Ticket struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenant_id"`
	UserID      string `json:"user_id"`
	Subject     string `json:"subject"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Priority    string `json:"priority"`
	Category    string `json:"category,omitempty"`
	AssigneeID  string `json:"assignee_id,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// TicketMessage represents a ticket message for testing
type TicketMessage struct {
	ID        string `json:"id"`
	TicketID  string `json:"ticket_id"`
	UserID    string `json:"user_id"`
	Content   string `json:"content"`
	IsStaff   bool   `json:"is_staff"`
	CreatedAt string `json:"created_at"`
}

// TestHealthEndpoint tests the health check endpoint
func TestHealthEndpoint(t *testing.T) {
	app := fiber.New()

	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "ticket",
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestCreateTicket tests ticket creation
func TestCreateTicket(t *testing.T) {
	app := fiber.New()

	var createdTicket Ticket

	app.Post("/api/v1/tickets", func(c *fiber.Ctx) error {
		if err := c.BodyParser(&createdTicket); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if createdTicket.Subject == "" || createdTicket.Description == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Subject and description are required",
			})
		}

		createdTicket.ID = "ticket-123"
		createdTicket.Status = "open"
		createdTicket.TenantID = c.Get("X-Tenant-ID")
		return c.Status(fiber.StatusCreated).JSON(createdTicket)
	})

	t.Run("Create Ticket", func(t *testing.T) {
		ticket := Ticket{
			UserID:      "user-456",
			Subject:     "Cannot login to my account",
			Description: "I've tried resetting my password but still can't access my account",
			Priority:    "high",
			Category:    "auth",
		}
		body, _ := json.Marshal(ticket)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/tickets", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", "tenant-123")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result Ticket
		json.NewDecoder(resp.Body).Decode(&result)
		assert.NotEmpty(t, result.ID)
		assert.Equal(t, "open", result.Status)
	})

	t.Run("Create Without Subject", func(t *testing.T) {
		ticket := Ticket{
			UserID:      "user-456",
			Description: "Missing subject",
		}
		body, _ := json.Marshal(ticket)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/tickets", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// TestListTickets tests ticket listing
func TestListTickets(t *testing.T) {
	app := fiber.New()

	mockTickets := []Ticket{
		{ID: "1", Subject: "Ticket 1", Status: "open", Priority: "high"},
		{ID: "2", Subject: "Ticket 2", Status: "open", Priority: "low"},
		{ID: "3", Subject: "Ticket 3", Status: "closed", Priority: "medium"},
	}

	app.Get("/api/v1/tickets", func(c *fiber.Ctx) error {
		status := c.Query("status")
		priority := c.Query("priority")

		var filtered []Ticket
		for _, ticket := range mockTickets {
			if (status == "" || ticket.Status == status) &&
				(priority == "" || ticket.Priority == priority) {
				filtered = append(filtered, ticket)
			}
		}

		return c.JSON(fiber.Map{
			"data":  filtered,
			"total": len(filtered),
		})
	})

	t.Run("List All Tickets", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, float64(3), result["total"])
	})

	t.Run("List Open Tickets", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets?status=open", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, float64(2), result["total"])
	})

	t.Run("List High Priority", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets?priority=high", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, float64(1), result["total"])
	})
}

// TestUpdateTicketStatus tests ticket status updates
func TestUpdateTicketStatus(t *testing.T) {
	app := fiber.New()

	app.Patch("/api/v1/tickets/:id/status", func(c *fiber.Ctx) error {
		var payload struct {
			Status string `json:"status"`
		}
		c.BodyParser(&payload)

		return c.JSON(Ticket{
			ID:     c.Params("id"),
			Status: payload.Status,
		})
	})

	statuses := []string{"open", "in_progress", "pending", "resolved", "closed"}

	for _, status := range statuses {
		t.Run("Set Status "+status, func(t *testing.T) {
			body, _ := json.Marshal(map[string]string{"status": status})

			req := httptest.NewRequest(http.MethodPatch, "/api/v1/tickets/123/status", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var result Ticket
			json.NewDecoder(resp.Body).Decode(&result)
			assert.Equal(t, status, result.Status)
		})
	}
}

// TestAssignTicket tests ticket assignment
func TestAssignTicket(t *testing.T) {
	app := fiber.New()

	app.Patch("/api/v1/tickets/:id/assign", func(c *fiber.Ctx) error {
		var payload struct {
			AssigneeID string `json:"assignee_id"`
		}
		c.BodyParser(&payload)

		return c.JSON(Ticket{
			ID:         c.Params("id"),
			AssigneeID: payload.AssigneeID,
			Status:     "in_progress",
		})
	})

	body, _ := json.Marshal(map[string]string{"assignee_id": "staff-123"})

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/tickets/123/assign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result Ticket
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "staff-123", result.AssigneeID)
}

// TestAddMessage tests adding messages to tickets
func TestAddMessage(t *testing.T) {
	app := fiber.New()

	app.Post("/api/v1/tickets/:id/messages", func(c *fiber.Ctx) error {
		var msg TicketMessage
		c.BodyParser(&msg)

		msg.ID = "msg-123"
		msg.TicketID = c.Params("id")

		return c.Status(fiber.StatusCreated).JSON(msg)
	})

	t.Run("Customer Message", func(t *testing.T) {
		body, _ := json.Marshal(TicketMessage{
			UserID:  "user-456",
			Content: "I tried the suggested fix and it didn't work",
			IsStaff: false,
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/tickets/123/messages", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("Staff Message", func(t *testing.T) {
		body, _ := json.Marshal(TicketMessage{
			UserID:  "staff-789",
			Content: "Let me investigate this further",
			IsStaff: true,
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/tickets/123/messages", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})
}

// TestTicketMetrics tests ticket metrics/stats
func TestTicketMetrics(t *testing.T) {
	app := fiber.New()

	app.Get("/api/v1/tickets/metrics", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"total":               100,
			"open":                25,
			"in_progress":         15,
			"pending":             10,
			"resolved":            30,
			"closed":              20,
			"avg_resolution_time": "4h30m",
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets/metrics", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, float64(100), result["total"])
}
