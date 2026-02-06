package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/minisource/ticket/internal/database"
	"github.com/minisource/ticket/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AgentRepository handles agent database operations
type AgentRepository struct {
	db *database.MongoDB
}

// NewAgentRepository creates a new agent repository
func NewAgentRepository(db *database.MongoDB) *AgentRepository {
	return &AgentRepository{db: db}
}

// Create creates a new agent
func (r *AgentRepository) Create(ctx context.Context, agent *models.Agent) error {
	agent.CreatedAt = time.Now()
	agent.UpdatedAt = time.Now()
	agent.IsActive = true

	result, err := r.db.Collection(database.CollectionAgents).InsertOne(ctx, agent)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	agent.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID gets an agent by ID
func (r *AgentRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Agent, error) {
	var agent models.Agent
	err := r.db.Collection(database.CollectionAgents).FindOne(ctx, bson.M{
		"_id":        id,
		"is_deleted": false,
	}).Decode(&agent)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	return &agent, nil
}

// GetByUserID gets an agent by user ID
func (r *AgentRepository) GetByUserID(ctx context.Context, tenantID, userID string) (*models.Agent, error) {
	var agent models.Agent
	err := r.db.Collection(database.CollectionAgents).FindOne(ctx, bson.M{
		"tenant_id":  tenantID,
		"user_id":    userID,
		"is_deleted": false,
	}).Decode(&agent)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	return &agent, nil
}

// GetByEmail gets an agent by email
func (r *AgentRepository) GetByEmail(ctx context.Context, tenantID, email string) (*models.Agent, error) {
	var agent models.Agent
	err := r.db.Collection(database.CollectionAgents).FindOne(ctx, bson.M{
		"tenant_id":  tenantID,
		"email":      email,
		"is_deleted": false,
	}).Decode(&agent)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	return &agent, nil
}

// Update updates an agent
func (r *AgentRepository) Update(ctx context.Context, agent *models.Agent) error {
	agent.UpdatedAt = time.Now()

	_, err := r.db.Collection(database.CollectionAgents).UpdateOne(
		ctx,
		bson.M{"_id": agent.ID},
		bson.M{"$set": agent},
	)
	if err != nil {
		return fmt.Errorf("failed to update agent: %w", err)
	}

	return nil
}

// Delete soft deletes an agent
func (r *AgentRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now()
	_, err := r.db.Collection(database.CollectionAgents).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"is_deleted": true,
			"deleted_at": now,
			"is_active":  false,
			"updated_at": now,
		}},
	)
	if err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}

	return nil
}

// List lists agents
func (r *AgentRepository) List(ctx context.Context, tenantID string, activeOnly bool) ([]models.Agent, error) {
	query := bson.M{
		"tenant_id":  tenantID,
		"is_deleted": false,
	}

	if activeOnly {
		query["is_active"] = true
	}

	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})

	cursor, err := r.db.Collection(database.CollectionAgents).Find(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	defer cursor.Close(ctx)

	var agents []models.Agent
	if err := cursor.All(ctx, &agents); err != nil {
		return nil, fmt.Errorf("failed to decode agents: %w", err)
	}

	return agents, nil
}

// GetByDepartmentID gets agents for a department
func (r *AgentRepository) GetByDepartmentID(ctx context.Context, tenantID string, departmentID primitive.ObjectID) ([]models.Agent, error) {
	query := bson.M{
		"tenant_id":      tenantID,
		"department_ids": departmentID,
		"is_deleted":     false,
		"is_active":      true,
	}

	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})

	cursor, err := r.db.Collection(database.CollectionAgents).Find(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	defer cursor.Close(ctx)

	var agents []models.Agent
	if err := cursor.All(ctx, &agents); err != nil {
		return nil, fmt.Errorf("failed to decode agents: %w", err)
	}

	return agents, nil
}

