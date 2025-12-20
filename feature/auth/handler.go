package auth

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/Keba777/levpay-backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Handler handles authentication HTTP requests
type Handler struct {
	repo          *Repository
	accessExpiry  int
	refreshExpiry int
	forgotExpiry  int
}

// NewHandler creates a new auth handler
func NewHandler(db *gorm.DB) *Handler {
	// Get token expiry times from environment (with defaults)
	accessExpiry := 900     // 15 minutes
	refreshExpiry := 604800 // 7 days
	forgotExpiry := 172800  // 48 hours

	return &Handler{
		repo:          NewRepository(db),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
		forgotExpiry:  forgotExpiry,
	}
}

// Register handles user registration
func (h *Handler) Register(c *fiber.Ctx) error {
	var req models.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.InfoResponse{
			Message: "Invalid request body",
		})
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" || req.Phone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.InfoResponse{
			Message: "All fields are required",
		})
	}

	// Check if user already exists
	if _, err := h.repo.GetUserByEmail(req.Email); err == nil {
		return c.Status(fiber.StatusConflict).JSON(models.InfoResponse{
			Message: "User with this email already exists",
		})
	}

	// Create user
	user, err := h.repo.CreateUser(req)
	if err != nil {
		log.Printf("[Register] Failed to create user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.InfoResponse{
			Message: "Failed to create user",
		})
	}

	// Generate forgot password token for email verification/password setup
	forgotToken, err := utils.JWT.GenerateToken(user.ID, user.Role, h.forgotExpiry, models.JWTForgot)
	if err != nil {
		log.Printf("[Register] Failed to generate forgot token: %v", err)
	} else {
		// TODO: Send welcome email with password setup link
		setupLink := fmt.Sprintf("%s/auth/reset-password?token=%s", os.Getenv("FRONTEND_URL"), forgotToken)
		log.Printf("[Register] Welcome email link for %s: %s", user.Email, setupLink)
	}

	return c.Status(fiber.StatusCreated).JSON(models.InfoResponse{
		Message: "User registered successfully. Please check your email to verify your account.",
	})
}

// Login handles user login
func (h *Handler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.InfoResponse{
			Message: "Invalid request body",
		})
	}

	// Validate credentials
	user, err := h.repo.ValidateCredentials(req.Email, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.InfoResponse{
			Message: "Invalid email or password",
		})
	}

	// Generate tokens
	accessToken, err := utils.JWT.GenerateToken(user.ID, user.Role, h.accessExpiry, models.JWTAccess)
	if err != nil {
		log.Printf("[Login] Failed to generate access token: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.InfoResponse{
			Message: "Failed to generate token",
		})
	}

	refreshToken, err := utils.JWT.GenerateToken(user.ID, user.Role, h.refreshExpiry, models.JWTRefresh)
	if err != nil {
		log.Printf("[Login] Failed to generate refresh token: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.InfoResponse{
			Message: "Failed to generate token",
		})
	}

	// Create session
	_, err = h.repo.CreateSession(user.ID, refreshToken, req.Fingerprint)
	if err != nil {
		log.Printf("[Login] Failed to create session: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.InfoResponse{
			Message: "Failed to create session",
		})
	}

	// Update last login time
	now := time.Now()
	user.UpdatedAt = now

	return c.JSON(models.LoggedInUserResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user.PrepareResponse(),
	})
}

// Refresh handles token refresh
func (h *Handler) Refresh(c *fiber.Ctx) error {
	var req models.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.InfoResponse{
			Message: "Invalid request body",
		})
	}

	// Validate refresh token
	decodedToken, err := utils.JWT.DecodeToken(req.RefreshToken, models.JWTRefresh)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.InfoResponse{
			Message: "Invalid or expired refresh token",
		})
	}

	// Check session exists and is active
	session, err := h.repo.GetSessionByToken(req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.InfoResponse{
			Message: "Invalid session",
		})
	}

	// Verify fingerprint matches
	if session.Fingerprint != req.Fingerprint {
		log.Printf("[Refresh] Fingerprint mismatch for user %s", decodedToken.UserID)
		return c.Status(fiber.StatusUnauthorized).JSON(models.InfoResponse{
			Message: "Invalid session",
		})
	}

	// Get user
	user, err := h.repo.GetUserByID(decodedToken.UserID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.InfoResponse{
			Message: "User not found",
		})
	}

	// Generate new tokens (token rotation)
	newAccessToken, err := utils.JWT.GenerateToken(user.ID, user.Role, h.accessExpiry, models.JWTAccess)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.InfoResponse{
			Message: "Failed to generate token",
		})
	}

	newRefreshToken, err := utils.JWT.GenerateToken(user.ID, user.Role, h.refreshExpiry, models.JWTRefresh)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.InfoResponse{
			Message: "Failed to generate token",
		})
	}

	// Terminate old session
	if err := h.repo.TerminateSession(session.ID); err != nil {
		log.Printf("[Refresh] Failed to terminate old session: %v", err)
	}

	// Create new session
	_, err = h.repo.CreateSession(user.ID, newRefreshToken, req.Fingerprint)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.InfoResponse{
			Message: "Failed to create session",
		})
	}

	return c.JSON(models.LoggedInUserResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		User:         user.PrepareResponse(),
	})
}

