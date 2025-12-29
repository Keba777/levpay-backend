package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Invoice Status Constants
const (
	InvoiceStatusDraft     = "draft"
	InvoiceStatusSent      = "sent"
	InvoiceStatusPaid      = "paid"
	InvoiceStatusOverdue   = "overdue"
	InvoiceStatusCancelled = "cancelled"
)

// Invoice represents a merchant invoice
type Invoice struct {
	gorm.Model
	ID            uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	InvoiceNumber string     `gorm:"unique;not null"`
	MerchantID    uuid.UUID  `gorm:"not null;type:uuid;index"`
	CustomerID    *uuid.UUID `gorm:"type:uuid;index"` // Optional - can be null for general invoices
	Amount        float64    `gorm:"not null"`
	Currency      string     `gorm:"not null;default:'ETB'"`
	Status        string     `gorm:"default:'draft';index"` // draft, sent, paid, overdue, cancelled
	Description   *string
	DueDate       *time.Time
	PaidAt        *time.Time
	TransactionID *uuid.UUID  `gorm:"type:uuid"`
	Merchant      User        `gorm:"foreignKey:MerchantID"`
	Customer      User        `gorm:"foreignKey:CustomerID"`
	Transaction   Transaction `gorm:"foreignKey:TransactionID"`
}

// InvoiceResponse for API responses
type InvoiceResponse struct {
	ID            uuid.UUID  `json:"id"`
	InvoiceNumber string     `json:"invoice_number"`
	MerchantID    uuid.UUID  `json:"merchant_id"`
	CustomerID    *uuid.UUID `json:"customer_id,omitempty"`
	Amount        float64    `json:"amount"`
	Currency      string     `json:"currency"`
	Status        string     `json:"status"`
	Description   *string    `json:"description,omitempty"`
	DueDate       *time.Time `json:"due_date,omitempty"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
	TransactionID *uuid.UUID `json:"transaction_id,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ToResponse converts invoice to API response format
func (i *Invoice) ToResponse() InvoiceResponse {
	return InvoiceResponse{
		ID:            i.ID,
		InvoiceNumber: i.InvoiceNumber,
		MerchantID:    i.MerchantID,
		CustomerID:    i.CustomerID,
		Amount:        i.Amount,
		Currency:      i.Currency,
		Status:        i.Status,
		Description:   i.Description,
		DueDate:       i.DueDate,
		PaidAt:        i.PaidAt,
		TransactionID: i.TransactionID,
		CreatedAt:     i.CreatedAt,
	}
}
