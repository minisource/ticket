package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/ticket/api/v1/handlers"
	"github.com/minisource/ticket/config"
	"github.com/minisource/ticket/internal/middleware"
)

// Router holds router dependencies
type Router struct {
	app           *fiber.App
	config        *config.Config
	logger        logging.Logger
	ticketHandler *handlers.TicketHandler
	adminHandler  *handlers.AdminHandler
	healthHandler *handlers.HealthHandler
}

// NewRouter creates a new router
func NewRouter(
	cfg *config.Config,
	logger logging.Logger,
	ticketHandler *handlers.TicketHandler,
	adminHandler *handlers.AdminHandler,
	healthHandler *handlers.HealthHandler,
) *Router {
	app := fiber.New(fiber.Config{
		AppName:      "Ticket Service",
		ErrorHandler: customErrorHandler,
	})

	return &Router{
		app:           app,
		config:        cfg,
		logger:        logger,
		ticketHandler: ticketHandler,
		adminHandler:  adminHandler,
		healthHandler: healthHandler,
	}
}

// Setup sets up all routes
func (r *Router) Setup() *fiber.App {
	// Global middleware
	r.app.Use(recover.New())
	r.app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Tenant-ID,X-User-ID,X-Request-ID,Accept-Language",
	}))
	r.app.Use(middleware.RequestIDMiddleware())
	r.app.Use(middleware.LoggingMiddleware(r.logger))

	// Health check routes
	r.app.Get("/health", r.healthHandler.Health)
	r.app.Get("/ready", r.healthHandler.Ready)
	r.app.Get("/live", r.healthHandler.Live)

	// Swagger documentation
	r.app.Get("/swagger/*", swagger.HandlerDefault)

	// API v1 routes
	api := r.app.Group("/api/v1")

	// Public routes (require tenant)
	public := api.Group("")
	public.Use(middleware.OptionalAuthMiddleware(r.config))
	public.Use(middleware.OptionalTenantMiddleware())

	// Departments and categories (public read)
	public.Get("/departments", r.adminHandler.ListDepartments)
	public.Get("/departments/:id", r.adminHandler.GetDepartment)
	public.Get("/categories", r.adminHandler.ListCategories)
	public.Get("/categories/:id", r.adminHandler.GetCategory)

	// Authenticated routes
	authenticated := api.Group("")
	authenticated.Use(middleware.AuthMiddleware(r.config))
	authenticated.Use(middleware.TenantMiddleware())

	// Customer ticket routes
	r.setupTicketRoutes(authenticated)

	// Agent routes
	agent := authenticated.Group("")
	agent.Use(middleware.AgentMiddleware())
	r.setupAgentRoutes(agent)

	// Admin routes
	admin := authenticated.Group("/admin")
	admin.Use(middleware.AdminMiddleware())
	r.setupAdminRoutes(admin)

	return r.app
}

// setupTicketRoutes sets up customer ticket routes
func (r *Router) setupTicketRoutes(group fiber.Router) {
	tickets := group.Group("/tickets")

	// Ticket CRUD
	tickets.Post("", r.ticketHandler.CreateTicket)
	tickets.Get("", r.ticketHandler.ListTickets)
	tickets.Get("/stats", r.ticketHandler.GetStats)
	tickets.Get("/number/:number", r.ticketHandler.GetTicketByNumber)
	tickets.Get("/:id", r.ticketHandler.GetTicket)
	tickets.Patch("/:id", r.ticketHandler.UpdateTicket)
	tickets.Delete("/:id", r.ticketHandler.DeleteTicket)

	// Ticket actions
	tickets.Patch("/:id/status", r.ticketHandler.ChangeStatus)
	tickets.Post("/:id/rate", r.ticketHandler.RateTicket)

	// Ticket messages
	tickets.Get("/:id/messages", r.ticketHandler.GetTicketMessages)
	tickets.Post("/:id/messages", r.ticketHandler.AddReply)

	// Ticket history
	tickets.Get("/:id/history", r.ticketHandler.GetTicketHistory)

	// Customer tickets
	group.Get("/customers/:customer_id/tickets", r.ticketHandler.GetCustomerTickets)
}

