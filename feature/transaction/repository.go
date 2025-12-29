package transaction

import (
	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository handles transaction-related database operations
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new transaction repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// CreateTransaction records a new transaction
func (r *Repository) CreateTransaction(tx *models.Transaction) error {
	return r.db.Create(tx).Error
}

// GetTransactionByID retrieves a transaction by ID
func (r *Repository) GetTransactionByID(id uuid.UUID) (*models.Transaction, error) {
	var transaction models.Transaction
	if err := r.db.Where("id = ?", id).First(&transaction).Error; err != nil {
		return nil, err
	}
	return &transaction, nil
}

// GetUserTransactions retrieves all transactions for a user with pagination
func (r *Repository) GetUserTransactions(userID uuid.UUID, req models.ListedRequest) ([]models.Transaction, int64, error) {
	var transactions []models.Transaction
	var total int64

	query := r.db.Model(&models.Transaction{}).
		Where("from_user_id = ? OR to_user_id = ?", userID, userID)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and ordering
	if err := query.
		Order(req.OrderBy + " " + req.Order).
		Limit(req.Limit).
		Offset(req.Offset).
		Find(&transactions).Error; err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

// UpdateTransactionStatus updates the status of a transaction
func (r *Repository) UpdateTransactionStatus(id uuid.UUID, status string) error {
	return r.db.Model(&models.Transaction{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// GetTransactionsByType retrieves transactions by type
func (r *Repository) GetTransactionsByType(userID uuid.UUID, transactionType string, limit int) ([]models.Transaction, error) {
	var transactions []models.Transaction

	err := r.db.Where("(from_user_id = ? OR to_user_id = ?) AND type = ?", userID, userID, transactionType).
		Order("created_at desc").
		Limit(limit).
		Find(&transactions).Error

	if err != nil {
		return nil, err
	}

	return transactions, nil
}

// GetPendingTransactions retrieves all pending transactions (admin/system use)
func (r *Repository) GetPendingTransactions() ([]models.Transaction, error) {
	var transactions []models.Transaction

	err := r.db.Where("status = ?", models.TransactionStatusPending).
		Order("created_at asc").
		Find(&transactions).Error

	if err != nil {
		return nil, err
	}

	return transactions, nil
}
