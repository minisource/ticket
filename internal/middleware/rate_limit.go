package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/minisource/go-common/i18n"
	"github.com/minisource/go-common/response"
)

// RateLimitConfig holds rate limit configuration
type RateLimitConfig struct {
	Max        int           // Maximum number of requests
	Expiration time.Duration // Time window
	KeyFunc    func(*fiber.Ctx) string
}

// DefaultRateLimitConfig returns default rate limit config
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Max:        100,
		Expiration: time.Minute,
		KeyFunc: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}
}

// RateLimiter holds rate limit state
type RateLimiter struct {
	mu      sync.RWMutex
	entries map[string]*rateLimitEntry
	config  RateLimitConfig
}

type rateLimitEntry struct {
	count     int
	expiresAt time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		entries: make(map[string]*rateLimitEntry),
		config:  config,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// cleanup removes expired entries
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		rl.mu.Lock()
		for key, entry := range rl.entries {
			if now.After(entry.expiresAt) {
				delete(rl.entries, key)
			}
		}
		rl.mu.Unlock()
	}
}

// Handler returns the rate limit middleware handler
func (rl *RateLimiter) Handler() fiber.Handler {
	translator := i18n.GetTranslator()

	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		key := rl.config.KeyFunc(c)
		now := time.Now()

		rl.mu.Lock()
		entry, exists := rl.entries[key]

		if !exists || now.After(entry.expiresAt) {
			rl.entries[key] = &rateLimitEntry{
				count:     1,
				expiresAt: now.Add(rl.config.Expiration),
			}
			rl.mu.Unlock()
			return c.Next()
		}

		if entry.count >= rl.config.Max {
			rl.mu.Unlock()
			return response.New().
				Status(fiber.StatusTooManyRequests).
				Error("RATE_LIMIT_EXCEEDED", translator.Translate(ctx, "error.rate_limit_exceeded", nil)).
				Send(c)
		}

		entry.count++
		rl.mu.Unlock()

		// Set rate limit headers
		c.Set("X-RateLimit-Limit", string(rune(rl.config.Max)))
		c.Set("X-RateLimit-Remaining", string(rune(rl.config.Max-entry.count)))

		return c.Next()
	}
}

// RateLimitMiddleware creates a rate limit middleware with default config
func RateLimitMiddleware() fiber.Handler {
	config := DefaultRateLimitConfig()
	limiter := NewRateLimiter(config)
	return limiter.Handler()
}

// RateLimitMiddlewareWithConfig creates a rate limit middleware with custom config
func RateLimitMiddlewareWithConfig(max int, expiration time.Duration) fiber.Handler {
	config := RateLimitConfig{
		Max:        max,
		Expiration: expiration,
		KeyFunc: func(c *fiber.Ctx) string {
			// Use tenant + IP for rate limiting
			tenantID := c.Get("X-Tenant-ID")
			if tenantID != "" {
				return tenantID + ":" + c.IP()
			}
			return c.IP()
		},
	}
	limiter := NewRateLimiter(config)
	return limiter.Handler()
}
