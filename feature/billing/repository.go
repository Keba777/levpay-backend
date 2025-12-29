package billing

import (
	"fmt"
	"time"

	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository handles billing-related database operations
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new billing repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// GenerateInvoiceNumber generates a unique invoice number
func (r *Repository) GenerateInvoiceNumber() (string, error) {
	var count int64
	year := time.Now().Year()

	// Count invoices this year
	r.db.Model(&models.Invoice{}).
		Where("EXTRACT(YEAR FROM created_at) = ?", year).
		Count(&count)

	return fmt.Sprintf("INV-%d-%05d", year, count+1), nil
}

// CreateInvoice creates a new invoice
func (r *Repository) CreateInvoice(invoice *models.Invoice) error {
	// Generate invoice number if not provided
	if invoice.InvoiceNumber == "" {
		number, err := r.GenerateInvoiceNumber()
		if err != nil {
			return err
		}
		invoice.InvoiceNumber = number
	}

	return r.db.Create(invoice).Error
}

// GetInvoiceByID retrieves an invoice by ID
func (r *Repository) GetInvoiceByID(id uuid.UUID) (*models.Invoice, error) {
	var invoice models.Invoice
	if err := r.db.Where("id = ?", id).First(&invoice).Error; err != nil {
		return nil, err
	}
	return &invoice, nil
}

// GetMerchantInvoices retrieves all invoices for a merchant with pagination
func (r *Repository) GetMerchantInvoices(merchantID uuid.UUID, req models.ListedRequest) ([]models.Invoice, int64, error) {
	var invoices []models.Invoice
	var total int64

	query := r.db.Model(&models.Invoice{}).Where("merchant_id = ?", merchantID)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and ordering
	if err := query.
		Order(req.OrderBy + " " + req.Order).
		Limit(req.Limit).
		Offset(req.Offset).
		Find(&invoices).Error; err != nil {
		return nil, 0, err
	}

	return invoices, total, nil
}

// GetCustomerInvoices retrieves all invoices for a customer with pagination
func (r *Repository) GetCustomerInvoices(customerID uuid.UUID, req models.ListedRequest) ([]models.Invoice, int64, error) {
	var invoices []models.Invoice
	var total int64

	query := r.db.Model(&models.Invoice{}).Where("customer_id = ?", customerID)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and ordering
	if err := query.
		Order(req.OrderBy + " " + req.Order).
		Limit(req.Limit).
		Offset(req.Offset).
		Find(&invoices).Error; err != nil {
		return nil, 0, err
	}

	return invoices, total, nil
}

// UpdateInvoiceStatus updates the status of an invoice
func (r *Repository) UpdateInvoiceStatus(id uuid.UUID, status string) error {
	return r.db.Model(&models.Invoice{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// MarkInvoiceAsPaid marks an invoice as paid
func (r *Repository) MarkInvoiceAsPaid(id uuid.UUID, transactionID uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&models.Invoice{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":         models.InvoiceStatusPaid,
			"transaction_id": transactionID,
			"paid_at":        now,
		}).Error
}

// GetOverdueInvoices retrieves all overdue invoices
func (r *Repository) GetOverdueInvoices() ([]models.Invoice, error) {
	var invoices []models.Invoice
	now := time.Now()

	err := r.db.Where("status != ? AND status != ? AND due_date < ?",
		models.InvoiceStatusPaid,
		models.InvoiceStatusCancelled,
		now).
		Find(&invoices).Error

	return invoices, err
}
