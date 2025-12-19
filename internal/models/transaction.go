package models

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Transaction represents a financial transaction
type Transaction struct {
	gorm.Model
	ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	FromUserID  uuid.UUID      `gorm:"not null;type:uuid"`
	ToUserID    *uuid.UUID     `gorm:"type:uuid"` // Optional
	Amount      float64        `gorm:"not null"`
	Currency    string         `gorm:"not null"`
	Type        string         `gorm:"not null"`          // Enum: deposit, withdraw, transfer, payment
	Status      string         `gorm:"default:'pending'"` // Enum: pending, completed, failed, refunded
	Description *string        // Optional
	Metadata    datatypes.JSON // JSONB
	Fee         float64        `gorm:"default:0"`
	FromWallet  Wallet         `gorm:"foreignKey:FromUserID;references:UserID"`
	ToWallet    Wallet         `gorm:"foreignKey:ToUserID;references:UserID"`
}

// ToResponse converts transaction to API response format
func (t *Transaction) ToResponse() TransactionResponse {
	return TransactionResponse{
		ID:          t.ID,
		FromUserID:  t.FromUserID,
		ToUserID:    t.ToUserID,
		Amount:      t.Amount,
		Currency:    t.Currency,
		Type:        t.Type,
		Status:      t.Status,
		Description: t.Description,
		Fee:         t.Fee,
		CreatedAt:   t.CreatedAt,
	}
}
