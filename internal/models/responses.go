package models

import (
	"time"

	"github.com/google/uuid"
)

// ==================== Common Responses ====================

// InfoResponse for simple message responses
type InfoResponse struct {
	Message string `json:"message"`
}

// ListedResponse for paginated list responses
type ListedResponse struct {
	Records []interface{} `json:"records"`
	Total   int           `json:"total"`
	Page    int           `json:"page"`
	Limit   int           `json:"limit"`
}

// ==================== Authentication Responses ====================

// LoggedInUserResponse returned after successful login
type LoggedInUserResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

// UserResponse sanitized user data for API responses
type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Username    string    `json:"username"`
	AvatarURL   string    `json:"avatar_url"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Preferences string    `json:"preferences"` // JSON string
	KYCStatus   string    `json:"kyc_status"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
}

// ==================== Wallet Responses ====================

// WalletResponse for wallet data
type WalletResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Balance     float64   `json:"balance"`
	Currency    string    `json:"currency"`
	Locked      bool      `json:"locked"`
	LastUpdated time.Time `json:"last_updated"`
}

// BalanceResponse for quick balance checks
type BalanceResponse struct {
	Balance  float64 `json:"balance"`
	Currency string  `json:"currency"`
}

// ==================== Transaction Responses ====================

// TransactionResponse for transaction data
type TransactionResponse struct {
	ID          uuid.UUID  `json:"id"`
	FromUserID  uuid.UUID  `json:"from_user_id"`
	ToUserID    *uuid.UUID `json:"to_user_id,omitempty"`
	Amount      float64    `json:"amount"`
	Currency    string     `json:"currency"`
	Type        string     `json:"type"`
	Status      string     `json:"status"`
	Description *string    `json:"description,omitempty"`
	Fee         float64    `json:"fee"`
	CreatedAt   time.Time  `json:"created_at"`
}

// ==================== KYC Responses ====================

// KYCDocumentResponse for KYC document data
type KYCDocumentResponse struct {
	ID         uuid.UUID `json:"id"`
	Type       string    `json:"type"`
	Status     string    `json:"status"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// ==================== Payment Method Responses ====================

// PaymentMethodResponse for payment method data
type PaymentMethodResponse struct {
	ID        uuid.UUID `json:"id"`
	Type      string    `json:"type"`
	IsDefault bool      `json:"is_default"`
	Verified  bool      `json:"verified"`
	// Details are intentionally omitted for security
	LastFourDigits string `json:"last_four_digits,omitempty"` // For cards
}
