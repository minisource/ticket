package database

import (
	"context"
	"fmt"
	"time"

	"github.com/minisource/ticket/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Collections
const (
	CollectionTickets         = "tickets"
	CollectionMessages        = "messages"
	CollectionDepartments     = "departments"
	CollectionCategories      = "categories"
	CollectionAgents          = "agents"
	CollectionTeams           = "teams"
	CollectionSLAPolicies     = "sla_policies"
	CollectionCannedResponses = "canned_responses"
	CollectionTicketHistory   = "ticket_history"
	CollectionTicketCounters  = "ticket_counters"
)

// MongoDB holds the MongoDB client and database
type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// NewMongoDB creates a new MongoDB connection
func NewMongoDB(cfg config.MongoDBConfig) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Client options
	clientOptions := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize).
		SetMaxConnIdleTime(cfg.MaxConnIdleTime)

	// Connect
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	db := client.Database(cfg.Database)

	mongodb := &MongoDB{
		Client:   client,
		Database: db,
	}

	// Create indexes
	if err := mongodb.CreateIndexes(ctx); err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	return mongodb, nil
}

// Close closes the MongoDB connection
func (m *MongoDB) Close(ctx context.Context) error {
	return m.Client.Disconnect(ctx)
}

// Collection returns a collection by name
func (m *MongoDB) Collection(name string) *mongo.Collection {
	return m.Database.Collection(name)
}

// CreateIndexes creates all necessary indexes
func (m *MongoDB) CreateIndexes(ctx context.Context) error {
	// Ticket indexes
	ticketIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "ticket_number", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "status", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "priority", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "customer_id", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "assigned_to_id", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "department_id", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "category_id", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "last_activity_at", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "sla_breached", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "first_response_due", Value: 1},
			},
			Options: options.Index().SetSparse(true),
		},
		{
			Keys: bson.D{
				{Key: "resolution_due", Value: 1},
			},
			Options: options.Index().SetSparse(true),
		},
		{
			Keys: bson.D{
				{Key: "subject", Value: "text"},
				{Key: "description", Value: "text"},
			},
			Options: options.Index().SetName("text_search"),
		},
		{
			Keys: bson.D{
				{Key: "tags", Value: 1},
			},
		},
	}

	if _, err := m.Collection(CollectionTickets).Indexes().CreateMany(ctx, ticketIndexes); err != nil {
		return fmt.Errorf("failed to create ticket indexes: %w", err)
	}

	// Message indexes
	messageIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "ticket_id", Value: 1},
				{Key: "created_at", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "ticket_id", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "sender_id", Value: 1},
			},
		},
	}

	if _, err := m.Collection(CollectionMessages).Indexes().CreateMany(ctx, messageIndexes); err != nil {
		return fmt.Errorf("failed to create message indexes: %w", err)
	}

	// Department indexes
	departmentIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "slug", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "is_active", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "parent_id", Value: 1},
			},
		},
	}

	if _, err := m.Collection(CollectionDepartments).Indexes().CreateMany(ctx, departmentIndexes); err != nil {
		return fmt.Errorf("failed to create department indexes: %w", err)
	}

	// Category indexes
	categoryIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "slug", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "is_active", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "department_id", Value: 1},
			},
		},
	}

	if _, err := m.Collection(CollectionCategories).Indexes().CreateMany(ctx, categoryIndexes); err != nil {
		return fmt.Errorf("failed to create category indexes: %w", err)
	}

	// Agent indexes
	agentIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "user_id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "email", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "department_ids", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "status", Value: 1},
			},
		},
	}

	if _, err := m.Collection(CollectionAgents).Indexes().CreateMany(ctx, agentIndexes); err != nil {
		return fmt.Errorf("failed to create agent indexes: %w", err)
	}

	// Ticket history indexes
	historyIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "ticket_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "ticket_id", Value: 1},
			},
		},
	}

	if _, err := m.Collection(CollectionTicketHistory).Indexes().CreateMany(ctx, historyIndexes); err != nil {
		return fmt.Errorf("failed to create history indexes: %w", err)
	}

	// SLA policy indexes
	slaIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "is_default", Value: 1},
			},
		},
	}

	if _, err := m.Collection(CollectionSLAPolicies).Indexes().CreateMany(ctx, slaIndexes); err != nil {
		return fmt.Errorf("failed to create SLA indexes: %w", err)
	}

	// Canned response indexes
	cannedIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "is_global", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "shortcut", Value: 1},
			},
			Options: options.Index().SetSparse(true),
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "department_id", Value: 1},
			},
		},
	}

	if _, err := m.Collection(CollectionCannedResponses).Indexes().CreateMany(ctx, cannedIndexes); err != nil {
		return fmt.Errorf("failed to create canned response indexes: %w", err)
	}

	// Team indexes
	teamIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "name", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	}

	if _, err := m.Collection(CollectionTeams).Indexes().CreateMany(ctx, teamIndexes); err != nil {
		return fmt.Errorf("failed to create team indexes: %w", err)
	}

	return nil
}