// GetAvailable gets available agents (for auto-assign)
func (r *AgentRepository) GetAvailable(ctx context.Context, tenantID string, departmentID *primitive.ObjectID) ([]models.Agent, error) {
	query := bson.M{
		"tenant_id":  tenantID,
		"is_deleted": false,
		"is_active":  true,
		"status":     models.AgentStatusAvailable,
		"$expr": bson.M{
			"$lt": bson.A{"$current_tickets", "$max_tickets"},
		},
	}

	if departmentID != nil {
		query["department_ids"] = *departmentID
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "current_tickets", Value: 1}}) // Least busy first

	cursor, err := r.db.Collection(database.CollectionAgents).Find(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list available agents: %w", err)
	}
	defer cursor.Close(ctx)

	var agents []models.Agent
	if err := cursor.All(ctx, &agents); err != nil {
		return nil, fmt.Errorf("failed to decode agents: %w", err)
	}

	return agents, nil
}

// UpdateStatus updates agent status
func (r *AgentRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status models.AgentStatus) error {
	update := bson.M{
		"$set": bson.M{
			"status":         status,
			"is_online":      status != models.AgentStatusOffline,
			"last_active_at": time.Now(),
			"updated_at":     time.Now(),
		},
	}

	_, err := r.db.Collection(database.CollectionAgents).UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

// IncrementTicketCount increments the current ticket count
func (r *AgentRepository) IncrementTicketCount(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.db.Collection(database.CollectionAgents).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$inc": bson.M{"current_tickets": 1, "tickets_today": 1},
			"$set": bson.M{"updated_at": time.Now()},
		},
	)
	return err
}

// DecrementTicketCount decrements the current ticket count
func (r *AgentRepository) DecrementTicketCount(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.db.Collection(database.CollectionAgents).UpdateOne(
		ctx,
		bson.M{"_id": id, "current_tickets": bson.M{"$gt": 0}},
		bson.M{
			"$inc": bson.M{"current_tickets": -1},
			"$set": bson.M{"updated_at": time.Now()},
		},
	)
	return err
}

// IncrementResolved increments the total resolved count
func (r *AgentRepository) IncrementResolved(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.db.Collection(database.CollectionAgents).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$inc": bson.M{"total_resolved": 1},
			"$set": bson.M{"updated_at": time.Now()},
		},
	)
	return err
}

// SLAPolicyRepository handles SLA policy database operations
type SLAPolicyRepository struct {
	db *database.MongoDB
}

// NewSLAPolicyRepository creates a new SLA policy repository
func NewSLAPolicyRepository(db *database.MongoDB) *SLAPolicyRepository {
	return &SLAPolicyRepository{db: db}
}

// Create creates a new SLA policy
func (r *SLAPolicyRepository) Create(ctx context.Context, policy *models.SLAPolicy) error {
	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()
	policy.IsActive = true

	result, err := r.db.Collection(database.CollectionSLAPolicies).InsertOne(ctx, policy)
	if err != nil {
		return fmt.Errorf("failed to create SLA policy: %w", err)
	}

	policy.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID gets an SLA policy by ID
func (r *SLAPolicyRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.SLAPolicy, error) {
	var policy models.SLAPolicy
	err := r.db.Collection(database.CollectionSLAPolicies).FindOne(ctx, bson.M{"_id": id}).Decode(&policy)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get SLA policy: %w", err)
	}

	return &policy, nil
}

// GetDefault gets the default SLA policy for a tenant
func (r *SLAPolicyRepository) GetDefault(ctx context.Context, tenantID string) (*models.SLAPolicy, error) {
	var policy models.SLAPolicy
	err := r.db.Collection(database.CollectionSLAPolicies).FindOne(ctx, bson.M{
		"tenant_id":  tenantID,
		"is_default": true,
		"is_active":  true,
	}).Decode(&policy)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get default SLA policy: %w", err)
	}

	return &policy, nil
}

// Update updates an SLA policy
func (r *SLAPolicyRepository) Update(ctx context.Context, policy *models.SLAPolicy) error {
	policy.UpdatedAt = time.Now()

	_, err := r.db.Collection(database.CollectionSLAPolicies).UpdateOne(
		ctx,
		bson.M{"_id": policy.ID},
		bson.M{"$set": policy},
	)
	if err != nil {
		return fmt.Errorf("failed to update SLA policy: %w", err)
	}

	return nil
}

