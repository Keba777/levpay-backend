package transaction

import (
	"github.com/Keba777/levpay-backend/feature/wallet"
	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Handler handles transaction HTTP requests
type Handler struct {
	repo       *Repository
	walletRepo *wallet.Repository
	db         *gorm.DB
}

// NewHandler creates a new transaction handler
func NewHandler(repo *Repository, walletRepo *wallet.Repository, db *gorm.DB) *Handler {
	return &Handler{
		repo:       repo,
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

// Transfer handles P2P money transfer between users
func (h *Handler) Transfer(c *fiber.Ctx) error {
	fromUserID, err := getUserID(c)
	if err != nil {
		return err
	}

	var req models.TransferRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate amount
	if req.Amount <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Amount must be positive")
	}

	// Cannot transfer to self
	if req.ToUserID == fromUserID {
		return fiber.NewError(fiber.StatusBadRequest, "Cannot transfer to yourself")
	}

	// Execute transfer in a transaction
	var transaction *models.Transaction
	err = h.db.Transaction(func(tx *gorm.DB) error {
		// Deduct from sender
		if err := h.walletRepo.UpdateBalance(fromUserID, -req.Amount); err != nil {
			return err
		}

		// Credit to receiver
		if err := h.walletRepo.UpdateBalance(req.ToUserID, req.Amount); err != nil {
			return err
		}

		// Create transaction record
		transaction = &models.Transaction{
			FromUserID:  fromUserID,
			ToUserID:    &req.ToUserID,
			Amount:      req.Amount,
			Currency:    req.Currency,
			Type:        models.TransactionTypeTransfer,
			Status:      models.TransactionStatusCompleted,
			Description: req.Description,
		}

		if err := h.repo.CreateTransaction(transaction); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if err.Error() == "insufficient balance" {
			return fiber.NewError(fiber.StatusBadRequest, "Insufficient balance")
		}
		if err.Error() == "wallet is locked" {
			return fiber.NewError(fiber.StatusForbidden, "Wallet is locked")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Transfer failed")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":     "Transfer successful",
		"transaction": transaction.ToResponse(),
	})
}

// Payment handles merchant payment
func (h *Handler) Payment(c *fiber.Ctx) error {
	fromUserID, err := getUserID(c)
	if err != nil {
		return err
	}

	var req models.PaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate amount
	if req.Amount <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Amount must be positive")
	}

	// Execute payment in a transaction
	var transaction *models.Transaction
	err = h.db.Transaction(func(tx *gorm.DB) error {
		// Deduct from payer
		if err := h.walletRepo.UpdateBalance(fromUserID, -req.Amount); err != nil {
			return err
		}

		// Credit to merchant
		if err := h.walletRepo.UpdateBalance(req.MerchantID, req.Amount); err != nil {
			return err
		}

		// Create transaction record
		transaction = &models.Transaction{
			FromUserID:  fromUserID,
			ToUserID:    &req.MerchantID,
			Amount:      req.Amount,
			Currency:    req.Currency,
			Type:        models.TransactionTypePayment,
			Status:      models.TransactionStatusCompleted,
			Description: req.Description,
		}

		if err := h.repo.CreateTransaction(transaction); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if err.Error() == "insufficient balance" {
			return fiber.NewError(fiber.StatusBadRequest, "Insufficient balance")
		}
		if err.Error() == "wallet is locked" {
			return fiber.NewError(fiber.StatusForbidden, "Wallet is locked")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Payment failed")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":     "Payment successful",
		"transaction": transaction.ToResponse(),
	})
}

// GetHistory retrieves transaction history for the authenticated user
func (h *Handler) GetHistory(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var req models.ListedRequest
	req.FromContext(c)

	transactions, total, err := h.repo.GetUserTransactions(userID, req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve transaction history")
	}

	// Convert to responses
	var responses []models.TransactionResponse
	for _, tx := range transactions {
		responses = append(responses, tx.ToResponse())
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

// GetTransactionDetails retrieves details of a specific transaction
func (h *Handler) GetTransactionDetails(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	txIDStr := c.Params("id")
	txID, err := uuid.Parse(txIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid transaction ID")
	}

	transaction, err := h.repo.GetTransactionByID(txID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Transaction not found")
	}

	// Verify user is part of this transaction
	if transaction.FromUserID != userID && (transaction.ToUserID == nil || *transaction.ToUserID != userID) {
		return fiber.NewError(fiber.StatusForbidden, "Access denied")
	}

	return c.JSON(transaction.ToResponse())
}
