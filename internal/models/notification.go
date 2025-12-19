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
