package router

import (
	"github.com/Keba777/levpay-backend/feature/kyc"
	"github.com/Keba777/levpay-backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupKYCRoutes sets up routes for KYC service
func SetupKYCRoutes(api fiber.Router, db *gorm.DB) {
	repo := kyc.NewRepository(db)
	handler := kyc.NewHandler(repo)

	kycGroup := api.Group("/kyc")

	// Apply JWT Middleware to all KYC routes
	kycGroup.Use(middleware.JWTMiddleware(db))

	// User Endpoints
	kycGroup.Post("/upload", handler.UploadDocument)
	kycGroup.Get("/status", handler.GetStatus)

	// Admin Endpoints
	admin := kycGroup.Group("/admin")
	admin.Use(middleware.RequireRole("admin"))

	admin.Get("/pending", handler.ListPending)
	admin.Post("/review/:id", handler.ReviewDocument)
}
