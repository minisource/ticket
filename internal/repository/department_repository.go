package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/minisource/ticket/internal/database"
	"github.com/minisource/ticket/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DepartmentRepository handles department database operations
type DepartmentRepository struct {
	db *database.MongoDB
}

// NewDepartmentRepository creates a new department repository
func NewDepartmentRepository(db *database.MongoDB) *DepartmentRepository {
	return &DepartmentRepository{db: db}
}

// Create creates a new department
func (r *DepartmentRepository) Create(ctx context.Context, department *models.Department) error {
	department.CreatedAt = time.Now()
	department.UpdatedAt = time.Now()
	department.Slug = r.generateSlug(department.Name)

	result, err := r.db.Collection(database.CollectionDepartments).InsertOne(ctx, department)
	if err != nil {
		return fmt.Errorf("failed to create department: %w", err)
	}

	department.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID gets a department by ID
func (r *DepartmentRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Department, error) {
	var department models.Department
	err := r.db.Collection(database.CollectionDepartments).FindOne(ctx, bson.M{
		"_id":        id,
		"is_deleted": false,
	}).Decode(&department)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get department: %w", err)
	}

	return &department, nil
}

// GetBySlug gets a department by slug
func (r *DepartmentRepository) GetBySlug(ctx context.Context, tenantID, slug string) (*models.Department, error) {
	var department models.Department
	err := r.db.Collection(database.CollectionDepartments).FindOne(ctx, bson.M{
		"tenant_id":  tenantID,
		"slug":       slug,
		"is_deleted": false,
	}).Decode(&department)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get department: %w", err)
	}

	return &department, nil
}

// Update updates a department
func (r *DepartmentRepository) Update(ctx context.Context, department *models.Department) error {
	department.UpdatedAt = time.Now()

	_, err := r.db.Collection(database.CollectionDepartments).UpdateOne(
		ctx,
		bson.M{"_id": department.ID},
		bson.M{"$set": department},
	)
	if err != nil {
		return fmt.Errorf("failed to update department: %w", err)
	}

	return nil
}

// Delete soft deletes a department
func (r *DepartmentRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now()
	_, err := r.db.Collection(database.CollectionDepartments).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"is_deleted": true,
			"deleted_at": now,
			"updated_at": now,
		}},
	)
	if err != nil {
		return fmt.Errorf("failed to delete department: %w", err)
	}

	return nil
}

// List lists departments
func (r *DepartmentRepository) List(ctx context.Context, tenantID string, activeOnly bool) ([]models.Department, error) {
	query := bson.M{
		"tenant_id":  tenantID,
		"is_deleted": false,
	}

	if activeOnly {
		query["is_active"] = true
	}

	opts := options.Find().SetSort(bson.D{{Key: "order", Value: 1}, {Key: "name", Value: 1}})

	cursor, err := r.db.Collection(database.CollectionDepartments).Find(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list departments: %w", err)
	}
	defer cursor.Close(ctx)

	var departments []models.Department
	if err := cursor.All(ctx, &departments); err != nil {
		return nil, fmt.Errorf("failed to decode departments: %w", err)
	}

	return departments, nil
}

// AddAgent adds an agent to a department
func (r *DepartmentRepository) AddAgent(ctx context.Context, departmentID primitive.ObjectID, agentID string) error {
	_, err := r.db.Collection(database.CollectionDepartments).UpdateOne(
		ctx,
		bson.M{"_id": departmentID},
		bson.M{
			"$addToSet": bson.M{"agent_ids": agentID},
			"$set":      bson.M{"updated_at": time.Now()},
		},
	)
	return err
}

// RemoveAgent removes an agent from a department
func (r *DepartmentRepository) RemoveAgent(ctx context.Context, departmentID primitive.ObjectID, agentID string) error {
	_, err := r.db.Collection(database.CollectionDepartments).UpdateOne(
		ctx,
		bson.M{"_id": departmentID},
		bson.M{
			"$pull": bson.M{"agent_ids": agentID},
			"$set":  bson.M{"updated_at": time.Now()},
		},
	)
	return err
}

