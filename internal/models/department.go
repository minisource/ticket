package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Department represents a support department
type Department struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TenantID    string             `bson:"tenant_id" json:"tenantId"`
	Name        string             `bson:"name" json:"name"`
	Slug        string             `bson:"slug" json:"slug"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	Email       string             `bson:"email,omitempty" json:"email,omitempty"`

	// Hierarchy
	ParentID *primitive.ObjectID `bson:"parent_id,omitempty" json:"parentId,omitempty"`
	Path     string              `bson:"path,omitempty" json:"path,omitempty"` // e.g., "/parent/child"
	Level    int                 `bson:"level" json:"level"`

	// Members
	ManagerID   string   `bson:"manager_id,omitempty" json:"managerId,omitempty"`
	ManagerName string   `bson:"manager_name,omitempty" json:"managerName,omitempty"`
	AgentIDs    []string `bson:"agent_ids,omitempty" json:"agentIds,omitempty"`

	// Settings
	AutoAssign      bool                `bson:"auto_assign" json:"autoAssign"`
	AutoAssignType  string              `bson:"auto_assign_type" json:"autoAssignType"` // round_robin, least_busy, random
	DefaultPriority TicketPriority      `bson:"default_priority" json:"defaultPriority"`
	SLAPolicyID     *primitive.ObjectID `bson:"sla_policy_id,omitempty" json:"slaPolicyId,omitempty"`

	// Business Hours
	BusinessHours *BusinessHours `bson:"business_hours,omitempty" json:"businessHours,omitempty"`

	// Stats
	OpenTickets     int   `bson:"open_tickets" json:"openTickets"`
	TotalTickets    int   `bson:"total_tickets" json:"totalTickets"`
	AvgResponseTime int64 `bson:"avg_response_time,omitempty" json:"avgResponseTime,omitempty"` // in minutes

	// Metadata
	Metadata map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`

	// Timestamps
	CreatedAt time.Time  `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time  `bson:"updated_at" json:"updatedAt"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deletedAt,omitempty"`
	IsDeleted bool       `bson:"is_deleted" json:"isDeleted"`
	IsActive  bool       `bson:"is_active" json:"isActive"`

	// Display
	Color string `bson:"color,omitempty" json:"color,omitempty"`
	Icon  string `bson:"icon,omitempty" json:"icon,omitempty"`
	Order int    `bson:"order" json:"order"`
}

// BusinessHours represents working hours configuration
type BusinessHours struct {
	Enabled  bool          `bson:"enabled" json:"enabled"`
	Timezone string        `bson:"timezone" json:"timezone"`
	Schedule []DaySchedule `bson:"schedule" json:"schedule"`
	Holidays []Holiday     `bson:"holidays,omitempty" json:"holidays,omitempty"`
}

// DaySchedule represents working hours for a day
type DaySchedule struct {
	Day       int         `bson:"day" json:"day"` // 0=Sunday, 6=Saturday
	IsWorkDay bool        `bson:"is_work_day" json:"isWorkDay"`
	StartTime string      `bson:"start_time" json:"startTime"` // "09:00"
	EndTime   string      `bson:"end_time" json:"endTime"`     // "17:00"
	Breaks    []TimeRange `bson:"breaks,omitempty" json:"breaks,omitempty"`
}

// TimeRange represents a time range
type TimeRange struct {
	Start string `bson:"start" json:"start"`
	End   string `bson:"end" json:"end"`
}

// Holiday represents a holiday
type Holiday struct {
	Date        time.Time `bson:"date" json:"date"`
	Name        string    `bson:"name" json:"name"`
	Description string    `bson:"description,omitempty" json:"description,omitempty"`
}

// Category represents a ticket category
type Category struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TenantID    string             `bson:"tenant_id" json:"tenantId"`
	Name        string             `bson:"name" json:"name"`
	Slug        string             `bson:"slug" json:"slug"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	Icon        string             `bson:"icon,omitempty" json:"icon,omitempty"`
	Color       string             `bson:"color,omitempty" json:"color,omitempty"`

	// Hierarchy
	ParentID *primitive.ObjectID `bson:"parent_id,omitempty" json:"parentId,omitempty"`
	Path     string              `bson:"path,omitempty" json:"path,omitempty"`
	Level    int                 `bson:"level" json:"level"`

	// Relations
	DepartmentID *primitive.ObjectID `bson:"department_id,omitempty" json:"departmentId,omitempty"`

	// Settings
	DefaultPriority TicketPriority   `bson:"default_priority,omitempty" json:"defaultPriority,omitempty"`
	RequiredFields  []string         `bson:"required_fields,omitempty" json:"requiredFields,omitempty"`
	CustomFields    []CustomFieldDef `bson:"custom_fields,omitempty" json:"customFields,omitempty"`

	// Display
	Order    int  `bson:"order" json:"order"`
	IsActive bool `bson:"is_active" json:"isActive"`
	IsPublic bool `bson:"is_public" json:"isPublic"` // Visible in customer portal

	// Timestamps
	CreatedAt time.Time  `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time  `bson:"updated_at" json:"updatedAt"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deletedAt,omitempty"`
	IsDeleted bool       `bson:"is_deleted" json:"isDeleted"`
}

// CustomFieldDef represents a custom field definition
type CustomFieldDef struct {
	Name         string      `bson:"name" json:"name"`
	Label        string      `bson:"label" json:"label"`
	Type         string      `bson:"type" json:"type"` // text, number, select, multiselect, date, checkbox
	Required     bool        `bson:"required" json:"required"`
	Options      []string    `bson:"options,omitempty" json:"options,omitempty"` // For select/multiselect
	DefaultValue interface{} `bson:"default_value,omitempty" json:"defaultValue,omitempty"`
	Placeholder  string      `bson:"placeholder,omitempty" json:"placeholder,omitempty"`
	HelpText     string      `bson:"help_text,omitempty" json:"helpText,omitempty"`
	Order        int         `bson:"order" json:"order"`
}
