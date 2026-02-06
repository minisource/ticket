package usecase

import (
	"context"
	"errors"

	"github.com/minisource/ticket/config"
	"github.com/minisource/ticket/internal/models"
	"github.com/minisource/ticket/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DepartmentUsecase handles department business logic
type DepartmentUsecase struct {
	departmentRepo *repository.DepartmentRepository
	categoryRepo   *repository.CategoryRepository
	agentRepo      *repository.AgentRepository
	slaRepo        *repository.SLAPolicyRepository
	config         *config.Config
}

// NewDepartmentUsecase creates a new department usecase
func NewDepartmentUsecase(
	departmentRepo *repository.DepartmentRepository,
	categoryRepo *repository.CategoryRepository,
	agentRepo *repository.AgentRepository,
	slaRepo *repository.SLAPolicyRepository,
	cfg *config.Config,
) *DepartmentUsecase {
	return &DepartmentUsecase{
		departmentRepo: departmentRepo,
		categoryRepo:   categoryRepo,
		agentRepo:      agentRepo,
		slaRepo:        slaRepo,
		config:         cfg,
	}
}

// CreateDepartment creates a new department
func (u *DepartmentUsecase) CreateDepartment(ctx context.Context, tenantID string, req models.CreateDepartmentRequest) (*models.Department, error) {
	if req.Name == "" {
		return nil, errors.New("name is required")
	}

	department := &models.Department{
		TenantID:        tenantID,
		Name:            req.Name,
		Description:     req.Description,
		Email:           req.Email,
		AutoAssign:      req.AutoAssign,
		AutoAssignType:  req.AutoAssignType,
		DefaultPriority: req.DefaultPriority,
		Color:           req.Color,
		Icon:            req.Icon,
		Order:           req.Order,
		IsActive:        true,
	}

	// Set parent
	if req.ParentID != "" {
		parentID, err := primitive.ObjectIDFromHex(req.ParentID)
		if err == nil {
			parent, err := u.departmentRepo.GetByID(ctx, parentID)
			if err == nil && parent != nil {
				department.ParentID = &parentID
				department.Path = parent.Path + "/" + department.Slug
				department.Level = parent.Level + 1
			}
		}
	}

	// Set manager
	if req.ManagerID != "" {
		agent, err := u.agentRepo.GetByUserID(ctx, tenantID, req.ManagerID)
		if err == nil && agent != nil {
			department.ManagerID = agent.UserID
			department.ManagerName = agent.Name
		}
	}

	// Set SLA policy
	if req.SLAPolicyID != "" {
		slaID, err := primitive.ObjectIDFromHex(req.SLAPolicyID)
		if err == nil {
			policy, err := u.slaRepo.GetByID(ctx, slaID)
			if err == nil && policy != nil {
				department.SLAPolicyID = &slaID
			}
		}
	}

	// Set default priority if not specified
	if department.DefaultPriority == "" {
		department.DefaultPriority = models.PriorityMedium
	}

	if err := u.departmentRepo.Create(ctx, department); err != nil {
		return nil, err
	}

	return department, nil
}

// GetDepartment gets a department by ID
func (u *DepartmentUsecase) GetDepartment(ctx context.Context, id string) (*models.Department, error) {
	deptID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid department ID")
	}

	department, err := u.departmentRepo.GetByID(ctx, deptID)
	if err != nil {
		return nil, err
	}
	if department == nil {
		return nil, errors.New("department not found")
	}

	return department, nil
}