// Logout handles user logout
func (h *Handler) Logout(c *fiber.Ctx) error {
	var req models.LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.InfoResponse{
			Message: "Invalid request body",
		})
	}

	// Validate refresh token
	_, err := utils.JWT.DecodeToken(req.RefreshToken, models.JWTRefresh)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.InfoResponse{
			Message: "Invalid refresh token",
		})
	}

	// Get session
	session, err := h.repo.GetSessionByToken(req.RefreshToken)
	if err != nil {
		// Already logged out or invalid session
		return c.JSON(models.InfoResponse{
			Message: "Logged out successfully",
		})
	}

	// Terminate session
	if err := h.repo.TerminateSession(session.ID); err != nil {
		log.Printf("[Logout] Failed to terminate session: %v", err)
	}

	return c.JSON(models.InfoResponse{
		Message: "Logged out successfully",
	})
}

// ForgotPassword handles password reset request
func (h *Handler) ForgotPassword(c *fiber.Ctx) error {
	var req models.ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.InfoResponse{
			Message: "Invalid request body",
		})
	}

	// Get user by email
	user, err := h.repo.GetUserByEmail(req.Email)
	if err != nil {
		// Don't reveal if email exists
		return c.JSON(models.InfoResponse{
			Message: "If the email exists, a password reset link has been sent",
		})
	}

	// Generate forgot password token
	forgotToken, err := utils.JWT.GenerateToken(user.ID, user.Role, h.forgotExpiry, models.JWTForgot)
	if err != nil {
		log.Printf("[ForgotPassword] Failed to generate forgot token: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.InfoResponse{
			Message: "Failed to process request",
		})
	}

	// TODO: Send password reset email
	resetLink := fmt.Sprintf("%s/auth/reset-password?token=%s", os.Getenv("FRONTEND_URL"), forgotToken)
	log.Printf("[ForgotPassword] Reset link for %s: %s", user.Email, resetLink)

	return c.JSON(models.InfoResponse{
		Message: "If the email exists, a password reset link has been sent",
	})
}

// ResetPassword handles password reset
func (h *Handler) ResetPassword(c *fiber.Ctx) error {
	var req models.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.InfoResponse{
			Message: "Invalid request body",
		})
	}

	// Validate passwords match
	if req.NewPassword != req.ConfirmPassword {
		return c.Status(fiber.StatusBadRequest).JSON(models.InfoResponse{
			Message: "Passwords do not match",
		})
	}

	// Get token from query
	token := c.Query("token")
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.InfoResponse{
			Message: "Reset token required",
		})
	}

	// Decode forgot password token
	decodedToken, err := utils.JWT.DecodeToken(token, models.JWTForgot)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.InfoResponse{
			Message: "Invalid or expired reset token",
		})
	}

	// Update password
	if err := h.repo.UpdatePassword(decodedToken.UserID, req.NewPassword); err != nil {
		log.Printf("[ResetPassword] Failed to update password: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.InfoResponse{
			Message: "Failed to reset password",
		})
	}

	// Terminate all sessions for security
	if err := h.repo.TerminateUserSessions(decodedToken.UserID); err != nil {
		log.Printf("[ResetPassword] Failed to terminate sessions: %v", err)
	}

	return c.JSON(models.InfoResponse{
		Message: "Password reset successfully",
	})
}

// ChangePassword handles password change for authenticated users
func (h *Handler) ChangePassword(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(models.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.InfoResponse{
			Message: "Unauthorized",
		})
	}

	var req models.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.InfoResponse{
			Message: "Invalid request body",
		})
	}

	// Validate passwords match
	if req.NewPassword != req.ConfirmPassword {
		return c.Status(fiber.StatusBadRequest).JSON(models.InfoResponse{
			Message: "Passwords do not match",
		})
	}

	// Verify old password
	if !utils.PWD.CheckPasswordHash(req.OldPassword, user.PasswordHash) {
		return c.Status(fiber.StatusUnauthorized).JSON(models.InfoResponse{
			Message: "Invalid old password",
		})
	}

	// Update password
	if err := h.repo.UpdatePassword(user.ID, req.NewPassword); err != nil {
		log.Printf("[ChangePassword] Failed to update password: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.InfoResponse{
			Message: "Failed to change password",
		})
	}

	// Terminate all other sessions
	if err := h.repo.TerminateUserSessions(user.ID); err != nil {
		log.Printf("[ChangePassword] Failed to terminate sessions: %v", err)
	}

	// Generate new tokens for current session
	newAccessToken, _ := utils.JWT.GenerateToken(user.ID, user.Role, h.accessExpiry, models.JWTAccess)
	newRefreshToken, _ := utils.JWT.GenerateToken(user.ID, user.Role, h.refreshExpiry, models.JWTRefresh)

	// Create new session
	h.repo.CreateSession(user.ID, newRefreshToken, req.Fingerprint)

	return c.JSON(models.LoggedInUserResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		User:         user.PrepareResponse(),
	})
}

// GetMe returns current authenticated user
func (h *Handler) GetMe(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(models.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.InfoResponse{
			Message: "Unauthorized",
		})
	}

	return c.JSON(user.PrepareResponse())
}
