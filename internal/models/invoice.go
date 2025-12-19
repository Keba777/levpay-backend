package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Invoice represents a merchant invoice
type Invoice struct {
	gorm.Model
	ID            uuid.UUID   `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	MerchantID    uuid.UUID   `gorm:"not null;type:uuid"`
	Amount        float64     `gorm:"not null"`
	Currency      string      `gorm:"not null"`
	Status        string      `gorm:"default:'unpaid'"` // Enum: unpaid, paid, expired
	DueDate       *time.Time  // Optional
	TransactionID *uuid.UUID  `gorm:"type:uuid"`
	Merchant      User        `gorm:"foreignKey:MerchantID"`
	Transaction   Transaction `gorm:"foreignKey:TransactionID"`
}