// UpdateDepartment updates a department
func (u *DepartmentUsecase) UpdateDepartment(ctx context.Context, id string, req models.UpdateDepartmentRequest) (*models.Department, error) {
	department, err := u.GetDepartment(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		department.Name = *req.Name
	}
	if req.Description != nil {
		department.Description = *req.Description
	}
	if req.Email != nil {
		department.Email = *req.Email
	}
	if req.AutoAssign != nil {
		department.AutoAssign = *req.AutoAssign
	}
	if req.AutoAssignType != nil {
		department.AutoAssignType = *req.AutoAssignType
	}
	if req.DefaultPriority != nil {
		department.DefaultPriority = *req.DefaultPriority
	}
	if req.Color != nil {
		department.Color = *req.Color
	}
	if req.Icon != nil {
		department.Icon = *req.Icon
	}
	if req.Order != nil {
		department.Order = *req.Order
	}
	if req.IsActive != nil {
		department.IsActive = *req.IsActive
	}

	// Update manager
	if req.ManagerID != nil {
		if *req.ManagerID == "" {
			department.ManagerID = ""
			department.ManagerName = ""
		} else {
			agent, err := u.agentRepo.GetByUserID(ctx, department.TenantID, *req.ManagerID)
			if err == nil && agent != nil {
				department.ManagerID = agent.UserID
				department.ManagerName = agent.Name
			}
		}
	}

	// Update SLA policy
	if req.SLAPolicyID != nil {
		if *req.SLAPolicyID == "" {
			department.SLAPolicyID = nil
		} else {
			slaID, err := primitive.ObjectIDFromHex(*req.SLAPolicyID)
			if err == nil {
				policy, err := u.slaRepo.GetByID(ctx, slaID)
				if err == nil && policy != nil {
					department.SLAPolicyID = &slaID
				}
			}
		}
	}

	if err := u.departmentRepo.Update(ctx, department); err != nil {
		return nil, err
	}

	return department, nil
}

// DeleteDepartment deletes a department
func (u *DepartmentUsecase) DeleteDepartment(ctx context.Context, id string) error {
	department, err := u.GetDepartment(ctx, id)
	if err != nil {
		return err
	}

	if department.OpenTickets > 0 {
		return errors.New("cannot delete department with open tickets")
	}

	return u.departmentRepo.Delete(ctx, department.ID)
}

// ListDepartments lists departments
func (u *DepartmentUsecase) ListDepartments(ctx context.Context, tenantID string, activeOnly bool) ([]models.Department, error) {
	return u.departmentRepo.List(ctx, tenantID, activeOnly)
}

// AddAgentToDepartment adds an agent to a department
func (u *DepartmentUsecase) AddAgentToDepartment(ctx context.Context, departmentID, agentID string) error {
	deptID, err := primitive.ObjectIDFromHex(departmentID)
	if err != nil {
		return errors.New("invalid department ID")
	}

	dept, err := u.departmentRepo.GetByID(ctx, deptID)
	if err != nil || dept == nil {
		return errors.New("department not found")
	}

	// Update agent's department list
	agent, err := u.agentRepo.GetByUserID(ctx, dept.TenantID, agentID)
	if err != nil || agent == nil {
		return errors.New("agent not found")
	}

	// Add department to agent
	found := false
	for _, d := range agent.DepartmentIDs {
		if d == deptID {
			found = true
			break
		}
	}
	if !found {
		agent.DepartmentIDs = append(agent.DepartmentIDs, deptID)
		if err := u.agentRepo.Update(ctx, agent); err != nil {
			return err
		}
	}

	return u.departmentRepo.AddAgent(ctx, deptID, agentID)
}

// RemoveAgentFromDepartment removes an agent from a department
func (u *DepartmentUsecase) RemoveAgentFromDepartment(ctx context.Context, departmentID, agentID string) error {
	deptID, err := primitive.ObjectIDFromHex(departmentID)
	if err != nil {
		return errors.New("invalid department ID")
	}

	dept, err := u.departmentRepo.GetByID(ctx, deptID)
	if err != nil || dept == nil {
		return errors.New("department not found")
	}

	// Update agent's department list
	agent, err := u.agentRepo.GetByUserID(ctx, dept.TenantID, agentID)
	if err != nil || agent == nil {
		return errors.New("agent not found")
	}

	// Remove department from agent
	newDepts := make([]primitive.ObjectID, 0)
	for _, d := range agent.DepartmentIDs {
		if d != deptID {
			newDepts = append(newDepts, d)
		}
	}
	agent.DepartmentIDs = newDepts
	if err := u.agentRepo.Update(ctx, agent); err != nil {
		return err
	}

	return u.departmentRepo.RemoveAgent(ctx, deptID, agentID)
}

