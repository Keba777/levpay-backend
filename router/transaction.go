package router

import (
	"github.com/Keba777/levpay-backend/feature/transaction"
	"github.com/Keba777/levpay-backend/feature/wallet"
	"github.com/Keba777/levpay-backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupTransactionRoutes sets up routes for Transaction service
func SetupTransactionRoutes(api fiber.Router, db *gorm.DB) {
	txRepo := transaction.NewRepository(db)
	walletRepo := wallet.NewRepository(db)
	handler := transaction.NewHandler(txRepo, walletRepo, db)

	txGroup := api.Group("/transaction")

	// Apply JWT Middleware to all transaction routes
	txGroup.Use(middleware.JWTMiddleware(db))

	// User Endpoints
	txGroup.Post("/transfer", handler.Transfer)
	txGroup.Post("/payment", handler.Payment)
	txGroup.Get("/history", handler.GetHistory)
	txGroup.Get("/:id", handler.GetTransactionDetails)
}
