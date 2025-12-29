package router

import (
	"github.com/Keba777/levpay-backend/feature/billing"
	"github.com/Keba777/levpay-backend/feature/transaction"
	"github.com/Keba777/levpay-backend/feature/wallet"
	"github.com/Keba777/levpay-backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupBillingRoutes sets up routes for Billing service
func SetupBillingRoutes(api fiber.Router, db *gorm.DB) {
	billingRepo := billing.NewRepository(db)
	txRepo := transaction.NewRepository(db)
	walletRepo := wallet.NewRepository(db)
	handler := billing.NewHandler(billingRepo, txRepo, walletRepo, db)

	billingGroup := api.Group("/billing")

	// Apply JWT Middleware to all billing routes
	billingGroup.Use(middleware.JWTMiddleware(db))

	// Invoice Endpoints
	billingGroup.Post("/invoices", handler.CreateInvoice)
	billingGroup.Get("/invoices", handler.ListInvoices)
	billingGroup.Get("/invoices/:id", handler.GetInvoice)
	billingGroup.Post("/invoices/:id/pay", handler.PayInvoice)
	billingGroup.Put("/invoices/:id/cancel", handler.CancelInvoice)
	billingGroup.Get("/stats", handler.GetInvoiceStats)
}
