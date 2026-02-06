package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TicketStatus represents the status of a ticket
type TicketStatus string

const (
	StatusOpen       TicketStatus = "open"
	StatusInProgress TicketStatus = "in_progress"
	StatusPending    TicketStatus = "pending"   // Waiting for customer response
	StatusOnHold     TicketStatus = "on_hold"   // Temporarily paused
	StatusResolved   TicketStatus = "resolved"  // Solution provided
	StatusClosed     TicketStatus = "closed"    // Ticket closed
	StatusReopened   TicketStatus = "reopened"  // Reopened by customer
	StatusEscalated  TicketStatus = "escalated" // Escalated to higher level
	StatusCancelled  TicketStatus = "cancelled" // Cancelled by customer or agent
)

// TicketPriority represents the priority of a ticket
type TicketPriority string

const (
	PriorityLow      TicketPriority = "low"
	PriorityMedium   TicketPriority = "medium"
	PriorityHigh     TicketPriority = "high"
	PriorityUrgent   TicketPriority = "urgent"
	PriorityCritical TicketPriority = "critical"
)

// TicketSource represents how the ticket was created
type TicketSource string

const (
	SourceWeb      TicketSource = "web"
	SourceEmail    TicketSource = "email"
	SourceAPI      TicketSource = "api"
	SourcePhone    TicketSource = "phone"
	SourceChat     TicketSource = "chat"
	SourceMobile   TicketSource = "mobile"
	SourceInternal TicketSource = "internal"
)

// TicketType represents the type of ticket
type TicketType string

const (
	TypeQuestion  TicketType = "question"
	TypeIncident  TicketType = "incident"
	TypeProblem   TicketType = "problem"
	TypeFeature   TicketType = "feature_request"
	TypeBug       TicketType = "bug"
	TypeTask      TicketType = "task"
	TypeComplaint TicketType = "complaint"
	TypeFeedback  TicketType = "feedback"
)

