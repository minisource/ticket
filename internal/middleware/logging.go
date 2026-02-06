package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/minisource/go-common/logging"
)

// LoggingMiddleware creates logging middleware
func LoggingMiddleware(logger logging.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get response status
		status := c.Response().StatusCode()

		// Log request
		extra := map[logging.ExtraKey]interface{}{
			"method":     c.Method(),
			"path":       c.Path(),
			"status":     status,
			"duration":   duration.String(),
			"ip":         c.IP(),
			"user_agent": c.Get("User-Agent"),
			"tenant_id":  c.Get("X-Tenant-ID"),
			"user_id":    c.Get("X-User-ID"),
			"request_id": c.Get("X-Request-ID"),
		}

		if err != nil {
			extra["error"] = err.Error()
		}

		// Log based on status code
		switch {
		case status >= 500:
			logger.Error(logging.General, logging.Api, "Request failed", extra)
		case status >= 400:
			logger.Warn(logging.General, logging.Api, "Request warning", extra)
		default:
			logger.Info(logging.General, logging.Api, "Request completed", extra)
		}

		return err
	}
}

// RequestIDMiddleware adds request ID to context
func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
			c.Request().Header.Set("X-Request-ID", requestID)
		}

		c.Set("X-Request-ID", requestID)
		c.Locals("request_id", requestID)

		return c.Next()
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}
