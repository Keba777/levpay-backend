package router

import (
	"github.com/Keba777/levpay-backend/feature/admin"
	"github.com/Keba777/levpay-backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupAdminRoutes sets up routes for the Admin service
func SetupAdminRoutes(api fiber.Router, db *gorm.DB) {
	repo := admin.NewRepository(db)
	handler := admin.NewHandler(repo)

	adminGroup := api.Group("/admin")

	// Apply JWT and Admin Role Middleware
	adminGroup.Use(middleware.JWTMiddleware(db))
	adminGroup.Use(middleware.RequireRole("admin"))

	// Dashboard
	adminGroup.Get("/dashboard", handler.GetDashboard)

	// User Management
	adminGroup.Get("/users", handler.ListUsers)
	adminGroup.Patch("/users/:id/status", handler.UpdateUserStatus)

	// Audit Logs
	adminGroup.Get("/audit-logs", handler.GetAuditLogs)
}
