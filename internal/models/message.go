package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MessageType represents the type of message
type MessageType string

const (
	MessageTypeReply        MessageType = "reply"         // Public reply
	MessageTypeInternalNote MessageType = "internal_note" // Internal note (agents only)
	MessageTypeSystem       MessageType = "system"        // System generated message
	MessageTypeAutoReply    MessageType = "auto_reply"    // Auto-generated reply
)

// SenderType represents who sent the message
type SenderType string

const (
	SenderCustomer SenderType = "customer"
	SenderAgent    SenderType = "agent"
	SenderSystem   SenderType = "system"
	SenderBot      SenderType = "bot"
)

// TicketMessage represents a message/reply in a ticket
type TicketMessage struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TicketID primitive.ObjectID `bson:"ticket_id" json:"ticketId"`
	TenantID string             `bson:"tenant_id" json:"tenantId"`

	// Message Content
	Type        MessageType `bson:"type" json:"type"`
	Content     string      `bson:"content" json:"content"`
	ContentHTML string      `bson:"content_html,omitempty" json:"contentHtml,omitempty"`

	// Sender Info
	SenderType   SenderType `bson:"sender_type" json:"senderType"`
	SenderID     string     `bson:"sender_id" json:"senderId"`
	SenderName   string     `bson:"sender_name" json:"senderName"`
	SenderEmail  string     `bson:"sender_email,omitempty" json:"senderEmail,omitempty"`
	SenderAvatar string     `bson:"sender_avatar,omitempty" json:"senderAvatar,omitempty"`

	// Attachments
	Attachments []Attachment `bson:"attachments,omitempty" json:"attachments,omitempty"`

	// Email Info (for messages from email)
	EmailMessageID string            `bson:"email_message_id,omitempty" json:"emailMessageId,omitempty"`
	EmailFrom      string            `bson:"email_from,omitempty" json:"emailFrom,omitempty"`
	EmailTo        []string          `bson:"email_to,omitempty" json:"emailTo,omitempty"`
	EmailCC        []string          `bson:"email_cc,omitempty" json:"emailCc,omitempty"`
	EmailHeaders   map[string]string `bson:"email_headers,omitempty" json:"emailHeaders,omitempty"`

	// Editing
	IsEdited        bool       `bson:"is_edited" json:"isEdited"`
	EditedAt        *time.Time `bson:"edited_at,omitempty" json:"editedAt,omitempty"`
	EditedBy        string     `bson:"edited_by,omitempty" json:"editedBy,omitempty"`
	OriginalContent string     `bson:"original_content,omitempty" json:"-"`

	// Visibility
	IsPrivate bool `bson:"is_private" json:"isPrivate"` // Only visible to agents

	// Metadata
	IPAddress string                 `bson:"ip_address,omitempty" json:"-"`
	UserAgent string                 `bson:"user_agent,omitempty" json:"-"`
	Metadata  map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`

	// Timestamps
	CreatedAt time.Time  `bson:"created_at" json:"createdAt"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deletedAt,omitempty"`
	IsDeleted bool       `bson:"is_deleted" json:"isDeleted"`
	DeletedBy string     `bson:"deleted_by,omitempty" json:"deletedBy,omitempty"`
}

// CannedResponse represents a pre-written response template
type CannedResponse struct {
	ID           primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	TenantID     string              `bson:"tenant_id" json:"tenantId"`
	Title        string              `bson:"title" json:"title"`
	Content      string              `bson:"content" json:"content"`
	Shortcut     string              `bson:"shortcut,omitempty" json:"shortcut,omitempty"` // e.g., /thank
	Category     string              `bson:"category,omitempty" json:"category,omitempty"`
	Tags         []string            `bson:"tags,omitempty" json:"tags,omitempty"`
	CreatedBy    string              `bson:"created_by" json:"createdBy"`
	IsGlobal     bool                `bson:"is_global" json:"isGlobal"` // Available to all agents
	DepartmentID *primitive.ObjectID `bson:"department_id,omitempty" json:"departmentId,omitempty"`
	UsageCount   int                 `bson:"usage_count" json:"usageCount"`
	CreatedAt    time.Time           `bson:"created_at" json:"createdAt"`
	UpdatedAt    time.Time           `bson:"updated_at" json:"updatedAt"`
	IsActive     bool                `bson:"is_active" json:"isActive"`
}
