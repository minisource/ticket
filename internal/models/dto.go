package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ========================
// Ticket DTOs
// ========================

// CreateTicketRequest represents a request to create a ticket
type CreateTicketRequest struct {
	TenantID     string                 `json:"tenantId,omitempty"`
	Subject      string                 `json:"subject" validate:"required,min=5,max=200"`
	Description  string                 `json:"description" validate:"required,min=10,max=10000"`
	Type         TicketType             `json:"type" validate:"required"`
	Priority     TicketPriority         `json:"priority,omitempty"`
	Source       TicketSource           `json:"source,omitempty"`
	DepartmentID string                 `json:"departmentId,omitempty"`
	CategoryID   string                 `json:"categoryId,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	CustomFields map[string]interface{} `json:"customFields,omitempty"`
	Attachments  []AttachmentInput      `json:"attachments,omitempty"`
	CCEmails     []string               `json:"ccEmails,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateTicketRequest represents a request to update a ticket
type UpdateTicketRequest struct {
	Subject      *string                `json:"subject,omitempty" validate:"omitempty,min=5,max=200"`
	Description  *string                `json:"description,omitempty" validate:"omitempty,min=10,max=10000"`
	Type         *TicketType            `json:"type,omitempty"`
	Priority     *TicketPriority        `json:"priority,omitempty"`
	DepartmentID *string                `json:"departmentId,omitempty"`
	CategoryID   *string                `json:"categoryId,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	CustomFields map[string]interface{} `json:"customFields,omitempty"`
	CCEmails     []string               `json:"ccEmails,omitempty"`
}

// ChangeStatusRequest represents a request to change ticket status
type ChangeStatusRequest struct {
	Status  TicketStatus `json:"status" validate:"required"`
	Comment string       `json:"comment,omitempty"`
}

// AssignTicketRequest represents a request to assign a ticket
type AssignTicketRequest struct {
	AssigneeID string `json:"assigneeId" validate:"required"`
	Comment    string `json:"comment,omitempty"`
}

// TransferTicketRequest represents a request to transfer a ticket
type TransferTicketRequest struct {
	DepartmentID string `json:"departmentId" validate:"required"`
	AssigneeID   string `json:"assigneeId,omitempty"`
	Comment      string `json:"comment,omitempty"`
}

// MergeTicketsRequest represents a request to merge tickets
type MergeTicketsRequest struct {
	SourceTicketIDs []string `json:"sourceTicketIds" validate:"required,min=1"`
	TargetTicketID  string   `json:"targetTicketId" validate:"required"`
	Comment         string   `json:"comment,omitempty"`
}

// LinkTicketsRequest represents a request to link tickets
type LinkTicketsRequest struct {
	RelatedTicketID string `json:"relatedTicketId" validate:"required"`
	LinkType        string `json:"linkType,omitempty"` // related, duplicate, parent, child
}

// RateTicketRequest represents a request to rate a ticket
type RateTicketRequest struct {
	Rating  int    `json:"rating" validate:"required,min=1,max=5"`
	Comment string `json:"comment,omitempty"`
}

// AddWatcherRequest represents a request to add a watcher
type AddWatcherRequest struct {
	UserID string `json:"userId" validate:"required"`
}

// BulkActionRequest represents a request for bulk operations
type BulkActionRequest struct {
	TicketIDs []string               `json:"ticketIds" validate:"required,min=1"`
	Action    string                 `json:"action" validate:"required"` // close, assign, change_priority, change_status, add_tag
	Data      map[string]interface{} `json:"data,omitempty"`
}

// AttachmentInput represents an attachment to upload
type AttachmentInput struct {
	Name     string `json:"name" validate:"required"`
	URL      string `json:"url" validate:"required,url"`
	Size     int64  `json:"size"`
	MimeType string `json:"mimeType"`
}

// ========================
// Message DTOs
// ========================

// CreateMessageRequest represents a request to create a message
type CreateMessageRequest struct {
	Content     string            `json:"content" validate:"required,min=1,max=10000"`
	Type        MessageType       `json:"type,omitempty"`
	IsPrivate   bool              `json:"isPrivate,omitempty"`
	Attachments []AttachmentInput `json:"attachments,omitempty"`
}

// UpdateMessageRequest represents a request to update a message
type UpdateMessageRequest struct {
	Content string `json:"content" validate:"required,min=1,max=10000"`
}

// ========================
// Department DTOs
// ========================

// CreateDepartmentRequest represents a request to create a department
type CreateDepartmentRequest struct {
	Name            string         `json:"name" validate:"required,min=2,max=100"`
	Description     string         `json:"description,omitempty" validate:"max=500"`
	Email           string         `json:"email,omitempty" validate:"omitempty,email"`
	ParentID        string         `json:"parentId,omitempty"`
	ManagerID       string         `json:"managerId,omitempty"`
	AutoAssign      bool           `json:"autoAssign,omitempty"`
	AutoAssignType  string         `json:"autoAssignType,omitempty"`
	DefaultPriority TicketPriority `json:"defaultPriority,omitempty"`
	SLAPolicyID     string         `json:"slaPolicyId,omitempty"`
	Color           string         `json:"color,omitempty"`
	Icon            string         `json:"icon,omitempty"`
	Order           int            `json:"order,omitempty"`
}

// UpdateDepartmentRequest represents a request to update a department
type UpdateDepartmentRequest struct {
	Name            *string         `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Description     *string         `json:"description,omitempty" validate:"omitempty,max=500"`
	Email           *string         `json:"email,omitempty" validate:"omitempty,email"`
	ManagerID       *string         `json:"managerId,omitempty"`
	AutoAssign      *bool           `json:"autoAssign,omitempty"`
	AutoAssignType  *string         `json:"autoAssignType,omitempty"`
	DefaultPriority *TicketPriority `json:"defaultPriority,omitempty"`
	SLAPolicyID     *string         `json:"slaPolicyId,omitempty"`
	Color           *string         `json:"color,omitempty"`
	Icon            *string         `json:"icon,omitempty"`
	Order           *int            `json:"order,omitempty"`
	IsActive        *bool           `json:"isActive,omitempty"`
}

// AddAgentToDepartmentRequest represents a request to add agent to department
type AddAgentToDepartmentRequest struct {
	AgentID string `json:"agentId" validate:"required"`
}

// ========================
// Category DTOs
// ========================

// CreateCategoryRequest represents a request to create a category
type CreateCategoryRequest struct {
	Name            string         `json:"name" validate:"required,min=2,max=100"`
	Description     string         `json:"description,omitempty"`
	ParentID        string         `json:"parentId,omitempty"`
	DepartmentID    string         `json:"departmentId,omitempty"`
	DefaultPriority TicketPriority `json:"defaultPriority,omitempty"`
	Icon            string         `json:"icon,omitempty"`
	Color           string         `json:"color,omitempty"`
	Order           int            `json:"order,omitempty"`
	IsPublic        bool           `json:"isPublic,omitempty"`
}

// UpdateCategoryRequest represents a request to update a category
type UpdateCategoryRequest struct {
	Name            *string         `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Description     *string         `json:"description,omitempty"`
	DefaultPriority *TicketPriority `json:"defaultPriority,omitempty"`
	Icon            *string         `json:"icon,omitempty"`
	Color           *string         `json:"color,omitempty"`
	Order           *int            `json:"order,omitempty"`
	IsActive        *bool           `json:"isActive,omitempty"`
	IsPublic        *bool           `json:"isPublic,omitempty"`
}

// ========================
// Agent DTOs
// ========================

// CreateAgentRequest represents a request to create an agent
type CreateAgentRequest struct {
	UserID        string    `json:"userId" validate:"required"`
	Name          string    `json:"name" validate:"required"`
	Email         string    `json:"email" validate:"required,email"`
	Phone         string    `json:"phone,omitempty"`
	Role          AgentRole `json:"role,omitempty"`
	DepartmentIDs []string  `json:"departmentIds,omitempty"`
	TeamID        string    `json:"teamId,omitempty"`
	MaxTickets    int       `json:"maxTickets,omitempty"`
	Skills        []string  `json:"skills,omitempty"`
	Languages     []string  `json:"languages,omitempty"`
}

// UpdateAgentRequest represents a request to update an agent
type UpdateAgentRequest struct {
	Name          *string    `json:"name,omitempty"`
	Phone         *string    `json:"phone,omitempty"`
	Role          *AgentRole `json:"role,omitempty"`
	DepartmentIDs []string   `json:"departmentIds,omitempty"`
	TeamID        *string    `json:"teamId,omitempty"`
	MaxTickets    *int       `json:"maxTickets,omitempty"`
	Skills        []string   `json:"skills,omitempty"`
	Languages     []string   `json:"languages,omitempty"`
	IsActive      *bool      `json:"isActive,omitempty"`
}

// UpdateAgentStatusRequest represents a request to update agent status
type UpdateAgentStatusRequest struct {
	Status AgentStatus `json:"status" validate:"required"`
}

// ========================
// SLA DTOs
// ========================

// CreateSLAPolicyRequest represents a request to create an SLA policy
type CreateSLAPolicyRequest struct {
	Name             string        `json:"name" validate:"required"`
	Description      string        `json:"description,omitempty"`
	IsDefault        bool          `json:"isDefault,omitempty"`
	Priorities       []SLAPriority `json:"priorities" validate:"required,min=1"`
	UseBusinessHours bool          `json:"useBusinessHours,omitempty"`
	BusinessHoursID  string        `json:"businessHoursId,omitempty"`
}

// UpdateSLAPolicyRequest represents a request to update an SLA policy
type UpdateSLAPolicyRequest struct {
	Name             *string       `json:"name,omitempty"`
	Description      *string       `json:"description,omitempty"`
	IsDefault        *bool         `json:"isDefault,omitempty"`
	Priorities       []SLAPriority `json:"priorities,omitempty"`
	UseBusinessHours *bool         `json:"useBusinessHours,omitempty"`
	BusinessHoursID  *string       `json:"businessHoursId,omitempty"`
	IsActive         *bool         `json:"isActive,omitempty"`
}

// ========================
// Canned Response DTOs
// ========================

// CreateCannedResponseRequest represents a request to create a canned response
type CreateCannedResponseRequest struct {
	Title        string   `json:"title" validate:"required"`
	Content      string   `json:"content" validate:"required"`
	Shortcut     string   `json:"shortcut,omitempty"`
	Category     string   `json:"category,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	IsGlobal     bool     `json:"isGlobal,omitempty"`
	DepartmentID string   `json:"departmentId,omitempty"`
}

// UpdateCannedResponseRequest represents a request to update a canned response
type UpdateCannedResponseRequest struct {
	Title    *string  `json:"title,omitempty"`
	Content  *string  `json:"content,omitempty"`
	Shortcut *string  `json:"shortcut,omitempty"`
	Category *string  `json:"category,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	IsGlobal *bool    `json:"isGlobal,omitempty"`
	IsActive *bool    `json:"isActive,omitempty"`
}

// ========================
// Filter/List DTOs
// ========================

// TicketFilter represents filter options for listing tickets
type TicketFilter struct {
	TenantID     string           `query:"tenantId"`
	Status       []TicketStatus   `query:"status"`
	Priority     []TicketPriority `query:"priority"`
	Type         []TicketType     `query:"type"`
	DepartmentID string           `query:"departmentId"`
	CategoryID   string           `query:"categoryId"`
	AssignedToID string           `query:"assignedToId"`
	CustomerID   string           `query:"customerId"`
	Tags         []string         `query:"tags"`
	SLABreached  *bool            `query:"slaBreached"`
	Unassigned   *bool            `query:"unassigned"`
	Search       string           `query:"search"`
	CreatedFrom  *time.Time       `query:"createdFrom"`
	CreatedTo    *time.Time       `query:"createdTo"`
	SortBy       string           `query:"sortBy"`
	SortOrder    string           `query:"sortOrder"`
	Page         int              `query:"page"`
	PerPage      int              `query:"perPage"`
}

// ========================
// Response DTOs
// ========================

// TicketResponse represents a ticket in API response
type TicketResponse struct {
	*Ticket
	Department     *DepartmentSummary `json:"department,omitempty"`
	Category       *CategorySummary   `json:"category,omitempty"`
	AssignedTo     *AgentSummary      `json:"assignedTo,omitempty"`
	RecentMessages []MessageSummary   `json:"recentMessages,omitempty"`
}

// DepartmentSummary represents a department summary
type DepartmentSummary struct {
	ID    primitive.ObjectID `json:"id"`
	Name  string             `json:"name"`
	Color string             `json:"color,omitempty"`
	Icon  string             `json:"icon,omitempty"`
}

// CategorySummary represents a category summary
type CategorySummary struct {
	ID   primitive.ObjectID `json:"id"`
	Name string             `json:"name"`
	Icon string             `json:"icon,omitempty"`
}

// AgentSummary represents an agent summary
type AgentSummary struct {
	ID     primitive.ObjectID `json:"id"`
	Name   string             `json:"name"`
	Email  string             `json:"email"`
	Avatar string             `json:"avatar,omitempty"`
	Status AgentStatus        `json:"status"`
}

// MessageSummary represents a message summary
type MessageSummary struct {
	ID         primitive.ObjectID `json:"id"`
	Content    string             `json:"content"`
	SenderName string             `json:"senderName"`
	SenderType SenderType         `json:"senderType"`
	CreatedAt  time.Time          `json:"createdAt"`
}

// TicketListResponse represents paginated ticket list
type TicketListResponse struct {
	Tickets    []TicketResponse `json:"tickets"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PerPage    int              `json:"perPage"`
	TotalPages int              `json:"totalPages"`
}

// TicketStats represents ticket statistics
type TicketStats struct {
	TotalTickets      int64            `json:"totalTickets"`
	OpenTickets       int64            `json:"openTickets"`
	PendingTickets    int64            `json:"pendingTickets"`
	ResolvedTickets   int64            `json:"resolvedTickets"`
	ClosedTickets     int64            `json:"closedTickets"`
	UnassignedTickets int64            `json:"unassignedTickets"`
	SLABreached       int64            `json:"slaBreached"`
	ByPriority        map[string]int64 `json:"byPriority"`
	ByDepartment      map[string]int64 `json:"byDepartment"`
	ByType            map[string]int64 `json:"byType"`
	AvgResponseTime   int64            `json:"avgResponseTime"`   // in minutes
	AvgResolutionTime int64            `json:"avgResolutionTime"` // in minutes
	AvgSatisfaction   float64          `json:"avgSatisfaction"`
}
