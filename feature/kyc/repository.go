package kyc

import (
	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository handles KYC-related database operations
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new KYC repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// CreateDocument records a new KYC document upload
func (r *Repository) CreateDocument(doc *models.KYCDocument) error {
	return r.db.Create(doc).Error
}

// GetDocumentsByUserID retrieves all documents for a user
func (r *Repository) GetDocumentsByUserID(userID uuid.UUID) ([]models.KYCDocument, error) {
	var docs []models.KYCDocument
	if err := r.db.Where("user_id = ?", userID).Find(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
}

// GetPendingDocuments retrieves all documents with 'pending' status
func (r *Repository) GetPendingDocuments() ([]models.KYCDocument, error) {
	var docs []models.KYCDocument
	if err := r.db.Where("status = ?", models.KYCStatusPending).Order("uploaded_at asc").Find(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
}

// UpdateDocumentStatus updates the status and notes of a document
func (r *Repository) UpdateDocumentStatus(id uuid.UUID, status, notes string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if notes != "" {
		updates["notes"] = notes
	}
	return r.db.Model(&models.KYCDocument{}).Where("id = ?", id).Updates(updates).Error
}
