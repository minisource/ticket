package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/minisource/ticket/config"
	"github.com/minisource/ticket/internal/models"
	"github.com/minisource/ticket/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TicketUsecase handles ticket business logic
type TicketUsecase struct {
	ticketRepo     *repository.TicketRepository
	messageRepo    *repository.MessageRepository
	historyRepo    *repository.HistoryRepository
	departmentRepo *repository.DepartmentRepository
	categoryRepo   *repository.CategoryRepository
	agentRepo      *repository.AgentRepository
	slaRepo        *repository.SLAPolicyRepository
	config         *config.Config
}

// NewTicketUsecase creates a new ticket usecase
func NewTicketUsecase(
	ticketRepo *repository.TicketRepository,
	messageRepo *repository.MessageRepository,
	historyRepo *repository.HistoryRepository,
	departmentRepo *repository.DepartmentRepository,
	categoryRepo *repository.CategoryRepository,
	agentRepo *repository.AgentRepository,
	slaRepo *repository.SLAPolicyRepository,
	cfg *config.Config,
) *TicketUsecase {
	return &TicketUsecase{
		ticketRepo:     ticketRepo,
		messageRepo:    messageRepo,
		historyRepo:    historyRepo,
		departmentRepo: departmentRepo,
		categoryRepo:   categoryRepo,
		agentRepo:      agentRepo,
		slaRepo:        slaRepo,
		config:         cfg,
	}
}

// CreateTicket creates a new ticket
func (u *TicketUsecase) CreateTicket(ctx context.Context, req models.CreateTicketRequest, customerID, customerName, customerEmail, ip, userAgent string) (*models.Ticket, error) {
	// Validate required fields
	if req.Subject == "" {
		return nil, errors.New("subject is required")
	}
	if req.Description == "" {
		return nil, errors.New("description is required")
	}

	// Generate ticket number
	ticketNumber, err := u.ticketRepo.GetNextTicketNumber(ctx, req.TenantID)
	if err != nil {
		return nil, err
	}

	// Set defaults
	if req.Priority == "" {
		req.Priority = models.PriorityMedium
	}
	if req.Source == "" {
		req.Source = models.SourceWeb
	}
	if req.Type == "" {
		req.Type = models.TypeQuestion
	}

	ticket := &models.Ticket{
		TenantID:      req.TenantID,
		TicketNumber:  ticketNumber,
		Subject:       req.Subject,
		Description:   req.Description,
		Type:          req.Type,
		Status:        models.StatusOpen,
		Priority:      req.Priority,
		Source:        req.Source,
		CustomerID:    customerID,
		CustomerName:  customerName,
		CustomerEmail: customerEmail,
		Tags:          req.Tags,
		CustomFields:  req.CustomFields,
		CCEmails:      req.CCEmails,
		IPAddress:     ip,
		UserAgent:     userAgent,
		Metadata:      req.Metadata,
		MessageCount:  0,
	}

	// Set department
	if req.DepartmentID != "" {
		deptID, err := primitive.ObjectIDFromHex(req.DepartmentID)
		if err == nil {
			dept, err := u.departmentRepo.GetByID(ctx, deptID)
			if err == nil && dept != nil {
				ticket.DepartmentID = &deptID
				ticket.DepartmentName = dept.Name
			}
		}
	}

	// Set category
	if req.CategoryID != "" {
		catID, err := primitive.ObjectIDFromHex(req.CategoryID)
		if err == nil {
			cat, err := u.categoryRepo.GetByID(ctx, catID)
			if err == nil && cat != nil {
				ticket.CategoryID = &catID
				ticket.CategoryName = cat.Name
			}
		}
	}

	// Process attachments
	for _, att := range req.Attachments {
		ticket.Attachments = append(ticket.Attachments, models.Attachment{
			ID:         uuid.New().String(),
			Name:       att.Name,
			URL:        att.URL,
			Size:       att.Size,
			MimeType:   att.MimeType,
			UploadedBy: customerID,
			UploadedAt: time.Now(),
		})
	}

	// Calculate SLA
	u.calculateSLA(ctx, ticket)

	// Create ticket
	if err := u.ticketRepo.Create(ctx, ticket); err != nil {
		return nil, err
	}

	// Update department stats
	if ticket.DepartmentID != nil {
		_ = u.departmentRepo.IncrementTicketCount(ctx, *ticket.DepartmentID, true)
	}

	// Create history entry
	u.createHistory(ctx, ticket.ID, req.TenantID, "created", "", nil, nil, customerID, customerName, "")

	// Auto-assign if enabled
	if u.config.Ticket.AutoAssignEnabled && ticket.DepartmentID != nil {
		u.autoAssign(ctx, ticket)
	}

	return ticket, nil
}

