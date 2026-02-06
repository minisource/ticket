package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/minisource/ticket/config"
	"github.com/minisource/ticket/internal/models"
	"github.com/minisource/ticket/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AdminUsecase handles admin business logic
type AdminUsecase struct {
	ticketRepo     *repository.TicketRepository
	messageRepo    *repository.MessageRepository
	historyRepo    *repository.HistoryRepository
	departmentRepo *repository.DepartmentRepository
	categoryRepo   *repository.CategoryRepository
	agentRepo      *repository.AgentRepository
	slaRepo        *repository.SLAPolicyRepository
	cannedRepo     *repository.CannedResponseRepository
	config         *config.Config
}

// NewAdminUsecase creates a new admin usecase
func NewAdminUsecase(
	ticketRepo *repository.TicketRepository,
	messageRepo *repository.MessageRepository,
	historyRepo *repository.HistoryRepository,
	departmentRepo *repository.DepartmentRepository,
	categoryRepo *repository.CategoryRepository,
	agentRepo *repository.AgentRepository,
	slaRepo *repository.SLAPolicyRepository,
	cannedRepo *repository.CannedResponseRepository,
	cfg *config.Config,
) *AdminUsecase {
	return &AdminUsecase{
		ticketRepo:     ticketRepo,
		messageRepo:    messageRepo,
		historyRepo:    historyRepo,
		departmentRepo: departmentRepo,
		categoryRepo:   categoryRepo,
		agentRepo:      agentRepo,
		slaRepo:        slaRepo,
		cannedRepo:     cannedRepo,
		config:         cfg,
	}
}

// ===== Agent Management =====

// CreateAgent creates a new agent
func (u *AdminUsecase) CreateAgent(ctx context.Context, tenantID string, req models.CreateAgentRequest) (*models.Agent, error) {
	if req.UserID == "" || req.Name == "" || req.Email == "" {
		return nil, errors.New("user_id, name, and email are required")
	}

	// Check if agent already exists
	existing, _ := u.agentRepo.GetByUserID(ctx, tenantID, req.UserID)
	if existing != nil {
		return nil, errors.New("agent already exists")
	}

	agent := &models.Agent{
		TenantID:   tenantID,
		UserID:     req.UserID,
		Name:       req.Name,
		Email:      req.Email,
		Phone:      req.Phone,
		Role:       req.Role,
		Skills:     req.Skills,
		Languages:  req.Languages,
		Status:     models.AgentStatusOffline,
		MaxTickets: req.MaxTickets,
		Preferences: models.AgentPreferences{
			EmailNotifications: true,
			PushNotifications:  true,
		},
		IsActive: true,
	}

	// Set defaults
	if agent.Role == "" {
		agent.Role = models.RoleAgent
	}
	if agent.MaxTickets == 0 {
		agent.MaxTickets = 20
	}

	// Set department IDs
	for _, deptID := range req.DepartmentIDs {
		objID, err := primitive.ObjectIDFromHex(deptID)
		if err == nil {
			agent.DepartmentIDs = append(agent.DepartmentIDs, objID)
		}
	}

	if err := u.agentRepo.Create(ctx, agent); err != nil {
		return nil, err
	}

	// Add agent to departments
	for _, deptID := range agent.DepartmentIDs {
		_ = u.departmentRepo.AddAgent(ctx, deptID, agent.UserID)
	}

	return agent, nil
}

// GetAgent gets an agent by ID
func (u *AdminUsecase) GetAgent(ctx context.Context, id string) (*models.Agent, error) {
	agentID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid agent ID")
	}

	agent, err := u.agentRepo.GetByID(ctx, agentID)
	if err != nil {
		return nil, err
	}
	if agent == nil {
		return nil, errors.New("agent not found")
	}

	return agent, nil
}

