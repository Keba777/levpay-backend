package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// KYC Status Constants
const (
	KYCStatusPending  = "pending"
	KYCStatusApproved = "approved"
	KYCStatusRejected = "rejected"
)

// KYC Document Type Constants
const (
	KYCDocTypeIDCard         = "id_card"
	KYCDocTypePassport       = "passport"
	KYCDocTypeDriverLicense  = "driver_license"
	KYCDocTypeProofOfAddress = "proof_of_address"
)

// KYCDocument represents a user's KYC document
type KYCDocument struct {
	gorm.Model
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID     uuid.UUID `gorm:"not null;type:uuid"`
	Type       string    `gorm:"not null"` // e.g., id_card, passport
	FilePath   string    `gorm:"not null"` // MinIO path
	Status     string    `gorm:"default:'pending'"`
	Notes      string    `gorm:"type:text"` // Admin reject reasons etc
	UploadedAt time.Time
}

// SubmitKYCRequest struct for multipart form data binding hints
// Note: File is handled via FormFile in handler
type SubmitKYCRequest struct {
	Type string `form:"type" binding:"required"`
}

// ReviewKYCRequest struct for admin review
type ReviewKYCRequest struct {
	Status string `json:"status" binding:"required"` // approved, rejected
	Notes  string `json:"notes"`
}

// KYCStatusResponse for standardized status return
type KYCStatusResponse struct {
	OverallStatus string        `json:"overall_status"`
	Documents     []KYCDocument `json:"documents"`
}
