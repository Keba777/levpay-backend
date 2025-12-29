package router

import (
	"github.com/Keba777/levpay-backend/feature/user"
	"github.com/Keba777/levpay-backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupUserRoutes sets up routes for user service
func SetupUserRoutes(api fiber.Router, db *gorm.DB) {
	repo := user.NewRepository(db)
	handler := user.NewHandler(repo)

	users := api.Group("/users")

	// Protected routes (require valid JWT)
	users.Use(middleware.JWTMiddleware(db))

	// Profile management
	users.Get("/me", handler.GetProfile)
	users.Put("/me", handler.UpdateProfile)
	users.Get("/search", handler.SearchUsers)

	// Settings
	users.Get("/settings", handler.GetSettings)
	users.Put("/settings", handler.UpdateSettings)

	// Admin only routes
	admin := users.Group("/")
	admin.Use(middleware.RequireRole("admin"))

	admin.Get("/", handler.ListUsers)
	admin.Patch("/:id/kyc", handler.UpdateKYC)
}
