package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/minisource/go-common/i18n"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/ticket/api/router"
	"github.com/minisource/ticket/api/v1/handlers"
	"github.com/minisource/ticket/config"
	_ "github.com/minisource/ticket/docs" // Swagger docs
	"github.com/minisource/ticket/internal/database"
	"github.com/minisource/ticket/internal/repository"
	"github.com/minisource/ticket/internal/usecase"
)

// @title Ticket Service API
// @version 1.0
// @description Customer Support Ticket Management Service for Minisource
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@minisource.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:5011
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @tag.name Tickets
// @tag.description Customer ticket operations
// @tag.name Agent
// @tag.description Agent operations for handling tickets
// @tag.name Admin
// @tag.description Administrative operations for ticket system

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := logging.NewLogger(&logging.LoggerConfig{
		FilePath: "logs/ticket.log",
		Encoding: "json",
		Level:    cfg.Logging.Level,
		Logger:   "zap",
	})

	logger.Info(logging.General, logging.Startup, "Starting ticket service...", nil)

	// Initialize i18n - translator is used globally via i18n.GetTranslator()
	if err := i18n.GetTranslator().LoadTranslations(); err != nil {
		logger.Warn(logging.General, logging.Startup, "Failed to load locales, using defaults", nil)
	}

	// Initialize MongoDB
	db, err := database.NewMongoDB(cfg.MongoDB)
	if err != nil {
		logger.Fatal(logging.General, logging.Startup, "Failed to connect to MongoDB", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
	}
	defer func() {
		if err := db.Close(context.Background()); err != nil {
			logger.Error(logging.General, logging.Startup, "Failed to close MongoDB", map[logging.ExtraKey]interface{}{
				"error": err.Error(),
			})
		}
	}()

	// Create indexes
	if err := db.CreateIndexes(context.Background()); err != nil {
		logger.Error(logging.General, logging.Startup, "Failed to create indexes", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
	}

	logger.Info(logging.General, logging.Startup, "MongoDB connected successfully", nil)

	// Initialize repositories
	ticketRepo := repository.NewTicketRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	historyRepo := repository.NewHistoryRepository(db)
	departmentRepo := repository.NewDepartmentRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	agentRepo := repository.NewAgentRepository(db)
	slaRepo := repository.NewSLAPolicyRepository(db)
	cannedRepo := repository.NewCannedResponseRepository(db)

	// Initialize usecases
	ticketUsecase := usecase.NewTicketUsecase(
		ticketRepo,
		messageRepo,
		historyRepo,
		departmentRepo,
		categoryRepo,
		agentRepo,
		slaRepo,
		cfg,
	)

	departmentUsecase := usecase.NewDepartmentUsecase(
		departmentRepo,
		categoryRepo,
		agentRepo,
		slaRepo,
		cfg,
	)

	categoryUsecase := usecase.NewCategoryUsecase(
		categoryRepo,
		departmentRepo,
	)

	adminUsecase := usecase.NewAdminUsecase(
		ticketRepo,
		messageRepo,
		historyRepo,
		departmentRepo,
		categoryRepo,
		agentRepo,
		slaRepo,
		cannedRepo,
		cfg,
	)

	// Initialize handlers
	ticketHandler := handlers.NewTicketHandler(ticketUsecase)
	adminHandler := handlers.NewAdminHandler(adminUsecase, departmentUsecase, categoryUsecase)
	healthHandler := handlers.NewHealthHandler()

	// Initialize router
	r := router.NewRouter(cfg, logger, ticketHandler, adminHandler, healthHandler)
	app := r.Setup()

	// Start server in goroutine
	go func() {
		addr := fmt.Sprintf(":%d", cfg.Server.Port)
		logger.Info(logging.General, logging.Startup, fmt.Sprintf("Server starting on %s", addr), nil)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info(logging.General, logging.Startup, "Shutting down server...", nil)

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Error(logging.General, logging.Startup, "Server forced to shutdown", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
	}

	logger.Info(logging.General, logging.Startup, "Server exited", nil)
}
