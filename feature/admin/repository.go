package admin

import (
	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository handles admin-related database operations
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new admin repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// GetSystemStats retrieves aggregated system statistics
func (r *Repository) GetSystemStats() (map[string]interface{}, error) {
	var userCount, walletCount, kycPending int64

	if err := r.db.Model(&models.User{}).Count(&userCount).Error; err != nil {
		return nil, err
	}
	if err := r.db.Model(&models.Wallet{}).Count(&walletCount).Error; err != nil {
		return nil, err
	}
	if err := r.db.Model(&models.KYCDocument{}).Where("status = ?", "pending").Count(&kycPending).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_users":   userCount,
		"total_wallets": walletCount,
		"kyc_pending":   kycPending,
	}, nil
}

// GetTransactionStats retrieves global transaction volume and counts
func (r *Repository) GetTransactionStats() (map[string]interface{}, error) {
	var totalVolume float64
	var txCount int64

	if err := r.db.Model(&models.Transaction{}).Select("COALESCE(SUM(amount), 0)").Scan(&totalVolume).Error; err != nil {
		return nil, err
	}
	if err := r.db.Model(&models.Transaction{}).Count(&txCount).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_volume":      totalVolume,
		"transaction_count": txCount,
	}, nil
}

// ListAllUsers retrieves all users with pagination
func (r *Repository) ListAllUsers(req models.ListedRequest) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.Model(&models.User{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Order(req.OrderBy + " " + req.Order).
		Limit(req.Limit).
		Offset(req.Offset).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// UpdateUserStatus updates the suspension status of a user
func (r *Repository) UpdateUserStatus(userID uuid.UUID, isActive bool) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("is_active", isActive).Error
}

// GetAuditLogs retrieves system activity logs with pagination
func (r *Repository) GetAuditLogs(req models.ListedRequest) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := r.db.Model(&models.AuditLog{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Order(req.OrderBy + " " + req.Order).
		Limit(req.Limit).
		Offset(req.Offset).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
