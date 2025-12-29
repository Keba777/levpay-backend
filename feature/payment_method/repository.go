package payment_method

import (
	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(pm *models.PaymentMethod) error {
	return r.db.Create(pm).Error
}

func (r *Repository) ListByUserID(userID uuid.UUID) ([]models.PaymentMethod, error) {
	var methods []models.PaymentMethod
	err := r.db.Where("user_id = ?", userID).Find(&methods).Error
	return methods, err
}

func (r *Repository) GetByID(id uuid.UUID) (*models.PaymentMethod, error) {
	var pm models.PaymentMethod
	if err := r.db.First(&pm, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &pm, nil
}

func (r *Repository) GetDefaultByUserID(userID uuid.UUID) (*models.PaymentMethod, error) {
	var pm models.PaymentMethod
	if err := r.db.Where("user_id = ? AND is_default = ?", userID, true).First(&pm).Error; err != nil {
		return nil, err
	}
	return &pm, nil
}

func (r *Repository) Update(pm *models.PaymentMethod) error {
	return r.db.Save(pm).Error
}

func (r *Repository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.PaymentMethod{}, "id = ?", id).Error
}

func (r *Repository) SetDefault(userID uuid.UUID, id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Remove default from all other methods
		if err := tx.Model(&models.PaymentMethod{}).Where("user_id = ?", userID).Update("is_default", false).Error; err != nil {
			return err
		}
		// Set new default
		return tx.Model(&models.PaymentMethod{}).Where("id = ? AND user_id = ?", id, userID).Update("is_default", true).Error
	})
}
