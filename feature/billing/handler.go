package billing

import (
	"time"

	"github.com/Keba777/levpay-backend/feature/transaction"
	"github.com/Keba777/levpay-backend/feature/wallet"
	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Handler handles billing HTTP requests
type Handler struct {
	repo       *Repository
	txRepo     *transaction.Repository
	walletRepo *wallet.Repository
	db         *gorm.DB
}

// NewHandler creates a new billing handler
func NewHandler(repo *Repository, txRepo *transaction.Repository, walletRepo *wallet.Repository, db *gorm.DB) *Handler {
	return &Handler{
		repo:       repo,
		txRepo:     txRepo,
		walletRepo: walletRepo,
		db:         db,
	}
}

// Helper to get userID from context
func getUserID(c *fiber.Ctx) (uuid.UUID, error) {
	user, ok := c.Locals("user").(models.User)
	if !ok {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid user session")
	}
	return user.ID, nil
}

// CreateInvoice creates a new invoice (merchant only)
func (h *Handler) CreateInvoice(c *fiber.Ctx) error {
	merchantID, err := getUserID(c)
	if err != nil {
		return err
	}

	var req models.CreateInvoiceRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.Amount <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Amount must be positive")
	}

	// Parse due date if provided
	var dueDate *time.Time
	if req.DueDate != nil {
		parsed, err := time.Parse(time.RFC3339, *req.DueDate)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid due date format")
		}
		dueDate = &parsed
	}

	invoice := &models.Invoice{
		MerchantID: merchantID,
		Amount:     req.Amount,
		Currency:   req.Currency,
		Status:     models.InvoiceStatusDraft,
		DueDate:    dueDate,
	}

	if err := h.repo.CreateInvoice(invoice); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create invoice")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Invoice created successfully",
		"invoice": invoice.ToResponse(),
	})
}

// GetInvoice retrieves invoice details
func (h *Handler) GetInvoice(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	invoiceIDStr := c.Params("id")
	invoiceID, err := uuid.Parse(invoiceIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid invoice ID")
	}

	invoice, err := h.repo.GetInvoiceByID(invoiceID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Invoice not found")
	}

	// Verify access (merchant or customer)
	if invoice.MerchantID != userID && (invoice.CustomerID == nil || *invoice.CustomerID != userID) {
		return fiber.NewError(fiber.StatusForbidden, "Access denied")
	}

	return c.JSON(invoice.ToResponse())
}

// ListInvoices lists invoices for the authenticated user
func (h *Handler) ListInvoices(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	viewType := c.Query("view", "merchant") // merchant or customer
	var req models.ListedRequest
	req.FromContext(c)

	var invoices []models.Invoice
	var total int64

	if viewType == "customer" {
		invoices, total, err = h.repo.GetCustomerInvoices(userID, req)
	} else {
		invoices, total, err = h.repo.GetMerchantInvoices(userID, req)
	}

	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve invoices")
	}

	// Convert to responses
	var responses []models.InvoiceResponse
	for _, inv := range invoices {
		responses = append(responses, inv.ToResponse())
	}

	return c.JSON(models.ListedResponse{
		Records: func() []interface{} {
			result := make([]interface{}, len(responses))
			for i, v := range responses {
				result[i] = v
			}
			return result
		}(),
		Total: int(total),
		Page:  req.Page,
		Limit: req.Limit,
	})
}

