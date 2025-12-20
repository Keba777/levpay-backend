package middleware

import (
	"fmt"
	"strings"

	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/Keba777/levpay-backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// JWTMiddleware verifies JWT access tokens and loads user into context
func JWTMiddleware(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(models.InfoResponse{
				Message: "Authorization header required",
			})
		}

		// Remove "Bearer " prefix
		authHeader = strings.TrimSpace(authHeader)
		authHeader = strings.TrimPrefix(authHeader, "Bearer ")

		// Decode and validate token
		decodedToken, err := utils.JWT.DecodeToken(authHeader, models.JWTAccess)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(models.InfoResponse{
				Message: "Invalid or expired token",
			})
		}

		// Load user from database
		var user models.User
		if err := db.Preload("Sessions").Preload("Wallet").First(&user, "id = ?", decodedToken.UserID).Error; err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(models.InfoResponse{
				Message: "User not found",
			})
		}

		// Store user and token in context
		c.Locals("user", user)
		c.Locals("decoded_token", decodedToken)

		return c.Next()
	}
}

// RequireRole ensures the user has the specified role
func RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(models.InfoResponse{
				Message: "Unauthorized",
			})
		}

		// Check if user's role is in allowed roles
		for _, role := range allowedRoles {
			if user.Role == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(models.InfoResponse{
			Message: fmt.Sprintf("Access denied. Required roles: %v", allowedRoles),
		})
	}
}

// OptionalJWT allows requests with or without JWT
func OptionalJWT(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		authHeader = strings.TrimSpace(authHeader)
		authHeader = strings.TrimPrefix(authHeader, "Bearer ")

		decodedToken, err := utils.JWT.DecodeToken(authHeader, models.JWTAccess)
		if err != nil {
			return c.Next()
		}

		var user models.User
		if err := db.Preload("Sessions").Preload("Wallet").First(&user, "id = ?", decodedToken.UserID).Error; err != nil {
			return c.Next()
		}

		c.Locals("user", user)
		c.Locals("decoded_token", decodedToken)

		return c.Next()
	}
}