// GetAgentByUserID gets an agent by user ID
func (u *AdminUsecase) GetAgentByUserID(ctx context.Context, tenantID, userID string) (*models.Agent, error) {
	agent, err := u.agentRepo.GetByUserID(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	if agent == nil {
		return nil, errors.New("agent not found")
	}

	return agent, nil
}

// UpdateAgent updates an agent
func (u *AdminUsecase) UpdateAgent(ctx context.Context, id string, req models.UpdateAgentRequest) (*models.Agent, error) {
	agent, err := u.GetAgent(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		agent.Name = *req.Name
	}
	if req.Phone != nil {
		agent.Phone = *req.Phone
	}
	if req.Role != nil {
		agent.Role = *req.Role
	}
	if req.Skills != nil {
		agent.Skills = req.Skills
	}
	if req.Languages != nil {
		agent.Languages = req.Languages
	}
	if req.IsActive != nil {
		agent.IsActive = *req.IsActive
	}
	if req.MaxTickets != nil {
		agent.MaxTickets = *req.MaxTickets
	}

	// Update department IDs
	if req.DepartmentIDs != nil {
		// Remove from old departments
		for _, deptID := range agent.DepartmentIDs {
			_ = u.departmentRepo.RemoveAgent(ctx, deptID, agent.UserID)
		}

		// Set new departments
		agent.DepartmentIDs = nil
		for _, deptID := range req.DepartmentIDs {
			objID, err := primitive.ObjectIDFromHex(deptID)
			if err == nil {
				agent.DepartmentIDs = append(agent.DepartmentIDs, objID)
				_ = u.departmentRepo.AddAgent(ctx, objID, agent.UserID)
			}
		}
	}

	if err := u.agentRepo.Update(ctx, agent); err != nil {
		return nil, err
	}

	return agent, nil
}

// DeleteAgent deletes an agent
func (u *AdminUsecase) DeleteAgent(ctx context.Context, id string) error {
	agent, err := u.GetAgent(ctx, id)
	if err != nil {
		return err
	}

	// Check if agent has open tickets
	if agent.CurrentTickets > 0 {
		return errors.New("cannot delete agent with open tickets")
	}

	// Remove from departments
	for _, deptID := range agent.DepartmentIDs {
		_ = u.departmentRepo.RemoveAgent(ctx, deptID, agent.UserID)
	}

	return u.agentRepo.Delete(ctx, agent.ID)
}

// ListAgents lists agents
func (u *AdminUsecase) ListAgents(ctx context.Context, tenantID string, activeOnly bool) ([]models.Agent, error) {
	return u.agentRepo.List(ctx, tenantID, activeOnly)
}

// UpdateAgentStatus updates an agent's status
func (u *AdminUsecase) UpdateAgentStatus(ctx context.Context, tenantID, userID string, status models.AgentStatus) error {
	agent, err := u.agentRepo.GetByUserID(ctx, tenantID, userID)
	if err != nil || agent == nil {
		return errors.New("agent not found")
	}

	return u.agentRepo.UpdateStatus(ctx, agent.ID, status)
}

// ===== SLA Policy Management =====

// CreateSLAPolicy creates a new SLA policy
func (u *AdminUsecase) CreateSLAPolicy(ctx context.Context, tenantID string, req models.CreateSLAPolicyRequest) (*models.SLAPolicy, error) {
	if req.Name == "" {
		return nil, errors.New("name is required")
	}

	policy := &models.SLAPolicy{
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		Priorities:  req.Priorities,
		IsActive:    true,
		IsDefault:   req.IsDefault,
	}

	if err := u.slaRepo.Create(ctx, policy); err != nil {
		return nil, err
	}

	return policy, nil
}

// GetSLAPolicy gets an SLA policy by ID
func (u *AdminUsecase) GetSLAPolicy(ctx context.Context, id string) (*models.SLAPolicy, error) {
	slaID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid SLA policy ID")
	}

	policy, err := u.slaRepo.GetByID(ctx, slaID)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, errors.New("SLA policy not found")
	}

	return policy, nil
}

// UpdateSLAPolicy updates an SLA policy
func (u *AdminUsecase) UpdateSLAPolicy(ctx context.Context, id string, req models.UpdateSLAPolicyRequest) (*models.SLAPolicy, error) {
	policy, err := u.GetSLAPolicy(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		policy.Name = *req.Name
	}
	if req.Description != nil {
		policy.Description = *req.Description
	}
	if req.IsActive != nil {
		policy.IsActive = *req.IsActive
	}
	if req.IsDefault != nil {
		policy.IsDefault = *req.IsDefault
	}
	if req.Priorities != nil {
		policy.Priorities = req.Priorities
	}

	if err := u.slaRepo.Update(ctx, policy); err != nil {
		return nil, err
	}

	return policy, nil
}

