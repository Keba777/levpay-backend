package router

import (
	"github.com/Keba777/levpay-backend/feature/notification"
	"github.com/Keba777/levpay-backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupNotificationRoutes sets up routes for Notification service
func SetupNotificationRoutes(api fiber.Router, db *gorm.DB) {
	repo := notification.NewRepository(db)
	handler := notification.NewHandler(repo)

	notifGroup := api.Group("/notifications")

	// Apply JWT Middleware to all notification routes
	notifGroup.Use(middleware.JWTMiddleware(db))

	// User Endpoints
	notifGroup.Get("/", handler.GetNotifications)
	notifGroup.Put("/:id/read", handler.MarkAsRead)
	notifGroup.Get("/unread-count", handler.GetUnreadCount)
}
