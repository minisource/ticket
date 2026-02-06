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

// TicketRepository handles ticket database operations
type TicketRepository struct {
	db *database.MongoDB
}

// NewTicketRepository creates a new ticket repository
func NewTicketRepository(db *database.MongoDB) *TicketRepository {
	return &TicketRepository{db: db}
}

// Create creates a new ticket
func (r *TicketRepository) Create(ctx context.Context, ticket *models.Ticket) error {
	ticket.CreatedAt = time.Now()
	ticket.UpdatedAt = time.Now()
	ticket.LastActivityAt = time.Now()

	result, err := r.db.Collection(database.CollectionTickets).InsertOne(ctx, ticket)
	if err != nil {
		return fmt.Errorf("failed to create ticket: %w", err)
	}

	ticket.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID gets a ticket by ID
func (r *TicketRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Ticket, error) {
	var ticket models.Ticket
	err := r.db.Collection(database.CollectionTickets).FindOne(ctx, bson.M{
		"_id":        id,
		"is_deleted": false,
	}).Decode(&ticket)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}

	return &ticket, nil
}

// GetByTicketNumber gets a ticket by ticket number
func (r *TicketRepository) GetByTicketNumber(ctx context.Context, tenantID, ticketNumber string) (*models.Ticket, error) {
	var ticket models.Ticket
	err := r.db.Collection(database.CollectionTickets).FindOne(ctx, bson.M{
		"tenant_id":     tenantID,
		"ticket_number": ticketNumber,
		"is_deleted":    false,
	}).Decode(&ticket)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}

	return &ticket, nil
}

// Update updates a ticket
func (r *TicketRepository) Update(ctx context.Context, ticket *models.Ticket) error {
	ticket.UpdatedAt = time.Now()

	_, err := r.db.Collection(database.CollectionTickets).UpdateOne(
		ctx,
		bson.M{"_id": ticket.ID},
		bson.M{"$set": ticket},
	)
	if err != nil {
		return fmt.Errorf("failed to update ticket: %w", err)
	}

	return nil
}

// UpdateFields updates specific fields of a ticket
func (r *TicketRepository) UpdateFields(ctx context.Context, id primitive.ObjectID, fields bson.M) error {
	fields["updated_at"] = time.Now()

	_, err := r.db.Collection(database.CollectionTickets).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": fields},
	)
	if err != nil {
		return fmt.Errorf("failed to update ticket fields: %w", err)
	}

	return nil
}

// Delete soft deletes a ticket
func (r *TicketRepository) Delete(ctx context.Context, id primitive.ObjectID, deletedBy string) error {
	now := time.Now()
	_, err := r.db.Collection(database.CollectionTickets).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"is_deleted": true,
			"deleted_at": now,
			"deleted_by": deletedBy,
			"updated_at": now,
		}},
	)
	if err != nil {
		return fmt.Errorf("failed to delete ticket: %w", err)
	}

	return nil
}