// IncrementTicketCount increments the ticket count
func (r *DepartmentRepository) IncrementTicketCount(ctx context.Context, departmentID primitive.ObjectID, isOpen bool) error {
	update := bson.M{
		"$inc": bson.M{"total_tickets": 1},
		"$set": bson.M{"updated_at": time.Now()},
	}

	if isOpen {
		update["$inc"].(bson.M)["open_tickets"] = 1
	}

	_, err := r.db.Collection(database.CollectionDepartments).UpdateOne(ctx, bson.M{"_id": departmentID}, update)
	return err
}

// DecrementOpenTickets decrements the open ticket count
func (r *DepartmentRepository) DecrementOpenTickets(ctx context.Context, departmentID primitive.ObjectID) error {
	_, err := r.db.Collection(database.CollectionDepartments).UpdateOne(
		ctx,
		bson.M{"_id": departmentID, "open_tickets": bson.M{"$gt": 0}},
		bson.M{
			"$inc": bson.M{"open_tickets": -1},
			"$set": bson.M{"updated_at": time.Now()},
		},
	)
	return err
}

func (r *DepartmentRepository) generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	return slug
}

// CategoryRepository handles category database operations
type CategoryRepository struct {
	db *database.MongoDB
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(db *database.MongoDB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// Create creates a new category
func (r *CategoryRepository) Create(ctx context.Context, category *models.Category) error {
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()
	category.Slug = r.generateSlug(category.Name)

	result, err := r.db.Collection(database.CollectionCategories).InsertOne(ctx, category)
	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}

	category.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID gets a category by ID
func (r *CategoryRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Category, error) {
	var category models.Category
	err := r.db.Collection(database.CollectionCategories).FindOne(ctx, bson.M{
		"_id":        id,
		"is_deleted": false,
	}).Decode(&category)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return &category, nil
}

// Update updates a category
func (r *CategoryRepository) Update(ctx context.Context, category *models.Category) error {
	category.UpdatedAt = time.Now()

	_, err := r.db.Collection(database.CollectionCategories).UpdateOne(
		ctx,
		bson.M{"_id": category.ID},
		bson.M{"$set": category},
	)
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	return nil
}

// Delete soft deletes a category
func (r *CategoryRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now()
	_, err := r.db.Collection(database.CollectionCategories).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"is_deleted": true,
			"deleted_at": now,
			"updated_at": now,
		}},
	)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}

// List lists categories
func (r *CategoryRepository) List(ctx context.Context, tenantID string, publicOnly bool) ([]models.Category, error) {
	query := bson.M{
		"tenant_id":  tenantID,
		"is_deleted": false,
		"is_active":  true,
	}

	if publicOnly {
		query["is_public"] = true
	}

	opts := options.Find().SetSort(bson.D{{Key: "order", Value: 1}, {Key: "name", Value: 1}})

	cursor, err := r.db.Collection(database.CollectionCategories).Find(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	defer cursor.Close(ctx)

	var categories []models.Category
	if err := cursor.All(ctx, &categories); err != nil {
		return nil, fmt.Errorf("failed to decode categories: %w", err)
	}

	return categories, nil
}

// GetByDepartmentID gets categories for a department
func (r *CategoryRepository) GetByDepartmentID(ctx context.Context, tenantID string, departmentID primitive.ObjectID) ([]models.Category, error) {
	query := bson.M{
		"tenant_id":     tenantID,
		"department_id": departmentID,
		"is_deleted":    false,
		"is_active":     true,
	}

	opts := options.Find().SetSort(bson.D{{Key: "order", Value: 1}, {Key: "name", Value: 1}})

	cursor, err := r.db.Collection(database.CollectionCategories).Find(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	defer cursor.Close(ctx)

	var categories []models.Category
	if err := cursor.All(ctx, &categories); err != nil {
		return nil, fmt.Errorf("failed to decode categories: %w", err)
	}

	return categories, nil
}

func (r *CategoryRepository) generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	return slug
}
