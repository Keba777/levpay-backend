package payment_method

import (
	"encoding/json"

	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func getUserID(c *fiber.Ctx) (uuid.UUID, error) {
	user, ok := c.Locals("user").(models.User)
	if !ok {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid user session")
	}
	return user.ID, nil
}

func (h *Handler) ListPaymentMethods(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	methods, err := h.repo.ListByUserID(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve payment methods")
	}

	var responses []models.PaymentMethodResponse
	for _, pm := range methods {
		responses = append(responses, models.PaymentMethodResponse{
			ID:        pm.ID,
			Type:      pm.Type,
			IsDefault: pm.IsDefault,
			Verified:  pm.Verified,
			// Simplified details for demo purposes
			LastFourDigits: "****",
		})
	}

	return c.JSON(responses)
}

func (h *Handler) AddPaymentMethod(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var req models.CreatePaymentMethodRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	detailsJSON, err := json.Marshal(req.Details)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid details format")
	}

	pm := &models.PaymentMethod{
		UserID:    userID,
		Type:      req.Type,
		Details:   datatypes.JSON(detailsJSON),
		IsDefault: req.IsDefault,
		Verified:  true, // Auto-verify for demo
	}

	if err := h.repo.Create(pm); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create payment method")
	}

	if req.IsDefault {
		if err := h.repo.SetDefault(userID, pm.ID); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to set as default")
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Payment method added successfully",
		"id":      pm.ID,
	})
}

func (h *Handler) RemovePaymentMethod(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid ID format")
	}

	pm, err := h.repo.GetByID(id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Payment method not found")
	}

	if pm.UserID != userID {
		return fiber.NewError(fiber.StatusForbidden, "Access denied")
	}

	if err := h.repo.Delete(id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete payment method")
	}

	return c.JSON(fiber.Map{"message": "Payment method removed successfully"})
}

func (h *Handler) SetDefaultPaymentMethod(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid ID format")
	}

	if err := h.repo.SetDefault(userID, id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update default payment method")
	}

	return c.JSON(fiber.Map{"message": "Default payment method updated successfully"})
}