// GetTicket gets a ticket by ID
func (u *TicketUsecase) GetTicket(ctx context.Context, id string) (*models.Ticket, error) {
	ticketID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid ticket ID")
	}

	ticket, err := u.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if ticket == nil {
		return nil, errors.New("ticket not found")
	}

	return ticket, nil
}

// GetTicketByNumber gets a ticket by ticket number
func (u *TicketUsecase) GetTicketByNumber(ctx context.Context, tenantID, ticketNumber string) (*models.Ticket, error) {
	ticket, err := u.ticketRepo.GetByTicketNumber(ctx, tenantID, ticketNumber)
	if err != nil {
		return nil, err
	}
	if ticket == nil {
		return nil, errors.New("ticket not found")
	}

	return ticket, nil
}

// UpdateTicket updates a ticket
func (u *TicketUsecase) UpdateTicket(ctx context.Context, id string, req models.UpdateTicketRequest, userID, userName string, isAgent bool) (*models.Ticket, error) {
	ticket, err := u.GetTicket(ctx, id)
	if err != nil {
		return nil, err
	}

	// Customers can only update their own tickets
	if !isAgent && ticket.CustomerID != userID {
		return nil, errors.New("you can only update your own tickets")
	}

	// Track changes for history
	changes := make(map[string][2]interface{})

	if req.Subject != nil && *req.Subject != ticket.Subject {
		changes["subject"] = [2]interface{}{ticket.Subject, *req.Subject}
		ticket.Subject = *req.Subject
	}

	if req.Description != nil && *req.Description != ticket.Description {
		changes["description"] = [2]interface{}{ticket.Description, *req.Description}
		ticket.Description = *req.Description
	}

	if req.Type != nil && *req.Type != ticket.Type {
		changes["type"] = [2]interface{}{ticket.Type, *req.Type}
		ticket.Type = *req.Type
	}

	if req.Priority != nil && *req.Priority != ticket.Priority {
		changes["priority"] = [2]interface{}{ticket.Priority, *req.Priority}
		ticket.Priority = *req.Priority
		// Recalculate SLA if priority changed
		u.calculateSLA(ctx, ticket)
	}

	if req.DepartmentID != nil {
		deptID, err := primitive.ObjectIDFromHex(*req.DepartmentID)
		if err == nil {
			dept, err := u.departmentRepo.GetByID(ctx, deptID)
			if err == nil && dept != nil {
				if ticket.DepartmentID == nil || *ticket.DepartmentID != deptID {
					changes["department"] = [2]interface{}{ticket.DepartmentName, dept.Name}
					ticket.DepartmentID = &deptID
					ticket.DepartmentName = dept.Name
				}
			}
		}
	}

	if req.CategoryID != nil {
		catID, err := primitive.ObjectIDFromHex(*req.CategoryID)
		if err == nil {
			cat, err := u.categoryRepo.GetByID(ctx, catID)
			if err == nil && cat != nil {
				changes["category"] = [2]interface{}{ticket.CategoryName, cat.Name}
				ticket.CategoryID = &catID
				ticket.CategoryName = cat.Name
			}
		}
	}

	if req.Tags != nil {
		ticket.Tags = req.Tags
	}

	if req.CCEmails != nil {
		ticket.CCEmails = req.CCEmails
	}

	if req.CustomFields != nil {
		ticket.CustomFields = req.CustomFields
	}

	// Save changes
	if err := u.ticketRepo.Update(ctx, ticket); err != nil {
		return nil, err
	}

	// Create history for each change
	for field, values := range changes {
		u.createHistory(ctx, ticket.ID, ticket.TenantID, "updated", field, values[0], values[1], userID, userName, "")
	}

	return ticket, nil
}