// setupAgentRoutes sets up agent-specific routes
func (r *Router) setupAgentRoutes(group fiber.Router) {
	tickets := group.Group("/tickets")

	// Agent ticket actions
	tickets.Post("/:id/assign", r.ticketHandler.AssignTicket)
	tickets.Post("/:id/transfer", r.ticketHandler.TransferTicket)

	// Agent tickets
	group.Get("/agents/:agent_id/tickets", r.ticketHandler.GetAgentTickets)
}

// setupAdminRoutes sets up admin routes
func (r *Router) setupAdminRoutes(admin fiber.Router) {
	// Agent management
	agents := admin.Group("/agents")
	agents.Post("", r.adminHandler.CreateAgent)
	agents.Get("", r.adminHandler.ListAgents)
	agents.Get("/:id", r.adminHandler.GetAgent)
	agents.Patch("/:id", r.adminHandler.UpdateAgent)
	agents.Delete("/:id", r.adminHandler.DeleteAgent)
	agents.Patch("/:id/status", r.adminHandler.UpdateAgentStatus)

	// Department management
	departments := admin.Group("/departments")
	departments.Post("", r.adminHandler.CreateDepartment)
	departments.Get("", r.adminHandler.ListDepartments)
	departments.Get("/:id", r.adminHandler.GetDepartment)
	departments.Patch("/:id", r.adminHandler.UpdateDepartment)
	departments.Delete("/:id", r.adminHandler.DeleteDepartment)
	departments.Get("/:id/agents", r.adminHandler.GetDepartmentAgents)
	departments.Post("/:id/agents", r.adminHandler.AddAgentToDepartment)
	departments.Delete("/:id/agents/:agent_id", r.adminHandler.RemoveAgentFromDepartment)

	// Category management
	categories := admin.Group("/categories")
	categories.Post("", r.adminHandler.CreateCategory)
	categories.Get("", r.adminHandler.ListCategories)
	categories.Get("/:id", r.adminHandler.GetCategory)
	categories.Patch("/:id", r.adminHandler.UpdateCategory)
	categories.Delete("/:id", r.adminHandler.DeleteCategory)

	// SLA policy management
	slaPolicies := admin.Group("/sla-policies")
	slaPolicies.Post("", r.adminHandler.CreateSLAPolicy)
	slaPolicies.Get("", r.adminHandler.ListSLAPolicies)
	slaPolicies.Get("/:id", r.adminHandler.GetSLAPolicy)
	slaPolicies.Patch("/:id", r.adminHandler.UpdateSLAPolicy)
	slaPolicies.Delete("/:id", r.adminHandler.DeleteSLAPolicy)

	// Canned response management
	cannedResponses := admin.Group("/canned-responses")
	cannedResponses.Post("", r.adminHandler.CreateCannedResponse)
	cannedResponses.Get("", r.adminHandler.ListCannedResponses)
	cannedResponses.Get("/:id", r.adminHandler.GetCannedResponse)
	cannedResponses.Patch("/:id", r.adminHandler.UpdateCannedResponse)
	cannedResponses.Delete("/:id", r.adminHandler.DeleteCannedResponse)

	// Bulk operations
	bulk := admin.Group("/tickets")
	bulk.Post("/bulk-assign", r.adminHandler.BulkAssignTickets)
	bulk.Post("/bulk-status", r.adminHandler.BulkChangeStatus)
	bulk.Post("/bulk-priority", r.adminHandler.BulkChangePriority)
	bulk.Post("/bulk-transfer", r.adminHandler.BulkTransferDepartment)
	bulk.Post("/bulk-delete", r.adminHandler.BulkDeleteTickets)

	// Dashboard
	dashboard := admin.Group("/dashboard")
	dashboard.Get("/stats", r.adminHandler.GetDashboardStats)
	dashboard.Get("/sla-breached", r.adminHandler.GetSLABreachedTickets)
	dashboard.Get("/due-soon", r.adminHandler.GetTicketsDueSoon)
	dashboard.Get("/unassigned", r.adminHandler.GetUnassignedTickets)
}

// customErrorHandler handles errors
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"message": message,
	})
}

// GetApp returns the fiber app
func (r *Router) GetApp() *fiber.App {
	return r.app
}