// DeleteSLAPolicy deletes an SLA policy
func (u *AdminUsecase) DeleteSLAPolicy(ctx context.Context, id string) error {
	policy, err := u.GetSLAPolicy(ctx, id)
	if err != nil {
		return err
	}

	if policy.IsDefault {
		return errors.New("cannot delete default SLA policy")
	}

	return u.slaRepo.Delete(ctx, policy.ID)
}

// ListSLAPolicies lists SLA policies
func (u *AdminUsecase) ListSLAPolicies(ctx context.Context, tenantID string) ([]models.SLAPolicy, error) {
	return u.slaRepo.List(ctx, tenantID)
}

// ===== Canned Response Management =====

// CreateCannedResponse creates a new canned response
func (u *AdminUsecase) CreateCannedResponse(ctx context.Context, tenantID, createdBy string, req models.CreateCannedResponseRequest) (*models.CannedResponse, error) {
	if req.Title == "" || req.Content == "" {
		return nil, errors.New("title and content are required")
	}

	cannedResp := &models.CannedResponse{
		TenantID:  tenantID,
		Title:     req.Title,
		Content:   req.Content,
		Shortcut:  req.Shortcut,
		Category:  req.Category,
		Tags:      req.Tags,
		CreatedBy: createdBy,
		IsActive:  true,
		IsGlobal:  req.IsGlobal,
	}

	// Set department
	if req.DepartmentID != "" {
		deptID, err := primitive.ObjectIDFromHex(req.DepartmentID)
		if err == nil {
			cannedResp.DepartmentID = &deptID
		}
	}

	if err := u.cannedRepo.Create(ctx, cannedResp); err != nil {
		return nil, err
	}

	return cannedResp, nil
}

// GetCannedResponse gets a canned response by ID
func (u *AdminUsecase) GetCannedResponse(ctx context.Context, id string) (*models.CannedResponse, error) {
	cannedID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid canned response ID")
	}

	cannedResp, err := u.cannedRepo.GetByID(ctx, cannedID)
	if err != nil {
		return nil, err
	}
	if cannedResp == nil {
		return nil, errors.New("canned response not found")
	}

	return cannedResp, nil
}

// UpdateCannedResponse updates a canned response
func (u *AdminUsecase) UpdateCannedResponse(ctx context.Context, id string, req models.UpdateCannedResponseRequest) (*models.CannedResponse, error) {
	cannedResp, err := u.GetCannedResponse(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		cannedResp.Title = *req.Title
	}
	if req.Content != nil {
		cannedResp.Content = *req.Content
	}
	if req.Shortcut != nil {
		cannedResp.Shortcut = *req.Shortcut
	}
	if req.Category != nil {
		cannedResp.Category = *req.Category
	}
	if req.Tags != nil {
		cannedResp.Tags = req.Tags
	}
	if req.IsActive != nil {
		cannedResp.IsActive = *req.IsActive
	}
	if req.IsGlobal != nil {
		cannedResp.IsGlobal = *req.IsGlobal
	}

	if err := u.cannedRepo.Update(ctx, cannedResp); err != nil {
		return nil, err
	}

	return cannedResp, nil
}

// DeleteCannedResponse deletes a canned response
func (u *AdminUsecase) DeleteCannedResponse(ctx context.Context, id string) error {
	cannedResp, err := u.GetCannedResponse(ctx, id)
	if err != nil {
		return err
	}

	return u.cannedRepo.Delete(ctx, cannedResp.ID)
}

// ListCannedResponses lists canned responses
func (u *AdminUsecase) ListCannedResponses(ctx context.Context, tenantID string, departmentID *primitive.ObjectID, globalOnly bool) ([]models.CannedResponse, error) {
	return u.cannedRepo.List(ctx, tenantID, departmentID, globalOnly)
}

// ===== Bulk Operations =====

