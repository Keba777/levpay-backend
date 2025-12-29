package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// File Category Constants
const (
	FileCategoryKYC      = "kyc"
	FileCategoryAvatar   = "avatar"
	FileCategoryDocument = "document"
	FileCategoryInvoice  = "invoice"
	FileCategoryReceipt  = "receipt"
	FileCategoryOther    = "other"
)

// File represents a file uploaded by a user
type File struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID      uuid.UUID `gorm:"not null;type:uuid;index"`
	FileName    string    `gorm:"not null"`
	FilePath    string    `gorm:"not null"`       // MinIO object path
	FileType    string    `gorm:"not null"`       // MIME type
	FileSize    int64     `gorm:"not null"`       // Size in bytes
	Category    string    `gorm:"not null;index"` // kyc, avatar, document, etc.
	Description *string   // Optional description
	UploadedAt  time.Time
}

// FileResponse for API responses
type FileResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	FileName    string    `json:"file_name"`
	FileType    string    `json:"file_type"`
	FileSize    int64     `json:"file_size"`
	Category    string    `json:"category"`
	Description *string   `json:"description,omitempty"`
	UploadedAt  time.Time `json:"uploaded_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// ToResponse converts file to API response format
func (f *File) ToResponse() FileResponse {
	return FileResponse{
		ID:          f.ID,
		UserID:      f.UserID,
		FileName:    f.FileName,
		FileType:    f.FileType,
		FileSize:    f.FileSize,
		Category:    f.Category,
		Description: f.Description,
		UploadedAt:  f.UploadedAt,
		CreatedAt:   f.CreatedAt,
	}
}
