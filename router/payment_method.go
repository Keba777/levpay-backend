package router

import (
	"github.com/Keba777/levpay-backend/feature/payment_method"
	"github.com/Keba777/levpay-backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupPaymentMethodRoutes(api fiber.Router, db *gorm.DB) {
	repo := payment_method.NewRepository(db)
	handler := payment_method.NewHandler(repo)

	pm := api.Group("/payment-methods")
	pm.Use(middleware.JWTMiddleware(db))

	pm.Get("/", handler.ListPaymentMethods)
	pm.Post("/", handler.AddPaymentMethod)
	pm.Delete("/:id", handler.RemovePaymentMethod)
	pm.Patch("/:id/default", handler.SetDefaultPaymentMethod)
}
