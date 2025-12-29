package models

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Transaction Type Constants
const (
	TransactionTypeDeposit  = "deposit"
	TransactionTypeWithdraw = "withdraw"
	TransactionTypeTransfer = "transfer"
	TransactionTypePayment  = "payment"
	TransactionTypeTopUp    = "topup"
)

// Transaction Status Constants
const (
	TransactionStatusPending   = "pending"
	TransactionStatusCompleted = "completed"
	TransactionStatusFailed    = "failed"
	TransactionStatusRefunded  = "refunded"
)

// Transaction represents a financial transaction
type Transaction struct {
	gorm.Model
	ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	FromUserID  uuid.UUID      `gorm:"not null;type:uuid"`
	ToUserID    *uuid.UUID     `gorm:"type:uuid"` // Optional (null for topup/withdraw)
	Amount      float64        `gorm:"not null"`
	Currency    string         `gorm:"not null;default:'ETB'"`
	Type        string         `gorm:"not null"`          // deposit, withdraw, transfer, payment, topup
	Status      string         `gorm:"default:'pending'"` // pending, completed, failed, refunded
	Description *string        // Optional
	Metadata    datatypes.JSON `gorm:"type:jsonb"` // JSONB for additional data
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
