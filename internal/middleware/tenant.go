package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/go-common/i18n"
	"github.com/minisource/go-common/response"
)

// TenantMiddleware creates tenant middleware
func TenantMiddleware() fiber.Handler {
	translator := i18n.GetTranslator()

	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		// First check if tenant_id is set from auth middleware
		tenantID, ok := c.Locals("tenant_id").(string)
		if ok && tenantID != "" {
			c.Request().Header.Set("X-Tenant-ID", tenantID)
			return c.Next()
		}

		// Check X-Tenant-ID header
		tenantID = c.Get("X-Tenant-ID")
		if tenantID == "" {
			return response.New().
				Status(fiber.StatusBadRequest).
				Error("TENANT_REQUIRED", translator.Translate(ctx, "error.tenant_required", nil)).
				Send(c)
		}

		// Set tenant ID in locals
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}
}

// OptionalTenantMiddleware creates optional tenant middleware
func OptionalTenantMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// First check if tenant_id is set from auth middleware
		tenantID, ok := c.Locals("tenant_id").(string)
		if ok && tenantID != "" {
			c.Request().Header.Set("X-Tenant-ID", tenantID)
			return c.Next()
		}

		// Check X-Tenant-ID header
		tenantID = c.Get("X-Tenant-ID")
		if tenantID != "" {
			c.Locals("tenant_id", tenantID)
		}

		return c.Next()
	}
}