// GetDepartmentAgents gets agents for a department
func (u *DepartmentUsecase) GetDepartmentAgents(ctx context.Context, tenantID, departmentID string) ([]models.Agent, error) {
	deptID, err := primitive.ObjectIDFromHex(departmentID)
	if err != nil {
		return nil, errors.New("invalid department ID")
	}

	return u.agentRepo.GetByDepartmentID(ctx, tenantID, deptID)
}

// CategoryUsecase handles category business logic
type CategoryUsecase struct {
	categoryRepo   *repository.CategoryRepository
	departmentRepo *repository.DepartmentRepository
}

// NewCategoryUsecase creates a new category usecase
func NewCategoryUsecase(
	categoryRepo *repository.CategoryRepository,
	departmentRepo *repository.DepartmentRepository,
) *CategoryUsecase {
	return &CategoryUsecase{
		categoryRepo:   categoryRepo,
		departmentRepo: departmentRepo,
	}
}

// CreateCategory creates a new category
func (u *CategoryUsecase) CreateCategory(ctx context.Context, tenantID string, req models.CreateCategoryRequest) (*models.Category, error) {
	if req.Name == "" {
		return nil, errors.New("name is required")
	}

	category := &models.Category{
		TenantID:        tenantID,
		Name:            req.Name,
		Description:     req.Description,
		DefaultPriority: req.DefaultPriority,
		Icon:            req.Icon,
		Color:           req.Color,
		Order:           req.Order,
		IsActive:        true,
		IsPublic:        req.IsPublic,
	}

	// Set parent
	if req.ParentID != "" {
		parentID, err := primitive.ObjectIDFromHex(req.ParentID)
		if err == nil {
			parent, err := u.categoryRepo.GetByID(ctx, parentID)
			if err == nil && parent != nil {
				category.ParentID = &parentID
				category.Path = parent.Path + "/" + category.Slug
				category.Level = parent.Level + 1
			}
		}
	}

	// Set department
	if req.DepartmentID != "" {
		deptID, err := primitive.ObjectIDFromHex(req.DepartmentID)
		if err == nil {
			dept, err := u.departmentRepo.GetByID(ctx, deptID)
			if err == nil && dept != nil {
				category.DepartmentID = &deptID
			}
		}
	}

	if err := u.categoryRepo.Create(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

// GetCategory gets a category by ID
func (u *CategoryUsecase) GetCategory(ctx context.Context, id string) (*models.Category, error) {
	catID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid category ID")
	}

	category, err := u.categoryRepo.GetByID(ctx, catID)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, errors.New("category not found")
	}

	return category, nil
}

// UpdateCategory updates a category
func (u *CategoryUsecase) UpdateCategory(ctx context.Context, id string, req models.UpdateCategoryRequest) (*models.Category, error) {
	category, err := u.GetCategory(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		category.Name = *req.Name
	}
	if req.Description != nil {
		category.Description = *req.Description
	}
	if req.DefaultPriority != nil {
		category.DefaultPriority = *req.DefaultPriority
	}
	if req.Icon != nil {
		category.Icon = *req.Icon
	}
	if req.Color != nil {
		category.Color = *req.Color
	}
	if req.Order != nil {
		category.Order = *req.Order
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}
	if req.IsPublic != nil {
		category.IsPublic = *req.IsPublic
	}

	if err := u.categoryRepo.Update(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

// DeleteCategory deletes a category
func (u *CategoryUsecase) DeleteCategory(ctx context.Context, id string) error {
	category, err := u.GetCategory(ctx, id)
	if err != nil {
		return err
	}

	return u.categoryRepo.Delete(ctx, category.ID)
}

// ListCategories lists categories
func (u *CategoryUsecase) ListCategories(ctx context.Context, tenantID string, publicOnly bool) ([]models.Category, error) {
	return u.categoryRepo.List(ctx, tenantID, publicOnly)
}

// GetCategoriesByDepartment gets categories for a department
func (u *CategoryUsecase) GetCategoriesByDepartment(ctx context.Context, tenantID, departmentID string) ([]models.Category, error) {
	deptID, err := primitive.ObjectIDFromHex(departmentID)
	if err != nil {
		return nil, errors.New("invalid department ID")
	}

	return u.categoryRepo.GetByDepartmentID(ctx, tenantID, deptID)
}