// List lists tickets with filters
func (r *TicketRepository) List(ctx context.Context, filter models.TicketFilter) ([]models.Ticket, int64, error) {
	query := bson.M{"is_deleted": false}

	if filter.TenantID != "" {
		query["tenant_id"] = filter.TenantID
	}

	if len(filter.Status) > 0 {
		query["status"] = bson.M{"$in": filter.Status}
	}

	if len(filter.Priority) > 0 {
		query["priority"] = bson.M{"$in": filter.Priority}
	}

	if len(filter.Type) > 0 {
		query["type"] = bson.M{"$in": filter.Type}
	}

	if filter.DepartmentID != "" {
		deptID, _ := primitive.ObjectIDFromHex(filter.DepartmentID)
		query["department_id"] = deptID
	}

	if filter.CategoryID != "" {
		catID, _ := primitive.ObjectIDFromHex(filter.CategoryID)
		query["category_id"] = catID
	}

	if filter.AssignedToID != "" {
		query["assigned_to_id"] = filter.AssignedToID
	}

	if filter.CustomerID != "" {
		query["customer_id"] = filter.CustomerID
	}

	if len(filter.Tags) > 0 {
		query["tags"] = bson.M{"$in": filter.Tags}
	}

	if filter.SLABreached != nil && *filter.SLABreached {
		query["sla_breached"] = true
	}

	if filter.Unassigned != nil && *filter.Unassigned {
		query["assigned_to_id"] = ""
	}

	if filter.Search != "" {
		query["$text"] = bson.M{"$search": filter.Search}
	}

	if filter.CreatedFrom != nil {
		if _, ok := query["created_at"]; !ok {
			query["created_at"] = bson.M{}
		}
		query["created_at"].(bson.M)["$gte"] = *filter.CreatedFrom
	}

	if filter.CreatedTo != nil {
		if _, ok := query["created_at"]; !ok {
			query["created_at"] = bson.M{}
		}
		query["created_at"].(bson.M)["$lte"] = *filter.CreatedTo
	}

	// Count total
	total, err := r.db.Collection(database.CollectionTickets).CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tickets: %w", err)
	}

	// Sort
	sortField := "created_at"
	sortOrder := -1
	if filter.SortBy != "" {
		sortField = filter.SortBy
	}
	if filter.SortOrder == "asc" {
		sortOrder = 1
	}

	// Pagination
	page := 1
	perPage := 20
	if filter.Page > 0 {
		page = filter.Page
	}
	if filter.PerPage > 0 {
		perPage = filter.PerPage
	}
	skip := int64((page - 1) * perPage)

	opts := options.Find().
		SetSort(bson.D{{Key: sortField, Value: sortOrder}}).
		SetSkip(skip).
		SetLimit(int64(perPage))

	cursor, err := r.db.Collection(database.CollectionTickets).Find(ctx, query, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tickets: %w", err)
	}
	defer cursor.Close(ctx)

	var tickets []models.Ticket
	if err := cursor.All(ctx, &tickets); err != nil {
		return nil, 0, fmt.Errorf("failed to decode tickets: %w", err)
	}

	return tickets, total, nil
}

// GetByCustomerID gets tickets for a customer
func (r *TicketRepository) GetByCustomerID(ctx context.Context, tenantID, customerID string, page, perPage int) ([]models.Ticket, int64, error) {
	filter := models.TicketFilter{
		TenantID:   tenantID,
		CustomerID: customerID,
		Page:       page,
		PerPage:    perPage,
		SortBy:     "created_at",
		SortOrder:  "desc",
	}
	return r.List(ctx, filter)
}

// GetByAssigneeID gets tickets assigned to an agent
func (r *TicketRepository) GetByAssigneeID(ctx context.Context, tenantID, assigneeID string, page, perPage int) ([]models.Ticket, int64, error) {
	filter := models.TicketFilter{
		TenantID:     tenantID,
		AssignedToID: assigneeID,
		Page:         page,
		PerPage:      perPage,
		SortBy:       "last_activity_at",
		SortOrder:    "desc",
	}
	return r.List(ctx, filter)
}

// GetByDepartmentID gets tickets for a department
func (r *TicketRepository) GetByDepartmentID(ctx context.Context, tenantID, departmentID string, page, perPage int) ([]models.Ticket, int64, error) {
	filter := models.TicketFilter{
		TenantID:     tenantID,
		DepartmentID: departmentID,
		Page:         page,
		PerPage:      perPage,
		SortBy:       "created_at",
		SortOrder:    "desc",
	}
	return r.List(ctx, filter)
}

// GetUnassigned gets unassigned tickets
func (r *TicketRepository) GetUnassigned(ctx context.Context, tenantID string, departmentID *string, page, perPage int) ([]models.Ticket, int64, error) {
	unassigned := true
	filter := models.TicketFilter{
		TenantID:   tenantID,
		Unassigned: &unassigned,
		Page:       page,
		PerPage:    perPage,
		SortBy:     "created_at",
		SortOrder:  "asc",
	}
	if departmentID != nil {
		filter.DepartmentID = *departmentID
	}
	return r.List(ctx, filter)
}

// GetSLABreached gets tickets with breached SLA
func (r *TicketRepository) GetSLABreached(ctx context.Context, tenantID string, page, perPage int) ([]models.Ticket, int64, error) {
	slaBreached := true
	filter := models.TicketFilter{
		TenantID:    tenantID,
		SLABreached: &slaBreached,
		Page:        page,
		PerPage:     perPage,
		SortBy:      "first_response_due",
		SortOrder:   "asc",
	}
	return r.List(ctx, filter)
}

