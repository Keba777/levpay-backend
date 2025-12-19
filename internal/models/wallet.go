package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Wallet represents a user's wallet
type Wallet struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID      uuid.UUID `gorm:"unique;not null;type:uuid"`
	Balance     float64   `gorm:"default:0"`
	Currency    string    `gorm:"default:'ETB'"`
	Locked      bool      `gorm:"default:false"`
	LastUpdated time.Time
}

// ToResponse converts wallet to API response format
func (w *Wallet) ToResponse() WalletResponse {
	return WalletResponse{
		ID:          w.ID,
		UserID:      w.UserID,
		Balance:     w.Balance,
		Currency:    w.Currency,
		Locked:      w.Locked,
		LastUpdated: w.LastUpdated,
	}
}
