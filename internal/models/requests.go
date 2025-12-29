package models

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ==================== Authentication Requests ====================

// RegisterRequest for user registration
type RegisterRequest struct {
	Email       string `json:"email" binding:"required"`
	Password    string `json:"password" binding:"required"`
	Fingerprint string `json:"fingerprint" binding:"required"`
	FirstName   string `json:"first_name" binding:"required"`
	LastName    string `json:"last_name" binding:"required"`
	Phone       string `json:"phone" binding:"required"`
}

// LoginRequest for user login
type LoginRequest struct {
	Email       string `json:"email" binding:"required"`
	Password    string `json:"password" binding:"required"`
	Fingerprint string `json:"fingerprint" binding:"required"`
}

// RefreshRequest for refreshing access token
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	Fingerprint  string `json:"fingerprint" binding:"required"`
}

// LogoutRequest for user logout
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	Fingerprint  string `json:"fingerprint" binding:"required"`
}

// ForgotPasswordRequest for password reset email
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required"`
}

// ResetPasswordRequest for resetting password
type ResetPasswordRequest struct {
	NewPassword     string `json:"new_password" binding:"required"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
	Fingerprint     string `json:"fingerprint" binding:"required"`
}

// ChangePasswordRequest for changing password
type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
	Fingerprint     string `json:"fingerprint" binding:"required"`
	RefreshToken    string `json:"refresh_token" binding:"required"`
}

// ==================== User Requests ====================

// UpdateUserRequest for updating user profile
// UpdateUserRequest for updating user profile
type UpdateUserRequest struct {
	FirstName string  `json:"first_name,omitempty"`
	LastName  string  `json:"last_name,omitempty"`
	Phone     string  `json:"phone,omitempty"`
	Address   *string `json:"address,omitempty"`
	Username  *string `json:"username,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// UpdateSettingsRequest for updating user preferences
type UpdateSettingsRequest struct {
	Currency      string `json:"currency,omitempty"`
	Language      string `json:"language,omitempty"`
	Notifications *bool  `json:"notifications,omitempty"`
}

// ==================== Wallet Requests ====================

// TopUpWalletRequest for adding funds to wallet
type TopUpWalletRequest struct {
	Amount        float64 `json:"amount" binding:"required"`
	Currency      string  `json:"currency"`
	PaymentMethod string  `json:"payment_method" binding:"required"` // bank, card, mobile_wallet
}

// WithdrawRequest for withdrawing funds
type WithdrawRequest struct {
	Amount        float64 `json:"amount" binding:"required"`
	Currency      string  `json:"currency"`
	PaymentMethod string  `json:"payment_method" binding:"required"`
}

// ==================== Transaction Requests ====================

// TransferRequest for P2P transfers
type TransferRequest struct {
	ToUserID    uuid.UUID `json:"to_user_id" binding:"required"`
	Amount      float64   `json:"amount" binding:"required"`
	Currency    string    `json:"currency"`
	Description *string   `json:"description,omitempty"`
}

// PaymentRequest for merchant payments
type PaymentRequest struct {
	MerchantID  uuid.UUID  `json:"merchant_id" binding:"required"`
	InvoiceID   *uuid.UUID `json:"invoice_id,omitempty"`
	Amount      float64    `json:"amount" binding:"required"`
	Currency    string     `json:"currency"`
	Description *string    `json:"description,omitempty"`
}

// ==================== Payment Method Requests ====================

// AddPaymentMethodRequest for linking payment methods
type AddPaymentMethodRequest struct {
	Type      string                 `json:"type" binding:"required"` // bank, card, mobile_wallet
	Details   map[string]interface{} `json:"details" binding:"required"`
	IsDefault bool                   `json:"is_default"`
}

// ==================== Invoice Requests ====================

// CreateInvoiceRequest for merchants to create invoices
type CreateInvoiceRequest struct {
	Amount   float64 `json:"amount" binding:"required"`
	Currency string  `json:"currency"`
	DueDate  *string `json:"due_date,omitempty"` // ISO 8601 format
}

// ==================== Pagination and Listing ====================

// ListedRequest is a helper for paginated list requests
type ListedRequest struct {
	Page     int    `json:"page"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
	Keywords string `json:"keywords"`
	OrderBy  string `json:"order_by"`
	Order    string `json:"order"`
}

// FromContext parses pagination params from Fiber context
func (lr *ListedRequest) FromContext(c *fiber.Ctx) {
	log.Printf("[ListedRequest.FromContext] Called for path: %s", c.Path())
	page := c.Query("page")
	limit := c.Query("limit")
	if page == "" {
		page = "1"
	}
	if limit == "" {
		limit = "10"
	}
	lr.Page, _ = strconv.Atoi(page)
	if lr.Page <= 0 {
		lr.Page = 1
	}
	lr.Limit, _ = strconv.Atoi(limit)
	if lr.Limit <= 0 {
		lr.Limit = 10
	}
	lr.Offset = (lr.Page - 1) * lr.Limit
	lr.Keywords = c.Query("keywords")
	lr.OrderBy = c.Query("order_by")
	if lr.OrderBy == "" {
		lr.OrderBy = "created_at"
	}
	lr.Order = c.Query("order")
	if lr.Order != "asc" && lr.Order != "desc" {
		lr.Order = "desc"
	}
	log.Printf("[ListedRequest.FromContext] Parsed: Page=%d, Limit=%d, Offset=%d, Keywords=%s, OrderBy=%s, Order=%s",
		lr.Page, lr.Limit, lr.Offset, lr.Keywords, lr.OrderBy, lr.Order)
}

// Getter methods for ListedRequest
func (lr *ListedRequest) GetPage() int {
	return lr.Page
}

func (lr *ListedRequest) GetLimit() int {
	return lr.Limit
}

func (lr *ListedRequest) GetOffset() int {
	return lr.Offset
}

func (lr *ListedRequest) GetKeywords() string {
	return lr.Keywords
}

func (lr *ListedRequest) GetOrderBy() string {
	return lr.OrderBy
}

func (lr *ListedRequest) GetOrder() string {
	return lr.Order
}
