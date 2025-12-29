package user

import (
	"errors"
	"fmt"

	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository handles user-related database operations
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new user repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// GetUserByID retrieves a user by their ID
func (r *Repository) GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates user fields
func (r *Repository) UpdateUser(userID uuid.UUID, updates map[string]interface{}) error {
	result := r.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// ListUsers retrieves a paginated list of users based on filter
func (r *Repository) ListUsers(filter models.ListedRequest) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.Model(&models.User{})

	if filter.Keywords != "" {
		keyword := "%" + filter.Keywords + "%"
		query = query.Where("first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ? OR username ILIKE ?", keyword, keyword, keyword, keyword)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	if filter.OrderBy != "" {
		query = query.Order(fmt.Sprintf("%s %s", filter.OrderBy, filter.Order))
	} else {
		query = query.Order("created_at desc")
	}

	// Apply pagination
	if err := query.Limit(filter.Limit).Offset(filter.Offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// UpdateKYCStatus updates the KYC status of a user
func (r *Repository) UpdateKYCStatus(userID uuid.UUID, status string) error {
	return r.UpdateUser(userID, map[string]interface{}{
		"kyc_status": status,
	})
}
