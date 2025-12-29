package router

import (
	"github.com/Keba777/levpay-backend/feature/wallet"
	"github.com/Keba777/levpay-backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupWalletRoutes sets up routes for Wallet service
func SetupWalletRoutes(api fiber.Router, db *gorm.DB) {
	repo := wallet.NewRepository(db)
	handler := wallet.NewHandler(repo)

	walletGroup := api.Group("/wallet")

	// Apply JWT Middleware to all wallet routes
	walletGroup.Use(middleware.JWTMiddleware(db))

	// User Endpoints
	walletGroup.Get("/balance", handler.GetBalance)
	walletGroup.Post("/topup", handler.TopUp)
	walletGroup.Post("/withdraw", handler.Withdraw)
	walletGroup.Post("/lock", handler.LockWallet)
	walletGroup.Post("/unlock", handler.UnlockWallet)
}
