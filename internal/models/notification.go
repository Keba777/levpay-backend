package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Notification Type Constants
const (
	NotificationTypeEmail = "email"
	NotificationTypeSMS   = "sms"
	NotificationTypePush  = "push"
)

// Notification Status Constants
const (
	NotificationStatusPending = "pending"
	NotificationStatusSent    = "sent"
	NotificationStatusFailed  = "failed"
)

// Notification represents a notification sent to a user
type Notification struct {
	gorm.Model
	ID      uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID  uuid.UUID `gorm:"not null;type:uuid"`
	Type    string    `gorm:"not null"` // email, sms, push
	Title   string    `gorm:"not null"`
	Content string    `gorm:"not null"`
	Status  string    `gorm:"default:'pending'"` // pending, sent, failed
	Read    bool      `gorm:"default:false"`
	SentAt  *time.Time
}

// Message represents an email message structure
type Message struct {
	From     string   `json:"from" binding:"required"`
	FromName string   `json:"from_name,omitempty"`
	To       []string `json:"to"`
	Subject  string   `json:"subject"`
	Body     string   `json:"body"`
}

// ProcessedFilesMessage represents a notification about processed files
type ProcessedFilesMessage struct {
	Recievers []string `json:"recievers"`
	Subject   string   `json:"subject"`
	Files     struct {
		Processed   []string            `json:"processed"`
		Unprocessed []UnprocessableFile `json:"unprocessed"`
	}
	Template string `json:"template" binding:"required" default:"processedFiles"`
}

// UnprocessableFile represents a file that couldn't be processed
type UnprocessableFile struct {
	FileName string `json:"fileName"`
	Reason   string `json:"reason"`
}

// NotificationResponse for API responses
type NotificationResponse struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Status    string     `json:"status"`
	Read      bool       `json:"read"`
	SentAt    *time.Time `json:"sent_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// ToResponse converts notification to API response format
func (n *Notification) ToResponse() NotificationResponse {
	return NotificationResponse{
		ID:        n.ID,
		UserID:    n.UserID,
		Type:      n.Type,
		Title:     n.Title,
		Content:   n.Content,
		Status:    n.Status,
		Read:      n.Read,
		SentAt:    n.SentAt,
		CreatedAt: n.CreatedAt,
	}
}