// ChangeStatus changes the ticket status
func (u *TicketUsecase) ChangeStatus(ctx context.Context, id string, req models.ChangeStatusRequest, userID, userName string, isAgent bool) (*models.Ticket, error) {
	ticket, err := u.GetTicket(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate status transition
	if !u.isValidStatusTransition(ticket.Status, req.Status, isAgent) {
		return nil, errors.New("invalid status transition")
	}

	oldStatus := ticket.Status
	ticket.Status = req.Status
	ticket.LastActivityAt = time.Now()

	// Set timestamps based on status
	now := time.Now()
	switch req.Status {
	case models.StatusResolved:
		ticket.ResolvedAt = &now
		if ticket.AssignedToID != "" {
			agentID, _ := primitive.ObjectIDFromHex(ticket.AssignedToID)
			_ = u.agentRepo.IncrementResolved(ctx, agentID)
			_ = u.agentRepo.DecrementTicketCount(ctx, agentID)
		}
		if ticket.DepartmentID != nil {
			_ = u.departmentRepo.DecrementOpenTickets(ctx, *ticket.DepartmentID)
		}
	case models.StatusClosed:
		ticket.ClosedAt = &now
		if oldStatus != models.StatusResolved {
			if ticket.AssignedToID != "" {
				agentID, _ := primitive.ObjectIDFromHex(ticket.AssignedToID)
				_ = u.agentRepo.DecrementTicketCount(ctx, agentID)
			}
			if ticket.DepartmentID != nil {
				_ = u.departmentRepo.DecrementOpenTickets(ctx, *ticket.DepartmentID)
			}
		}
	case models.StatusReopened:
		ticket.ResolvedAt = nil
		ticket.ClosedAt = nil
		ticket.ReopenCount++
		if ticket.DepartmentID != nil {
			_ = u.departmentRepo.IncrementTicketCount(ctx, *ticket.DepartmentID, true)
		}
	}

	if err := u.ticketRepo.Update(ctx, ticket); err != nil {
		return nil, err
	}

	u.createHistory(ctx, ticket.ID, ticket.TenantID, "status_changed", "status", oldStatus, req.Status, userID, userName, req.Comment)

	return ticket, nil
}

// AssignTicket assigns a ticket to an agent
func (u *TicketUsecase) AssignTicket(ctx context.Context, id string, req models.AssignTicketRequest, assignedByID, assignedByName string) (*models.Ticket, error) {
	ticket, err := u.GetTicket(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get agent
	agent, err := u.agentRepo.GetByUserID(ctx, ticket.TenantID, req.AssigneeID)
	if err != nil {
		return nil, err
	}
	if agent == nil {
		return nil, errors.New("agent not found")
	}

	// Check capacity
	if agent.CurrentTickets >= agent.MaxTickets {
		return nil, errors.New("agent has reached maximum capacity")
	}

	// Update old assignee
	if ticket.AssignedToID != "" {
		oldAgentID, _ := primitive.ObjectIDFromHex(ticket.AssignedToID)
		_ = u.agentRepo.DecrementTicketCount(ctx, oldAgentID)
	}

	oldAssignee := ticket.AssignedToName
	now := time.Now()
	ticket.AssignedToID = agent.UserID
	ticket.AssignedToName = agent.Name
	ticket.AssignedToEmail = agent.Email
	ticket.AssignedAt = &now
	ticket.AssignedByID = assignedByID
	ticket.LastActivityAt = now

	// Set first response due if not set
	if ticket.FirstResponseDue == nil {
		u.calculateSLA(ctx, ticket)
	}

	// Update status if open
	if ticket.Status == models.StatusOpen {
		ticket.Status = models.StatusInProgress
	}

	if err := u.ticketRepo.Update(ctx, ticket); err != nil {
		return nil, err
	}

	// Update agent stats
	_ = u.agentRepo.IncrementTicketCount(ctx, agent.ID)

	u.createHistory(ctx, ticket.ID, ticket.TenantID, "assigned", "assignee", oldAssignee, agent.Name, assignedByID, assignedByName, req.Comment)

	return ticket, nil
}

// TransferTicket transfers a ticket to another department
func (u *TicketUsecase) TransferTicket(ctx context.Context, id string, req models.TransferTicketRequest, userID, userName string) (*models.Ticket, error) {
	ticket, err := u.GetTicket(ctx, id)
	if err != nil {
		return nil, err
	}

	deptID, err := primitive.ObjectIDFromHex(req.DepartmentID)
	if err != nil {
		return nil, errors.New("invalid department ID")
	}

	dept, err := u.departmentRepo.GetByID(ctx, deptID)
	if err != nil || dept == nil {
		return nil, errors.New("department not found")
	}

	oldDept := ticket.DepartmentName

	// Update old department stats
	if ticket.DepartmentID != nil {
		_ = u.departmentRepo.DecrementOpenTickets(ctx, *ticket.DepartmentID)
	}

	// Remove old assignee
	if ticket.AssignedToID != "" {
		oldAgentID, _ := primitive.ObjectIDFromHex(ticket.AssignedToID)
		_ = u.agentRepo.DecrementTicketCount(ctx, oldAgentID)
	}

	ticket.DepartmentID = &deptID
	ticket.DepartmentName = dept.Name
	ticket.AssignedToID = ""
	ticket.AssignedToName = ""
	ticket.AssignedToEmail = ""
	ticket.AssignedAt = nil
	ticket.LastActivityAt = time.Now()

	// Update new department stats
	_ = u.departmentRepo.IncrementTicketCount(ctx, deptID, true)

	if err := u.ticketRepo.Update(ctx, ticket); err != nil {
		return nil, err
	}

	u.createHistory(ctx, ticket.ID, ticket.TenantID, "transferred", "department", oldDept, dept.Name, userID, userName, req.Comment)

	// Auto-assign if requested
	if req.AssigneeID != "" {
		_, _ = u.AssignTicket(ctx, id, models.AssignTicketRequest{AssigneeID: req.AssigneeID}, userID, userName)
	} else if u.config.Ticket.AutoAssignEnabled {
		u.autoAssign(ctx, ticket)
	}

	return ticket, nil
}

// AddReply adds a reply to a ticket
func (u *TicketUsecase) AddReply(ctx context.Context, ticketID string, req models.CreateMessageRequest, senderID, senderName, senderEmail string, senderType models.SenderType, ip, userAgent string) (*models.TicketMessage, error) {
	ticket, err := u.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	// Determine message type
	msgType := models.MessageTypeReply
	if req.Type != "" {
		msgType = req.Type
	}
	if req.IsPrivate {
		msgType = models.MessageTypeInternalNote
	}

	message := &models.TicketMessage{
		TicketID:    ticket.ID,
		TenantID:    ticket.TenantID,
		Type:        msgType,
		Content:     req.Content,
		SenderType:  senderType,
		SenderID:    senderID,
		SenderName:  senderName,
		SenderEmail: senderEmail,
		IsPrivate:   req.IsPrivate,
		IPAddress:   ip,
		UserAgent:   userAgent,
	}

	// Process attachments
	for _, att := range req.Attachments {
		message.Attachments = append(message.Attachments, models.Attachment{
			ID:         uuid.New().String(),
			Name:       att.Name,
			URL:        att.URL,
			Size:       att.Size,
			MimeType:   att.MimeType,
			UploadedBy: senderID,
			UploadedAt: time.Now(),
		})
	}

	if err := u.messageRepo.Create(ctx, message); err != nil {
		return nil, err
	}

	// Update ticket
	_ = u.ticketRepo.IncrementMessageCount(ctx, ticket.ID, req.IsPrivate)

	now := time.Now()
	updates := map[string]interface{}{
		"last_activity_at": now,
	}

	if senderType == models.SenderCustomer {
		updates["last_customer_reply_at"] = now
		// If pending, set to open
		if ticket.Status == models.StatusPending {
			updates["status"] = models.StatusOpen
		}
	} else if senderType == models.SenderAgent && !req.IsPrivate {
		updates["last_agent_reply_at"] = now
		// Set first response time
		if ticket.FirstResponsedAt == nil {
			updates["first_responsed_at"] = now
		}
	}

	_ = u.ticketRepo.UpdateFields(ctx, ticket.ID, updates)

	return message, nil
}

// RateTicket adds a satisfaction rating
func (u *TicketUsecase) RateTicket(ctx context.Context, id string, req models.RateTicketRequest, userID string) (*models.Ticket, error) {
	ticket, err := u.GetTicket(ctx, id)
	if err != nil {
		return nil, err
	}

	if ticket.CustomerID != userID {
		return nil, errors.New("you can only rate your own tickets")
	}

	if ticket.Status != models.StatusResolved && ticket.Status != models.StatusClosed {
		return nil, errors.New("can only rate resolved or closed tickets")
	}

	now := time.Now()
	ticket.SatisfactionRating = &req.Rating
	ticket.SatisfactionComment = req.Comment
	ticket.RatedAt = &now

	if err := u.ticketRepo.Update(ctx, ticket); err != nil {
		return nil, err
	}

	// Update agent rating
	if ticket.AssignedToID != "" {
		// Calculate new average rating (simplified - in production, use proper aggregation)
		agentID, _ := primitive.ObjectIDFromHex(ticket.AssignedToID)
		agent, _ := u.agentRepo.GetByID(ctx, agentID)
		if agent != nil {
			newAvg := (agent.AvgRating*float64(agent.TotalResolved-1) + float64(req.Rating)) / float64(agent.TotalResolved)
			agent.AvgRating = newAvg
			_ = u.agentRepo.Update(ctx, agent)
		}
	}

	return ticket, nil
}

// ListTickets lists tickets with filters
func (u *TicketUsecase) ListTickets(ctx context.Context, filter models.TicketFilter) ([]models.Ticket, int64, error) {
	return u.ticketRepo.List(ctx, filter)
}

// GetCustomerTickets gets tickets for a customer
func (u *TicketUsecase) GetCustomerTickets(ctx context.Context, tenantID, customerID string, page, perPage int) ([]models.Ticket, int64, error) {
	return u.ticketRepo.GetByCustomerID(ctx, tenantID, customerID, page, perPage)
}

// GetAgentTickets gets tickets assigned to an agent
func (u *TicketUsecase) GetAgentTickets(ctx context.Context, tenantID, agentID string, page, perPage int) ([]models.Ticket, int64, error) {
	return u.ticketRepo.GetByAssigneeID(ctx, tenantID, agentID, page, perPage)
}

// GetTicketMessages gets messages for a ticket
func (u *TicketUsecase) GetTicketMessages(ctx context.Context, ticketID string, includePrivate bool, page, perPage int) ([]models.TicketMessage, int64, error) {
	id, err := primitive.ObjectIDFromHex(ticketID)
	if err != nil {
		return nil, 0, errors.New("invalid ticket ID")
	}
	return u.messageRepo.GetByTicketID(ctx, id, includePrivate, page, perPage)
}

// GetTicketHistory gets history for a ticket
func (u *TicketUsecase) GetTicketHistory(ctx context.Context, ticketID string, page, perPage int) ([]models.TicketHistory, int64, error) {
	id, err := primitive.ObjectIDFromHex(ticketID)
	if err != nil {
		return nil, 0, errors.New("invalid ticket ID")
	}
	return u.historyRepo.GetByTicketID(ctx, id, page, perPage)
}

// GetStats gets ticket statistics
func (u *TicketUsecase) GetStats(ctx context.Context, tenantID string) (*models.TicketStats, error) {
	return u.ticketRepo.GetStats(ctx, tenantID)
}

// DeleteTicket soft deletes a ticket
func (u *TicketUsecase) DeleteTicket(ctx context.Context, id string, deletedBy string) error {
	ticket, err := u.GetTicket(ctx, id)
	if err != nil {
		return err
	}

	// Update agent stats
	if ticket.AssignedToID != "" {
		agentID, _ := primitive.ObjectIDFromHex(ticket.AssignedToID)
		_ = u.agentRepo.DecrementTicketCount(ctx, agentID)
	}

	// Update department stats
	if ticket.DepartmentID != nil && ticket.Status != models.StatusClosed && ticket.Status != models.StatusResolved {
		_ = u.departmentRepo.DecrementOpenTickets(ctx, *ticket.DepartmentID)
	}

	return u.ticketRepo.Delete(ctx, ticket.ID, deletedBy)
}

// Helper functions

func (u *TicketUsecase) calculateSLA(ctx context.Context, ticket *models.Ticket) {
	if !u.config.SLA.Enabled {
		return
	}

	// Get SLA policy
	var policy *models.SLAPolicy
	var err error

	if ticket.DepartmentID != nil {
		dept, _ := u.departmentRepo.GetByID(ctx, *ticket.DepartmentID)
		if dept != nil && dept.SLAPolicyID != nil {
			policy, err = u.slaRepo.GetByID(ctx, *dept.SLAPolicyID)
		}
	}

	if policy == nil {
		policy, err = u.slaRepo.GetDefault(ctx, ticket.TenantID)
	}

	if err != nil || policy == nil {
		// Use config defaults
		now := time.Now()
		responseDue := now.Add(time.Duration(u.config.SLA.DefaultResponseHours) * time.Hour)
		resolveDue := now.Add(time.Duration(u.config.SLA.DefaultResolveHours) * time.Hour)
		ticket.FirstResponseDue = &responseDue
		ticket.ResolutionDue = &resolveDue
		return
	}

	// Find priority settings
	for _, p := range policy.Priorities {
		if p.Priority == ticket.Priority {
			now := time.Now()
			responseDue := now.Add(time.Duration(p.FirstResponseMins) * time.Minute)
			resolveDue := now.Add(time.Duration(p.ResolutionMins) * time.Minute)
			ticket.FirstResponseDue = &responseDue
			ticket.ResolutionDue = &resolveDue
			ticket.SLAPolicyID = &policy.ID
			break
		}
	}
}

func (u *TicketUsecase) autoAssign(ctx context.Context, ticket *models.Ticket) {
	agents, err := u.agentRepo.GetAvailable(ctx, ticket.TenantID, ticket.DepartmentID)
	if err != nil || len(agents) == 0 {
		return
	}

	// Use least busy agent (already sorted)
	agent := agents[0]
	now := time.Now()
	ticket.AssignedToID = agent.UserID
	ticket.AssignedToName = agent.Name
	ticket.AssignedToEmail = agent.Email
	ticket.AssignedAt = &now

	_ = u.ticketRepo.Update(ctx, ticket)
	_ = u.agentRepo.IncrementTicketCount(ctx, agent.ID)

	u.createHistory(ctx, ticket.ID, ticket.TenantID, "auto_assigned", "assignee", "", agent.Name, "system", "System", "")
}

func (u *TicketUsecase) isValidStatusTransition(from, to models.TicketStatus, isAgent bool) bool {
	// Agents can transition to any status
	if isAgent {
		return true
	}

	// Customer transitions
	switch from {
	case models.StatusResolved:
		return to == models.StatusReopened || to == models.StatusClosed
	case models.StatusPending:
		return to == models.StatusOpen
	case models.StatusOpen:
		return to == models.StatusCancelled
	}

	return false
}

func (u *TicketUsecase) createHistory(ctx context.Context, ticketID primitive.ObjectID, tenantID, action, field string, oldValue, newValue interface{}, changedBy, changedByName, comment string) {
	history := &models.TicketHistory{
		TicketID:      ticketID,
		TenantID:      tenantID,
		Action:        action,
		Field:         field,
		OldValue:      oldValue,
		NewValue:      newValue,
		ChangedBy:     changedBy,
		ChangedByName: changedByName,
		Comment:       comment,
	}
	_ = u.historyRepo.Create(ctx, history)
}
