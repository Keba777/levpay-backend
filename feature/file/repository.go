package file

import (
	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository handles file-related database operations
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new file repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// CreateFile records a new file upload
func (r *Repository) CreateFile(file *models.File) error {
	return r.db.Create(file).Error
}

// GetFileByID retrieves a file by ID
func (r *Repository) GetFileByID(id uuid.UUID) (*models.File, error) {
	var file models.File
	if err := r.db.Where("id = ?", id).First(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

// GetUserFiles retrieves all files for a user with optional category filter and pagination
func (r *Repository) GetUserFiles(userID uuid.UUID, category string, req models.ListedRequest) ([]models.File, int64, error) {
	var files []models.File
	var total int64

	query := r.db.Model(&models.File{}).Where("user_id = ?", userID)

	// Apply category filter if provided
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and ordering
	if err := query.
		Order(req.OrderBy + " " + req.Order).
		Limit(req.Limit).
		Offset(req.Offset).
		Find(&files).Error; err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

// DeleteFile removes a file record from database
func (r *Repository) DeleteFile(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.File{}).Error
}

// GetFilesByCategory retrieves files by category
func (r *Repository) GetFilesByCategory(userID uuid.UUID, category string) ([]models.File, error) {
	var files []models.File
	err := r.db.Where("user_id = ? AND category = ?", userID, category).
		Order("uploaded_at desc").
		Find(&files).Error
	return files, err
}
