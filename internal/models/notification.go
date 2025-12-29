package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Notification represents a notification sent to a user
type Notification struct {
	gorm.Model
	ID      uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID  uuid.UUID  `gorm:"not null;type:uuid"`
	Type    string     `gorm:"not null"` // Enum: email, sms, push
	Content string     `gorm:"not null"`
	Status  string     `gorm:"default:'sent'"` // Enum: pending, sent, failed
	SentAt  *time.Time // Optional
}

type Message struct {
	From     string   `json:"from" binding:"required"`
	FromName string   `json:"from_name,omitempty"`
	To       []string `json:"to"`
	Subject  string   `json:"subject"`
	Body     string   `json:"body"`
}

type ProcessedFilesMessage struct {
	Recievers []string `json:"recievers"`
	Subject   string   `json:"subject"`
	Files     struct {
		Processed   []string            `json:"processed"`
		Unprocessed []UnprocessableFile `json:"unprocessed"`
	}
	Template string `json:"template" binding:"required" default:"processedFiles"`
}

type UnprocessableFile struct {
	FileName string `json:"fileName"`
	Reason   string `json:"reason"`
}