// Delete deletes an SLA policy
func (r *SLAPolicyRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.db.Collection(database.CollectionSLAPolicies).DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete SLA policy: %w", err)
	}
	return nil
}

// List lists SLA policies
func (r *SLAPolicyRepository) List(ctx context.Context, tenantID string) ([]models.SLAPolicy, error) {
	query := bson.M{"tenant_id": tenantID}

	cursor, err := r.db.Collection(database.CollectionSLAPolicies).Find(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list SLA policies: %w", err)
	}
	defer cursor.Close(ctx)

	var policies []models.SLAPolicy
	if err := cursor.All(ctx, &policies); err != nil {
		return nil, fmt.Errorf("failed to decode SLA policies: %w", err)
	}

	return policies, nil
}

// CannedResponseRepository handles canned response database operations
type CannedResponseRepository struct {
	db *database.MongoDB
}

// NewCannedResponseRepository creates a new canned response repository
func NewCannedResponseRepository(db *database.MongoDB) *CannedResponseRepository {
	return &CannedResponseRepository{db: db}
}

// Create creates a new canned response
func (r *CannedResponseRepository) Create(ctx context.Context, response *models.CannedResponse) error {
	response.CreatedAt = time.Now()
	response.UpdatedAt = time.Now()
	response.IsActive = true

	result, err := r.db.Collection(database.CollectionCannedResponses).InsertOne(ctx, response)
	if err != nil {
		return fmt.Errorf("failed to create canned response: %w", err)
	}

	response.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID gets a canned response by ID
func (r *CannedResponseRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.CannedResponse, error) {
	var response models.CannedResponse
	err := r.db.Collection(database.CollectionCannedResponses).FindOne(ctx, bson.M{"_id": id}).Decode(&response)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get canned response: %w", err)
	}

	return &response, nil
}

// Update updates a canned response
func (r *CannedResponseRepository) Update(ctx context.Context, response *models.CannedResponse) error {
	response.UpdatedAt = time.Now()

	_, err := r.db.Collection(database.CollectionCannedResponses).UpdateOne(
		ctx,
		bson.M{"_id": response.ID},
		bson.M{"$set": response},
	)
	if err != nil {
		return fmt.Errorf("failed to update canned response: %w", err)
	}

	return nil
}

// Delete deletes a canned response
func (r *CannedResponseRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.db.Collection(database.CollectionCannedResponses).DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete canned response: %w", err)
	}
	return nil
}

// List lists canned responses
func (r *CannedResponseRepository) List(ctx context.Context, tenantID string, departmentID *primitive.ObjectID, globalOnly bool) ([]models.CannedResponse, error) {
	query := bson.M{
		"tenant_id": tenantID,
		"is_active": true,
	}

	if globalOnly {
		query["is_global"] = true
	} else if departmentID != nil {
		query["$or"] = []bson.M{
			{"is_global": true},
			{"department_id": *departmentID},
		}
	}

	opts := options.Find().SetSort(bson.D{{Key: "title", Value: 1}})

	cursor, err := r.db.Collection(database.CollectionCannedResponses).Find(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list canned responses: %w", err)
	}
	defer cursor.Close(ctx)

	var responses []models.CannedResponse
	if err := cursor.All(ctx, &responses); err != nil {
		return nil, fmt.Errorf("failed to decode canned responses: %w", err)
	}

	return responses, nil
}

// GetByShortcut gets a canned response by shortcut
func (r *CannedResponseRepository) GetByShortcut(ctx context.Context, tenantID, shortcut string) (*models.CannedResponse, error) {
	var response models.CannedResponse
	err := r.db.Collection(database.CollectionCannedResponses).FindOne(ctx, bson.M{
		"tenant_id": tenantID,
		"shortcut":  shortcut,
		"is_active": true,
	}).Decode(&response)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get canned response: %w", err)
	}

	return &response, nil
}

// IncrementUsage increments the usage count
func (r *CannedResponseRepository) IncrementUsage(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.db.Collection(database.CollectionCannedResponses).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$inc": bson.M{"usage_count": 1}},
	)
	return err
}
