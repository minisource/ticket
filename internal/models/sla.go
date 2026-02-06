package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SLAPriority defines SLA targets for a specific priority level
type SLAPriority struct {
	Priority            TicketPriority `bson:"priority" json:"priority"`
	FirstResponseMins   int            `bson:"first_response_mins" json:"firstResponseMins"`
	ResolutionMins      int            `bson:"resolution_mins" json:"resolutionMins"`
	NextResponseMins    int            `bson:"next_response_mins" json:"nextResponseMins"`
	EscalationEnabled   bool           `bson:"escalation_enabled" json:"escalationEnabled"`
	EscalationAfterMins int            `bson:"escalation_after_mins" json:"escalationAfterMins"`
}

// SLAPolicy represents an SLA policy
type SLAPolicy struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TenantID    string             `bson:"tenant_id" json:"tenantId"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	IsDefault   bool               `bson:"is_default" json:"isDefault"`
	IsActive    bool               `bson:"is_active" json:"isActive"`

	// Targets per priority
	Priorities []SLAPriority `bson:"priorities" json:"priorities"`

	// Business Hours
	UseBusinessHours bool                `bson:"use_business_hours" json:"useBusinessHours"`
	BusinessHoursID  *primitive.ObjectID `bson:"business_hours_id,omitempty" json:"businessHoursId,omitempty"`

	// Escalation
	EscalationContacts []EscalationContact `bson:"escalation_contacts,omitempty" json:"escalationContacts,omitempty"`

	// Timestamps
	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time `bson:"updated_at" json:"updatedAt"`
}

// EscalationContact represents who to notify on escalation
type EscalationContact struct {
	Level           int      `bson:"level" json:"level"` // 1, 2, 3...
	UserIDs         []string `bson:"user_ids,omitempty" json:"userIds,omitempty"`
	Emails          []string `bson:"emails,omitempty" json:"emails,omitempty"`
	NotifyAfterMins int      `bson:"notify_after_mins" json:"notifyAfterMins"`
}

// Agent represents a support agent
type Agent struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	TenantID      string               `bson:"tenant_id" json:"tenantId"`
	UserID        string               `bson:"user_id" json:"userId"` // Reference to auth service user
	Name          string               `bson:"name" json:"name"`
	Email         string               `bson:"email" json:"email"`
	Avatar        string               `bson:"avatar,omitempty" json:"avatar,omitempty"`
	Phone         string               `bson:"phone,omitempty" json:"phone,omitempty"`
	Role          AgentRole            `bson:"role" json:"role"`
	DepartmentIDs []primitive.ObjectID `bson:"department_ids,omitempty" json:"departmentIds,omitempty"`
	TeamID        *primitive.ObjectID  `bson:"team_id,omitempty" json:"teamId,omitempty"`

	// Availability
	Status       AgentStatus `bson:"status" json:"status"`
	IsOnline     bool        `bson:"is_online" json:"isOnline"`
	LastActiveAt *time.Time  `bson:"last_active_at,omitempty" json:"lastActiveAt,omitempty"`

	// Capacity
	MaxTickets       int `bson:"max_tickets" json:"maxTickets"`
	CurrentTickets   int `bson:"current_tickets" json:"currentTickets"`
	TicketsToday     int `bson:"tickets_today" json:"ticketsToday"`
	TicketsThisWeek  int `bson:"tickets_this_week" json:"ticketsThisWeek"`
	TicketsThisMonth int `bson:"tickets_this_month" json:"ticketsThisMonth"`

	// Skills
	Skills    []string `bson:"skills,omitempty" json:"skills,omitempty"`
	Languages []string `bson:"languages,omitempty" json:"languages,omitempty"`

	// Stats
	AvgResponseTime int64   `bson:"avg_response_time,omitempty" json:"avgResponseTime,omitempty"` // in minutes
	AvgRating       float64 `bson:"avg_rating,omitempty" json:"avgRating,omitempty"`
	TotalResolved   int     `bson:"total_resolved" json:"totalResolved"`

	// Preferences
	Preferences AgentPreferences `bson:"preferences,omitempty" json:"preferences,omitempty"`

	// Timestamps
	CreatedAt time.Time  `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time  `bson:"updated_at" json:"updatedAt"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deletedAt,omitempty"`
	IsDeleted bool       `bson:"is_deleted" json:"isDeleted"`
	IsActive  bool       `bson:"is_active" json:"isActive"`
}

// AgentRole represents the role of an agent
type AgentRole string

const (
	RoleAgent      AgentRole = "agent"
	RoleSupervisor AgentRole = "supervisor"
	RoleManager    AgentRole = "manager"
	RoleAdmin      AgentRole = "admin"
)

// AgentStatus represents the availability status
type AgentStatus string

const (
	AgentStatusAvailable AgentStatus = "available"
	AgentStatusBusy      AgentStatus = "busy"
	AgentStatusAway      AgentStatus = "away"
	AgentStatusOffline   AgentStatus = "offline"
	AgentStatusOnBreak   AgentStatus = "on_break"
)

// AgentPreferences represents agent preferences
type AgentPreferences struct {
	EmailNotifications   bool   `bson:"email_notifications" json:"emailNotifications"`
	PushNotifications    bool   `bson:"push_notifications" json:"pushNotifications"`
	DesktopNotifications bool   `bson:"desktop_notifications" json:"desktopNotifications"`
	SoundEnabled         bool   `bson:"sound_enabled" json:"soundEnabled"`
	Language             string `bson:"language" json:"language"`
	Timezone             string `bson:"timezone" json:"timezone"`
	Signature            string `bson:"signature,omitempty" json:"signature,omitempty"`
}

// Team represents a group of agents
type Team struct {
	ID           primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	TenantID     string              `bson:"tenant_id" json:"tenantId"`
	Name         string              `bson:"name" json:"name"`
	Description  string              `bson:"description,omitempty" json:"description,omitempty"`
	LeaderID     string              `bson:"leader_id,omitempty" json:"leaderId,omitempty"`
	LeaderName   string              `bson:"leader_name,omitempty" json:"leaderName,omitempty"`
	MemberIDs    []string            `bson:"member_ids,omitempty" json:"memberIds,omitempty"`
	DepartmentID *primitive.ObjectID `bson:"department_id,omitempty" json:"departmentId,omitempty"`
	IsActive     bool                `bson:"is_active" json:"isActive"`
	CreatedAt    time.Time           `bson:"created_at" json:"createdAt"`
	UpdatedAt    time.Time           `bson:"updated_at" json:"updatedAt"`
}
