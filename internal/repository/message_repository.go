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

// MessageRepository handles message database operations
type MessageRepository struct {
	db *database.MongoDB
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *database.MongoDB) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create creates a new message
func (r *MessageRepository) Create(ctx context.Context, message *models.TicketMessage) error {
	message.CreatedAt = time.Now()

	result, err := r.db.Collection(database.CollectionMessages).InsertOne(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	message.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID gets a message by ID
func (r *MessageRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.TicketMessage, error) {
	var message models.TicketMessage
	err := r.db.Collection(database.CollectionMessages).FindOne(ctx, bson.M{
		"_id":        id,
		"is_deleted": false,
	}).Decode(&message)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return &message, nil
}

// Update updates a message
func (r *MessageRepository) Update(ctx context.Context, message *models.TicketMessage) error {
	_, err := r.db.Collection(database.CollectionMessages).UpdateOne(
		ctx,
		bson.M{"_id": message.ID},
		bson.M{"$set": message},
	)
	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}

	return nil
}

// Delete soft deletes a message
func (r *MessageRepository) Delete(ctx context.Context, id primitive.ObjectID, deletedBy string) error {
	now := time.Now()
	_, err := r.db.Collection(database.CollectionMessages).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"is_deleted": true,
			"deleted_at": now,
			"deleted_by": deletedBy,
		}},
	)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

// GetByTicketID gets messages for a ticket
func (r *MessageRepository) GetByTicketID(ctx context.Context, ticketID primitive.ObjectID, includePrivate bool, page, perPage int) ([]models.TicketMessage, int64, error) {
	query := bson.M{
		"ticket_id":  ticketID,
		"is_deleted": false,
	}

	if !includePrivate {
		query["is_private"] = false
	}

	// Count total
	total, err := r.db.Collection(database.CollectionMessages).CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count messages: %w", err)
	}

	// Pagination
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 50
	}
	skip := int64((page - 1) * perPage)

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: 1}}).
		SetSkip(skip).
		SetLimit(int64(perPage))

	cursor, err := r.db.Collection(database.CollectionMessages).Find(ctx, query, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list messages: %w", err)
	}
	defer cursor.Close(ctx)

	var messages []models.TicketMessage
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, 0, fmt.Errorf("failed to decode messages: %w", err)
	}

	return messages, total, nil
}

// GetLatestByTicketID gets the latest messages for a ticket
func (r *MessageRepository) GetLatestByTicketID(ctx context.Context, ticketID primitive.ObjectID, limit int) ([]models.TicketMessage, error) {
	query := bson.M{
		"ticket_id":  ticketID,
		"is_deleted": false,
		"is_private": false,
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.db.Collection(database.CollectionMessages).Find(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest messages: %w", err)
	}
	defer cursor.Close(ctx)

	var messages []models.TicketMessage
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, fmt.Errorf("failed to decode messages: %w", err)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// CountByTicketID counts messages for a ticket
func (r *MessageRepository) CountByTicketID(ctx context.Context, ticketID primitive.ObjectID) (int64, error) {
	return r.db.Collection(database.CollectionMessages).CountDocuments(ctx, bson.M{
		"ticket_id":  ticketID,
		"is_deleted": false,
	})
}

// HistoryRepository handles ticket history operations
type HistoryRepository struct {
	db *database.MongoDB
}

// NewHistoryRepository creates a new history repository
func NewHistoryRepository(db *database.MongoDB) *HistoryRepository {
	return &HistoryRepository{db: db}
}

// Create creates a new history entry
func (r *HistoryRepository) Create(ctx context.Context, history *models.TicketHistory) error {
	history.CreatedAt = time.Now()

	result, err := r.db.Collection(database.CollectionTicketHistory).InsertOne(ctx, history)
	if err != nil {
		return fmt.Errorf("failed to create history: %w", err)
	}

	history.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByTicketID gets history for a ticket
func (r *HistoryRepository) GetByTicketID(ctx context.Context, ticketID primitive.ObjectID, page, perPage int) ([]models.TicketHistory, int64, error) {
	query := bson.M{"ticket_id": ticketID}

	total, err := r.db.Collection(database.CollectionTicketHistory).CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count history: %w", err)
	}

	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 50
	}
	skip := int64((page - 1) * perPage)

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(skip).
		SetLimit(int64(perPage))

	cursor, err := r.db.Collection(database.CollectionTicketHistory).Find(ctx, query, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list history: %w", err)
	}
	defer cursor.Close(ctx)

	var history []models.TicketHistory
	if err := cursor.All(ctx, &history); err != nil {
		return nil, 0, fmt.Errorf("failed to decode history: %w", err)
	}

	return history, total, nil
}