// BulkAssignTickets assigns multiple tickets to an agent
func (u *AdminUsecase) BulkAssignTickets(ctx context.Context, tenantID string, ticketIDs []string, agentID, changedBy, changedByName string) (int, error) {
	agent, err := u.agentRepo.GetByUserID(ctx, tenantID, agentID)
	if err != nil || agent == nil {
		return 0, errors.New("agent not found")
	}

	successCount := 0
	now := time.Now()

	for _, ticketIDStr := range ticketIDs {
		ticketID, err := primitive.ObjectIDFromHex(ticketIDStr)
		if err != nil {
			continue
		}

		ticket, err := u.ticketRepo.GetByID(ctx, ticketID)
		if err != nil || ticket == nil {
			continue
		}

		oldAssignee := ticket.AssignedToID

		// Update ticket
		updates := map[string]interface{}{
			"assigned_to_id":   agent.UserID,
			"assigned_to_name": agent.Name,
			"assigned_at":      now,
			"updated_at":       now,
		}

		if ticket.Status == models.StatusOpen {
			updates["status"] = models.StatusInProgress
		}

		if err := u.ticketRepo.UpdateFields(ctx, ticketID, updates); err != nil {
			continue
		}

		// Create history
		history := &models.TicketHistory{
			TenantID:      ticket.TenantID,
			TicketID:      ticket.ID,
			Action:        "assigned",
			ChangedBy:     changedBy,
			ChangedByName: changedByName,
			OldValue:      oldAssignee,
			NewValue:      agent.UserID,
			CreatedAt:     now,
		}
		_ = u.historyRepo.Create(ctx, history)

		// Update agent stats
		_ = u.agentRepo.IncrementTicketCount(ctx, agent.ID)

		successCount++
	}

	return successCount, nil
}

// BulkChangeStatus changes status of multiple tickets
func (u *AdminUsecase) BulkChangeStatus(ctx context.Context, tenantID string, ticketIDs []string, status models.TicketStatus, changedBy, changedByName string) (int, error) {
	successCount := 0
	now := time.Now()

	for _, ticketIDStr := range ticketIDs {
		ticketID, err := primitive.ObjectIDFromHex(ticketIDStr)
		if err != nil {
			continue
		}

		ticket, err := u.ticketRepo.GetByID(ctx, ticketID)
		if err != nil || ticket == nil {
			continue
		}

		oldStatus := ticket.Status

		// Update ticket
		updates := map[string]interface{}{
			"status":     status,
			"updated_at": now,
		}

		if status == models.StatusResolved || status == models.StatusClosed {
			updates["resolved_at"] = now
		}

		if err := u.ticketRepo.UpdateFields(ctx, ticketID, updates); err != nil {
			continue
		}

		// Create history
		history := &models.TicketHistory{
			TenantID:      ticket.TenantID,
			TicketID:      ticket.ID,
			Action:        "status_changed",
			ChangedBy:     changedBy,
			ChangedByName: changedByName,
			OldValue:      string(oldStatus),
			NewValue:      string(status),
			CreatedAt:     now,
		}
		_ = u.historyRepo.Create(ctx, history)

		successCount++
	}

	return successCount, nil
}

// BulkChangePriority changes priority of multiple tickets
func (u *AdminUsecase) BulkChangePriority(ctx context.Context, tenantID string, ticketIDs []string, priority models.TicketPriority, changedBy, changedByName string) (int, error) {
	successCount := 0
	now := time.Now()

	for _, ticketIDStr := range ticketIDs {
		ticketID, err := primitive.ObjectIDFromHex(ticketIDStr)
		if err != nil {
			continue
		}

		ticket, err := u.ticketRepo.GetByID(ctx, ticketID)
		if err != nil || ticket == nil {
			continue
		}

		oldPriority := ticket.Priority

		// Update ticket
		updates := map[string]interface{}{
			"priority":   priority,
			"updated_at": now,
		}

		if err := u.ticketRepo.UpdateFields(ctx, ticketID, updates); err != nil {
			continue
		}

		// Create history
		history := &models.TicketHistory{
			TenantID:      ticket.TenantID,
			TicketID:      ticket.ID,
			Action:        "priority_changed",
			ChangedBy:     changedBy,
			ChangedByName: changedByName,
			OldValue:      string(oldPriority),
			NewValue:      string(priority),
			CreatedAt:     now,
		}
		_ = u.historyRepo.Create(ctx, history)

		successCount++
	}

	return successCount, nil
}

