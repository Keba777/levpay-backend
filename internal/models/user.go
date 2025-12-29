package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	gorm.Model
	ID             uuid.UUID       `gorm:"type:uuid;default:gen_random_uuid();primaryKey"` // Use UUID as primary key
	FirstName      string          `gorm:"not null"`
	LastName       string          `gorm:"not null"`
	Username       *string         `gorm:"unique;index"` // Unique handle (e.g. @kaybee)
	AvatarURL      *string         // Profile picture URL
	Email          string          `gorm:"unique;not null"`
	Phone          *string         `gorm:"unique"` // Nullable for OAuth users
	PasswordHash   *string         // Nullable to support OAuth-only accounts
	GoogleID       *string         `gorm:"unique"` // For tracking Google accounts
	DOB            *time.Time      // Optional
	Address        *string         // Optional
	Preferences    *string         `gorm:"type:jsonb"`        // JSON string for user settings {currency, language, notifications}
	KYCStatus      string          `gorm:"default:'pending'"` // Enum: pending, verified, rejected
	Role           string          `gorm:"default:'user'"`    // Enum: user, merchant, admin
	TwoFAEnabled   bool            `gorm:"default:false"`
	Wallet         Wallet          `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Transactions   []Transaction   `gorm:"foreignKey:FromUserID"`
	KYCDocuments   []KYCDocument   `gorm:"foreignKey:UserID"`
	PaymentMethods []PaymentMethod `gorm:"foreignKey:UserID"`
	Sessions       []Session       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// PrepareResponse sanitizes user data for API responses
// PrepareResponse sanitizes user data for API responses
func (u *User) PrepareResponse() UserResponse {
	phone := ""
	if u.Phone != nil {
		phone = *u.Phone
	}
	username := ""
	if u.Username != nil {
		username = *u.Username
	}
	avatar := ""
	if u.AvatarURL != nil {
		avatar = *u.AvatarURL
	}
	prefs := "{}"
	if u.Preferences != nil {
		prefs = *u.Preferences
	}

	return UserResponse{
		ID:           u.ID,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Username:     username,
		AvatarURL:    avatar,
		Email:        u.Email,
		Phone:        phone,
		Preferences:  prefs,
		KYCStatus:    u.KYCStatus,
		Role:         u.Role,
		Is2FAEnabled: u.TwoFAEnabled,
		CreatedAt:    u.CreatedAt,
	}
}
