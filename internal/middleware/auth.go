package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/minisource/go-common/i18n"
	"github.com/minisource/go-common/response"
	"github.com/minisource/go-sdk/auth"
	"github.com/minisource/ticket/config"
)

// AuthMiddleware creates authentication middleware
func AuthMiddleware(cfg *config.Config) fiber.Handler {
	authClient := auth.NewClient(auth.ClientConfig{
		BaseURL:      cfg.Auth.ServiceURL,
		ClientID:     cfg.Auth.ClientID,
		ClientSecret: cfg.Auth.ClientSecret,
	})
	translator := i18n.GetTranslator()

	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		// Get token from header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Unauthorized(c, translator.Translate(ctx, "error.unauthorized", nil))
		}

		// Extract token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return response.Unauthorized(c, translator.Translate(ctx, "error.invalid_token", nil))
		}

		token := parts[1]

		// Validate token
		claims, err := authClient.ValidateToken(ctx, token)
		if err != nil {
			return response.Unauthorized(c, translator.Translate(ctx, "error.invalid_token", nil))
		}

		userID := claims.UserID
		if userID == "" {
			userID = claims.ClientID
		}
		// Set user info in context
		c.Locals("user_id", userID)
		c.Locals("roles", claims.Roles)
		c.Locals("permissions", claims.Permissions)
		c.Locals("service_name", claims.ServiceName)
		c.Locals("scopes", claims.Scopes)

		// Set headers for downstream handlers
		if userID != "" {
			c.Request().Header.Set("X-User-ID", userID)
		}
		if len(claims.Roles) > 0 {
			c.Request().Header.Set("X-User-Roles", strings.Join(claims.Roles, ","))
		}

		return c.Next()
	}
}

// OptionalAuthMiddleware creates optional authentication middleware
func OptionalAuthMiddleware(cfg *config.Config) fiber.Handler {
	authClient := auth.NewClient(auth.ClientConfig{
		BaseURL:      cfg.Auth.ServiceURL,
		ClientID:     cfg.Auth.ClientID,
		ClientSecret: cfg.Auth.ClientSecret,
	})

	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		// Get token from header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		// Extract token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.Next()
		}

		token := parts[1]

		// Validate token
		claims, err := authClient.ValidateToken(ctx, token)
		if err != nil {
			return c.Next()
		}

		userID := claims.UserID
		if userID == "" {
			userID = claims.ClientID
		}
		// Set user info in context
		c.Locals("user_id", userID)
		c.Locals("roles", claims.Roles)
		c.Locals("permissions", claims.Permissions)
		c.Locals("service_name", claims.ServiceName)
		c.Locals("scopes", claims.Scopes)

		if userID != "" {
			c.Request().Header.Set("X-User-ID", userID)
		}

		return c.Next()
	}
}

// RequireRole creates middleware that requires specific roles
func RequireRole(roles ...string) fiber.Handler {
	translator := i18n.GetTranslator()

	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		userRoles, ok := c.Locals("roles").([]string)
		if !ok {
			return response.Forbidden(c, translator.Translate(ctx, "error.forbidden", nil))
		}

		// Check if user has any of the required roles
		for _, requiredRole := range roles {
			for _, userRole := range userRoles {
				if userRole == requiredRole {
					return c.Next()
				}
			}
		}

		return response.Forbidden(c, translator.Translate(ctx, "error.forbidden", nil))
	}
}

// RequirePermission creates middleware that requires specific permissions
func RequirePermission(permissions ...string) fiber.Handler {
	translator := i18n.GetTranslator()

	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		userPermissions, ok := c.Locals("permissions").([]string)
		if !ok {
			return response.Forbidden(c, translator.Translate(ctx, "error.forbidden", nil))
		}

		// Check if user has all required permissions
		for _, requiredPerm := range permissions {
			found := false
			for _, userPerm := range userPermissions {
				if userPerm == requiredPerm {
					found = true
					break
				}
			}
			if !found {
				return response.Forbidden(c, translator.Translate(ctx, "error.forbidden", nil))
			}
		}

		return c.Next()
	}
}

// AgentMiddleware creates middleware that checks if user is an agent
func AgentMiddleware() fiber.Handler {
	return RequireRole("agent", "admin", "supervisor", "super_admin")
}

// AdminMiddleware creates middleware that checks if user is an admin
func AdminMiddleware() fiber.Handler {
	return RequireRole("admin", "supervisor", "super_admin")
}