// BulkTransferDepartment transfers multiple tickets to a department
func (u *AdminUsecase) BulkTransferDepartment(ctx context.Context, tenantID string, ticketIDs []string, departmentID, changedBy, changedByName string) (int, error) {
	deptID, err := primitive.ObjectIDFromHex(departmentID)
	if err != nil {
		return 0, errors.New("invalid department ID")
	}

	dept, err := u.departmentRepo.GetByID(ctx, deptID)
	if err != nil || dept == nil {
		return 0, errors.New("department not found")
	}

	successCount := 0
	now := time.Now()

	for _, ticketIDStr := range ticketIDs {
		ticketID, err := primitive.ObjectIDFromHex(ticketIDStr)
		if err != nil {
			continue
		}

		ticket, err := u.ticketRepo.GetByID(ctx, ticketID)
		if err != nil || ticket == nil {
			continue
		}

		oldDeptID := ""
		if ticket.DepartmentID != nil {
			oldDeptID = ticket.DepartmentID.Hex()
		}

		// Update ticket
		updates := map[string]interface{}{
			"department_id":   deptID,
			"department_name": dept.Name,
			"updated_at":      now,
		}

		if err := u.ticketRepo.UpdateFields(ctx, ticketID, updates); err != nil {
			continue
		}

		// Update department counts
		if ticket.DepartmentID != nil {
			_ = u.departmentRepo.DecrementOpenTickets(ctx, *ticket.DepartmentID)
		}
		_ = u.departmentRepo.IncrementTicketCount(ctx, deptID, true)

		// Create history
		history := &models.TicketHistory{
			TenantID:      ticket.TenantID,
			TicketID:      ticket.ID,
			Action:        "transferred",
			ChangedBy:     changedBy,
			ChangedByName: changedByName,
			OldValue:      oldDeptID,
			NewValue:      deptID.Hex(),
			CreatedAt:     now,
		}
		_ = u.historyRepo.Create(ctx, history)

		successCount++
	}

	return successCount, nil
}

// BulkDeleteTickets deletes multiple tickets
func (u *AdminUsecase) BulkDeleteTickets(ctx context.Context, tenantID string, ticketIDs []string) (int, error) {
	successCount := 0

	for _, ticketIDStr := range ticketIDs {
		ticketID, err := primitive.ObjectIDFromHex(ticketIDStr)
		if err != nil {
			continue
		}

		ticket, err := u.ticketRepo.GetByID(ctx, ticketID)
		if err != nil || ticket == nil {
			continue
		}

		// Verify tenant
		if ticket.TenantID != tenantID {
			continue
		}

		if err := u.ticketRepo.Delete(ctx, ticketID, ""); err != nil {
			continue
		}

		// Update department counts
		if ticket.DepartmentID != nil {
			_ = u.departmentRepo.DecrementOpenTickets(ctx, *ticket.DepartmentID)
		}

		// Update agent stats
		if ticket.AssignedToID != "" {
			agent, _ := u.agentRepo.GetByUserID(ctx, tenantID, ticket.AssignedToID)
			if agent != nil {
				_ = u.agentRepo.DecrementTicketCount(ctx, agent.ID)
			}
		}

		successCount++
	}

	return successCount, nil
}

// ===== Dashboard & Statistics =====

// GetDashboardStats gets dashboard statistics
func (u *AdminUsecase) GetDashboardStats(ctx context.Context, tenantID string) (*models.TicketStats, error) {
	return u.ticketRepo.GetStats(ctx, tenantID)
}

// GetAgentStats gets agent statistics
func (u *AdminUsecase) GetAgentStats(ctx context.Context, tenantID, agentID string) (*models.Agent, error) {
	return u.agentRepo.GetByUserID(ctx, tenantID, agentID)
}

// GetDepartmentStats gets department statistics
func (u *AdminUsecase) GetDepartmentStats(ctx context.Context, id string) (*models.Department, error) {
	deptID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid department ID")
	}

	return u.departmentRepo.GetByID(ctx, deptID)
}

// GetSLABreachedTickets gets tickets with breached SLA
func (u *AdminUsecase) GetSLABreachedTickets(ctx context.Context, tenantID string, page, perPage int) ([]models.Ticket, int64, error) {
	return u.ticketRepo.GetSLABreached(ctx, tenantID, page, perPage)
}

// GetTicketsDueSoon gets tickets due soon
func (u *AdminUsecase) GetTicketsDueSoon(ctx context.Context, tenantID string, hours int) ([]models.Ticket, error) {
	return u.ticketRepo.GetDueSoon(ctx, tenantID, hours)
}

// GetUnassignedTickets gets unassigned tickets
func (u *AdminUsecase) GetUnassignedTickets(ctx context.Context, tenantID string, page, perPage int) ([]models.Ticket, int64, error) {
	return u.ticketRepo.GetUnassigned(ctx, tenantID, nil, page, perPage)
}