// PayInvoice processes payment for an invoice
func (h *Handler) PayInvoice(c *fiber.Ctx) error {
	customerID, err := getUserID(c)
	if err != nil {
		return err
	}

	invoiceIDStr := c.Params("id")
	invoiceID, err := uuid.Parse(invoiceIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid invoice ID")
	}

	invoice, err := h.repo.GetInvoiceByID(invoiceID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Invoice not found")
	}

	// Check if already paid
	if invoice.Status == models.InvoiceStatusPaid {
		return fiber.NewError(fiber.StatusBadRequest, "Invoice already paid")
	}

	// Check if cancelled
	if invoice.Status == models.InvoiceStatusCancelled {
		return fiber.NewError(fiber.StatusBadRequest, "Invoice is cancelled")
	}

	// Process payment in transaction
	var txRecord *models.Transaction
	err = h.db.Transaction(func(tx *gorm.DB) error {
		// Deduct from customer
		if err := h.walletRepo.UpdateBalance(customerID, -invoice.Amount); err != nil {
			return err
		}

		// Credit to merchant
		if err := h.walletRepo.UpdateBalance(invoice.MerchantID, invoice.Amount); err != nil {
			return err
		}

		// Create transaction record
		txRecord = &models.Transaction{
			FromUserID: customerID,
			ToUserID:   &invoice.MerchantID,
			Amount:     invoice.Amount,
			Currency:   invoice.Currency,
			Type:       models.TransactionTypePayment,
			Status:     models.TransactionStatusCompleted,
		}

		if err := h.txRepo.CreateTransaction(txRecord); err != nil {
			return err
		}

		// Mark invoice as paid
		if err := h.repo.MarkInvoiceAsPaid(invoiceID, txRecord.ID); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if err.Error() == "insufficient balance" {
			return fiber.NewError(fiber.StatusBadRequest, "Insufficient balance")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Payment failed")
	}

	return c.JSON(fiber.Map{
		"message":     "Invoice paid successfully",
		"transaction": txRecord.ToResponse(),
	})
}

// CancelInvoice cancels an unpaid invoice
func (h *Handler) CancelInvoice(c *fiber.Ctx) error {
	merchantID, err := getUserID(c)
	if err != nil {
		return err
	}

	invoiceIDStr := c.Params("id")
	invoiceID, err := uuid.Parse(invoiceIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid invoice ID")
	}

	invoice, err := h.repo.GetInvoiceByID(invoiceID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Invoice not found")
	}

	// Verify ownership
	if invoice.MerchantID != merchantID {
		return fiber.NewError(fiber.StatusForbidden, "Access denied")
	}

	// Check if already paid
	if invoice.Status == models.InvoiceStatusPaid {
		return fiber.NewError(fiber.StatusBadRequest, "Cannot cancel paid invoice")
	}

	if err := h.repo.UpdateInvoiceStatus(invoiceID, models.InvoiceStatusCancelled); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to cancel invoice")
	}

	return c.JSON(fiber.Map{"message": "Invoice cancelled successfully"})
}

// GetInvoiceStats retrieves billing statistics
func (h *Handler) GetInvoiceStats(c *fiber.Ctx) error {
	merchantID, err := getUserID(c)
	if err != nil {
		return err
	}

	// Get all merchant invoices (simplified stats)
	var totalInvoices, paidInvoices, pendingInvoices int64
	var totalAmount, paidAmount float64

	h.db.Model(&models.Invoice{}).Where("merchant_id = ?", merchantID).Count(&totalInvoices)
	h.db.Model(&models.Invoice{}).Where("merchant_id = ? AND status = ?", merchantID, models.InvoiceStatusPaid).Count(&paidInvoices)
	h.db.Model(&models.Invoice{}).Where("merchant_id = ? AND status != ? AND status != ?", merchantID, models.InvoiceStatusPaid, models.InvoiceStatusCancelled).Count(&pendingInvoices)

	h.db.Model(&models.Invoice{}).Where("merchant_id = ?", merchantID).Select("COALESCE(SUM(amount), 0)").Scan(&totalAmount)
	h.db.Model(&models.Invoice{}).Where("merchant_id = ? AND status = ?", merchantID, models.InvoiceStatusPaid).Select("COALESCE(SUM(amount), 0)").Scan(&paidAmount)

	return c.JSON(fiber.Map{
		"total_invoices":   totalInvoices,
		"paid_invoices":    paidInvoices,
		"pending_invoices": pendingInvoices,
		"total_amount":     totalAmount,
		"paid_amount":      paidAmount,
		"pending_amount":   totalAmount - paidAmount,
	})
}