// Ticket represents a support ticket
type Ticket struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TenantID     string             `bson:"tenant_id" json:"tenantId"`
	TicketNumber string             `bson:"ticket_number" json:"ticketNumber"` // Human-readable ticket number

	// Basic Info
	Subject     string         `bson:"subject" json:"subject"`
	Description string         `bson:"description" json:"description"`
	Type        TicketType     `bson:"type" json:"type"`
	Status      TicketStatus   `bson:"status" json:"status"`
	Priority    TicketPriority `bson:"priority" json:"priority"`
	Source      TicketSource   `bson:"source" json:"source"`

	// Customer Info
	CustomerID     string `bson:"customer_id" json:"customerId"`
	CustomerName   string `bson:"customer_name" json:"customerName"`
	CustomerEmail  string `bson:"customer_email" json:"customerEmail"`
	CustomerPhone  string `bson:"customer_phone,omitempty" json:"customerPhone,omitempty"`
	CustomerAvatar string `bson:"customer_avatar,omitempty" json:"customerAvatar,omitempty"`

	// Organization
	DepartmentID    *primitive.ObjectID `bson:"department_id,omitempty" json:"departmentId,omitempty"`
	DepartmentName  string              `bson:"department_name,omitempty" json:"departmentName,omitempty"`
	CategoryID      *primitive.ObjectID `bson:"category_id,omitempty" json:"categoryId,omitempty"`
	CategoryName    string              `bson:"category_name,omitempty" json:"categoryName,omitempty"`
	SubcategoryID   *primitive.ObjectID `bson:"subcategory_id,omitempty" json:"subcategoryId,omitempty"`
	SubcategoryName string              `bson:"subcategory_name,omitempty" json:"subcategoryName,omitempty"`

	// Assignment
	AssignedToID    string     `bson:"assigned_to_id,omitempty" json:"assignedToId,omitempty"`
	AssignedToName  string     `bson:"assigned_to_name,omitempty" json:"assignedToName,omitempty"`
	AssignedToEmail string     `bson:"assigned_to_email,omitempty" json:"assignedToEmail,omitempty"`
	AssignedAt      *time.Time `bson:"assigned_at,omitempty" json:"assignedAt,omitempty"`
	AssignedByID    string     `bson:"assigned_by_id,omitempty" json:"assignedById,omitempty"`
	TeamID          string     `bson:"team_id,omitempty" json:"teamId,omitempty"`
	TeamName        string     `bson:"team_name,omitempty" json:"teamName,omitempty"`

	// SLA
	SLAPolicyID         *primitive.ObjectID `bson:"sla_policy_id,omitempty" json:"slaPolicyId,omitempty"`
	FirstResponseDue    *time.Time          `bson:"first_response_due,omitempty" json:"firstResponseDue,omitempty"`
	ResolutionDue       *time.Time          `bson:"resolution_due,omitempty" json:"resolutionDue,omitempty"`
	FirstResponsedAt    *time.Time          `bson:"first_responsed_at,omitempty" json:"firstResponsedAt,omitempty"`
	SLABreached         bool                `bson:"sla_breached" json:"slaBreached"`
	ResponseSLABreached bool                `bson:"response_sla_breached" json:"responseSlaBreached"`
	ResolveSLABreached  bool                `bson:"resolve_sla_breached" json:"resolveSlaBreached"`

	// Related
	ParentTicketID   *primitive.ObjectID  `bson:"parent_ticket_id,omitempty" json:"parentTicketId,omitempty"`
	RelatedTicketIDs []primitive.ObjectID `bson:"related_ticket_ids,omitempty" json:"relatedTicketIds,omitempty"`
	MergedIntoID     *primitive.ObjectID  `bson:"merged_into_id,omitempty" json:"mergedIntoId,omitempty"`
	MergedTicketIDs  []primitive.ObjectID `bson:"merged_ticket_ids,omitempty" json:"mergedTicketIds,omitempty"`

	// Attachments & Tags
	Attachments  []Attachment           `bson:"attachments,omitempty" json:"attachments,omitempty"`
	Tags         []string               `bson:"tags,omitempty" json:"tags,omitempty"`
	CustomFields map[string]interface{} `bson:"custom_fields,omitempty" json:"customFields,omitempty"`

	// Stats
	MessageCount    int `bson:"message_count" json:"messageCount"`
	InternalNotes   int `bson:"internal_notes" json:"internalNotes"`
	ReopenCount     int `bson:"reopen_count" json:"reopenCount"`
	EscalationLevel int `bson:"escalation_level" json:"escalationLevel"`

	// Rating
	SatisfactionRating  *int       `bson:"satisfaction_rating,omitempty" json:"satisfactionRating,omitempty"`
	SatisfactionComment string     `bson:"satisfaction_comment,omitempty" json:"satisfactionComment,omitempty"`
	RatedAt             *time.Time `bson:"rated_at,omitempty" json:"ratedAt,omitempty"`

	// Watchers
	WatcherIDs []string `bson:"watcher_ids,omitempty" json:"watcherIds,omitempty"`
	CCEmails   []string `bson:"cc_emails,omitempty" json:"ccEmails,omitempty"`

	// Metadata
	IPAddress string                 `bson:"ip_address,omitempty" json:"-"`
	UserAgent string                 `bson:"user_agent,omitempty" json:"-"`
	Metadata  map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`

	// Timestamps
	CreatedAt  time.Time  `bson:"created_at" json:"createdAt"`
	UpdatedAt  time.Time  `bson:"updated_at" json:"updatedAt"`
	ResolvedAt *time.Time `bson:"resolved_at,omitempty" json:"resolvedAt,omitempty"`
	ClosedAt   *time.Time `bson:"closed_at,omitempty" json:"closedAt,omitempty"`
	DeletedAt  *time.Time `bson:"deleted_at,omitempty" json:"deletedAt,omitempty"`
	IsDeleted  bool       `bson:"is_deleted" json:"isDeleted"`
	DeletedBy  string     `bson:"deleted_by,omitempty" json:"deletedBy,omitempty"`

	// Last Activity
	LastActivityAt      time.Time  `bson:"last_activity_at" json:"lastActivityAt"`
	LastCustomerReplyAt *time.Time `bson:"last_customer_reply_at,omitempty" json:"lastCustomerReplyAt,omitempty"`
	LastAgentReplyAt    *time.Time `bson:"last_agent_reply_at,omitempty" json:"lastAgentReplyAt,omitempty"`
}

// Attachment represents a file attached to a ticket or message
type Attachment struct {
	ID         string    `bson:"id" json:"id"`
	Name       string    `bson:"name" json:"name"`
	URL        string    `bson:"url" json:"url"`
	Size       int64     `bson:"size" json:"size"`
	MimeType   string    `bson:"mime_type" json:"mimeType"`
	UploadedBy string    `bson:"uploaded_by" json:"uploadedBy"`
	UploadedAt time.Time `bson:"uploaded_at" json:"uploadedAt"`
}

// TicketHistory represents a change in ticket history
type TicketHistory struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TicketID      primitive.ObjectID `bson:"ticket_id" json:"ticketId"`
	TenantID      string             `bson:"tenant_id" json:"tenantId"`
	Action        string             `bson:"action" json:"action"` // created, updated, assigned, status_changed, etc.
	Field         string             `bson:"field,omitempty" json:"field,omitempty"`
	OldValue      interface{}        `bson:"old_value,omitempty" json:"oldValue,omitempty"`
	NewValue      interface{}        `bson:"new_value,omitempty" json:"newValue,omitempty"`
	ChangedBy     string             `bson:"changed_by" json:"changedBy"`
	ChangedByName string             `bson:"changed_by_name" json:"changedByName"`
	Comment       string             `bson:"comment,omitempty" json:"comment,omitempty"`
	CreatedAt     time.Time          `bson:"created_at" json:"createdAt"`
}

// TicketCounter for generating sequential ticket numbers per tenant
type TicketCounter struct {
	ID       string `bson:"_id" json:"id"` // tenant_id
	Sequence int64  `bson:"sequence" json:"sequence"`
}
