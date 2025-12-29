package wallet

import (
	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Handler handles wallet HTTP requests
type Handler struct {
	repo *Repository
}

// NewHandler creates a new wallet handler
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// Helper to get userID from context
func getUserID(c *fiber.Ctx) (uuid.UUID, error) {
	user, ok := c.Locals("user").(models.User)
	if !ok {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid user session")
	}
	return user.ID, nil
}

// GetBalance returns the current wallet balance
func (h *Handler) GetBalance(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	wallet, err := h.repo.GetOrCreateWallet(userID, "ETB")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve wallet")
	}

	return c.JSON(wallet.ToResponse())
}

// TopUp adds funds to the wallet
func (h *Handler) TopUp(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var req models.TopUpWalletRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.Amount <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Amount must be positive")
	}

	// Ensure wallet exists
	_, err = h.repo.GetOrCreateWallet(userID, req.Currency)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to access wallet")
	}

	// Update balance atomically
	if err := h.repo.UpdateBalance(userID, req.Amount); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update balance")
	}

	// Fetch updated wallet
	wallet, err := h.repo.GetWalletByUserID(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve updated wallet")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Top-up successful",
		"wallet":  wallet.ToResponse(),
	})
}

// Withdraw deducts funds from the wallet
func (h *Handler) Withdraw(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var req models.WithdrawRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.Amount <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Amount must be positive")
	}

	// Update balance (negative amount for withdrawal)
	if err := h.repo.UpdateBalance(userID, -req.Amount); err != nil {
		if err.Error() == "insufficient balance" {
			return fiber.NewError(fiber.StatusBadRequest, "Insufficient balance")
		}
		if err.Error() == "wallet is locked" {
			return fiber.NewError(fiber.StatusForbidden, "Wallet is locked")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to process withdrawal")
	}

	// Fetch updated wallet
	wallet, err := h.repo.GetWalletByUserID(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve updated wallet")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Withdrawal successful",
		"wallet":  wallet.ToResponse(),
	})
}

// LockWallet locks the user's wallet (admin or security feature)
func (h *Handler) LockWallet(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	if err := h.repo.LockWallet(userID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to lock wallet")
	}

	return c.JSON(fiber.Map{"message": "Wallet locked successfully"})
}

// UnlockWallet unlocks the user's wallet
func (h *Handler) UnlockWallet(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	if err := h.repo.UnlockWallet(userID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to unlock wallet")
	}

	return c.JSON(fiber.Map{"message": "Wallet unlocked successfully"})
}
