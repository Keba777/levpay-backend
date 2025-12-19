package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// KYCDocument represents a user's KYC document
type KYCDocument struct {
	gorm.Model
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID     uuid.UUID `gorm:"not null;type:uuid"`
	Type       string    `gorm:"not null"`          // Enum: id_card, passport, photo, proof_of_address
	FilePath   string    `gorm:"not null"`          // S3/MinIO path
	Status     string    `gorm:"default:'pending'"` // Enum: pending, approved, rejected
	UploadedAt time.Time
}
