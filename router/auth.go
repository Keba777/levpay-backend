package router

import (
	"github.com/Keba777/levpay-backend/feature/auth"
	"github.com/Keba777/levpay-backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupAuthRoutes configures all authentication routes
func SetupAuthRoutes(app *fiber.App, db *gorm.DB) {
	authHandler := auth.NewHandler(db)

	// Auth routes group
	authRoutes := app.Group("/api/auth")

	// Public routes (no authentication required)
	authRoutes.Post("/register", authHandler.Register)
	authRoutes.Post("/login", authHandler.Login)
	authRoutes.Post("/refresh", authHandler.Refresh)
	authRoutes.Post("/logout", authHandler.Logout)
	authRoutes.Post("/forgot-password", authHandler.ForgotPassword)
	authRoutes.Post("/reset-password", authHandler.ResetPassword)
	authRoutes.Post("/google", authHandler.GoogleAuth)

	// Protected routes (require authentication)
	authRoutes.Get("/me", middleware.JWTMiddleware(db), authHandler.GetMe)
	authRoutes.Post("/change-password", middleware.JWTMiddleware(db), authHandler.ChangePassword)
}
