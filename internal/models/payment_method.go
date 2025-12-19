package models

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// PaymentMethod represents a user's linked payment method
type PaymentMethod struct {
	gorm.Model
	ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    uuid.UUID      `gorm:"not null;type:uuid"`
	Type      string         `gorm:"not null"` // Enum: bank, card, mobile_wallet
	Details   datatypes.JSON `gorm:"not null"` // Encrypted JSON
	IsDefault bool           `gorm:"default:false"`
	Verified  bool           `gorm:"default:false"`
}
