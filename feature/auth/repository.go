package auth

import (
	"fmt"
	"time"

	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/Keba777/levpay-backend/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository handles database operations for authentication
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new auth repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// CreateUser creates a new user with hashed password
func (r *Repository) CreateUser(req models.RegisterRequest) (*models.User, error) {
	// Hash password
	hashedPassword, err := utils.PWD.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: hashedPassword,
		Role:         "user", // Default role
		KYCStatus:    "pending",
	}

	if err := r.db.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create default wallet for user
	wallet := &models.Wallet{
		UserID:      user.ID,
		Currency:    "ETB",
		Balance:     0,
		LastUpdated: time.Now(),
	}

	if err := r.db.Create(wallet).Error; err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	return user, nil
}

// GetUserByEmail finds a user by email
func (r *Repository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Preload("Wallet").Preload("Sessions").First(&user, "email = ?", email).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID finds a user by ID
func (r *Repository) GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.Preload("Wallet").Preload("Sessions").First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdatePassword updates a user's password
func (r *Repository) UpdatePassword(userID uuid.UUID, newPassword string) error {
	hashedPassword, err := utils.PWD.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := r.db.Model(&models.User{}).Where("id = ?", userID).Update("password_hash", hashedPassword).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// CreateSession creates a new session with refresh token
func (r *Repository) CreateSession(userID uuid.UUID, refreshToken, fingerprint string) (*models.Session, error) {
	session := &models.Session{
		UserID:       userID,
		RefreshToken: refreshToken,
		Fingerprint:  fingerprint,
		Active:       true,
	}

	if err := r.db.Create(session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// GetSessionByToken finds an active session by refresh token
func (r *Repository) GetSessionByToken(refreshToken string) (*models.Session, error) {
	var session models.Session
	if err := r.db.First(&session, "refresh_token = ? AND active = ?", refreshToken, true).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

// TerminateSession marks a session as terminated
func (r *Repository) TerminateSession(sessionID uuid.UUID) error {
	now := time.Now()
	if err := r.db.Model(&models.Session{}).
		Where("id = ?", sessionID).
		Updates(map[string]interface{}{
			"active":     false,
			"terminated": now,
		}).Error; err != nil {
		return fmt.Errorf("failed to terminate session: %w", err)
	}
	return nil
}

// TerminateUserSessions terminates all sessions for a user
func (r *Repository) TerminateUserSessions(userID uuid.UUID) error {
	now := time.Now()
	if err := r.db.Model(&models.Session{}).
		Where("user_id = ? AND active = ?", userID, true).
		Updates(map[string]interface{}{
			"active":     false,
			"terminated": now,
		}).Error; err != nil {
		return fmt.Errorf("failed to terminate sessions: %w", err)
	}
	return nil
}

// ValidateCredentials checks if email and password match
func (r *Repository) ValidateCredentials(email, password string) (*models.User, error) {
	user, err := r.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if !utils.PWD.CheckPasswordHash(password, user.PasswordHash) {
		return nil, fmt.Errorf("invalid credentials")
	}

	return user, nil
}
