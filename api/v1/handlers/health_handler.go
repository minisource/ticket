package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/go-common/response"
)

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health returns service health status
// @Summary Health check
// @Tags Health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) Health(c *fiber.Ctx) error {
	return response.OK(c, map[string]string{
		"service": "ticket-service",
		"status":  "healthy",
	})
}

// Ready returns service readiness status
// @Summary Readiness check
// @Tags Health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /ready [get]
func (h *HealthHandler) Ready(c *fiber.Ctx) error {
	return response.OK(c, map[string]string{
		"service": "ticket-service",
		"status":  "ready",
	})
}

// Live returns service liveness status
// @Summary Liveness check
// @Tags Health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /live [get]
func (h *HealthHandler) Live(c *fiber.Ctx) error {
	return response.OK(c, map[string]string{
		"service": "ticket-service",
		"status":  "live",
	})
}

// HealthResponse represents health check response for swagger
type HealthResponse struct {
	Service string `json:"service"`
	Status  string `json:"status"`
}
