package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Session represents a user session with refresh token
type Session struct {
	gorm.Model
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID       uuid.UUID `gorm:"type:uuid;not null"`
	RefreshToken string    `gorm:"not null"`
	Fingerprint  string    `gorm:"not null"` // Device fingerprint for security
	Active       bool      `gorm:"type:bool;default:true"`
	Terminated   *time.Time
	User         User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}