// GetDueSoon gets tickets with SLA due soon
func (r *TicketRepository) GetDueSoon(ctx context.Context, tenantID string, hours int) ([]models.Ticket, error) {
	now := time.Now()
	deadline := now.Add(time.Duration(hours) * time.Hour)

	query := bson.M{
		"tenant_id":  tenantID,
		"is_deleted": false,
		"status":     bson.M{"$nin": []models.TicketStatus{models.StatusClosed, models.StatusResolved}},
		"$or": []bson.M{
			{
				"first_response_due": bson.M{
					"$gte": now,
					"$lte": deadline,
				},
				"first_responsed_at": nil,
			},
			{
				"resolution_due": bson.M{
					"$gte": now,
					"$lte": deadline,
				},
			},
		},
	}

	cursor, err := r.db.Collection(database.CollectionTickets).Find(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get due soon tickets: %w", err)
	}
	defer cursor.Close(ctx)

	var tickets []models.Ticket
	if err := cursor.All(ctx, &tickets); err != nil {
		return nil, fmt.Errorf("failed to decode tickets: %w", err)
	}

	return tickets, nil
}

// IncrementMessageCount increments the message count
func (r *TicketRepository) IncrementMessageCount(ctx context.Context, id primitive.ObjectID, isInternal bool) error {
	update := bson.M{
		"$inc": bson.M{"message_count": 1},
		"$set": bson.M{
			"updated_at":       time.Now(),
			"last_activity_at": time.Now(),
		},
	}

	if isInternal {
		update["$inc"] = bson.M{"internal_notes": 1}
	}

	_, err := r.db.Collection(database.CollectionTickets).UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

// GetNextTicketNumber gets the next ticket number for a tenant
func (r *TicketRepository) GetNextTicketNumber(ctx context.Context, tenantID string) (string, error) {
	filter := bson.M{"_id": tenantID}
	update := bson.M{"$inc": bson.M{"sequence": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var counter models.TicketCounter
	err := r.db.Collection(database.CollectionTicketCounters).FindOneAndUpdate(ctx, filter, update, opts).Decode(&counter)
	if err != nil {
		return "", fmt.Errorf("failed to get next ticket number: %w", err)
	}

	return fmt.Sprintf("TKT-%06d", counter.Sequence), nil
}

// GetStats gets ticket statistics
func (r *TicketRepository) GetStats(ctx context.Context, tenantID string) (*models.TicketStats, error) {
	stats := &models.TicketStats{
		ByPriority:   make(map[string]int64),
		ByDepartment: make(map[string]int64),
		ByType:       make(map[string]int64),
	}

	baseQuery := bson.M{"tenant_id": tenantID, "is_deleted": false}

	// Total tickets
	total, err := r.db.Collection(database.CollectionTickets).CountDocuments(ctx, baseQuery)
	if err != nil {
		return nil, err
	}
	stats.TotalTickets = total

	// Open tickets
	openQuery := bson.M{"tenant_id": tenantID, "is_deleted": false, "status": models.StatusOpen}
	stats.OpenTickets, _ = r.db.Collection(database.CollectionTickets).CountDocuments(ctx, openQuery)

	// Pending tickets
	pendingQuery := bson.M{"tenant_id": tenantID, "is_deleted": false, "status": models.StatusPending}
	stats.PendingTickets, _ = r.db.Collection(database.CollectionTickets).CountDocuments(ctx, pendingQuery)

	// Resolved tickets
	resolvedQuery := bson.M{"tenant_id": tenantID, "is_deleted": false, "status": models.StatusResolved}
	stats.ResolvedTickets, _ = r.db.Collection(database.CollectionTickets).CountDocuments(ctx, resolvedQuery)

	// Closed tickets
	closedQuery := bson.M{"tenant_id": tenantID, "is_deleted": false, "status": models.StatusClosed}
	stats.ClosedTickets, _ = r.db.Collection(database.CollectionTickets).CountDocuments(ctx, closedQuery)

	// Unassigned tickets
	unassignedQuery := bson.M{"tenant_id": tenantID, "is_deleted": false, "assigned_to_id": ""}
	stats.UnassignedTickets, _ = r.db.Collection(database.CollectionTickets).CountDocuments(ctx, unassignedQuery)

	// SLA breached
	slaQuery := bson.M{"tenant_id": tenantID, "is_deleted": false, "sla_breached": true}
	stats.SLABreached, _ = r.db.Collection(database.CollectionTickets).CountDocuments(ctx, slaQuery)

	return stats, nil
}
