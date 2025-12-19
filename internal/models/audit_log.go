package models

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AuditLog represents a system audit log
type AuditLog struct {
	gorm.Model
	ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    *uuid.UUID     `gorm:"type:uuid"` // Optional
	Action    string         `gorm:"not null"`
	IPAddress *string        // Optional
	Details   datatypes.JSON // Flexible JSON data
	User      User           `gorm:"foreignKey:UserID"`
}
