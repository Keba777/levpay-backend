package admin

import (
	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Handler handles admin-related HTTP requests
type Handler struct {
	repo *Repository
}

// NewHandler creates a new admin handler
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// GetDashboard returns aggregated system and transaction statistics
func (h *Handler) GetDashboard(c *fiber.Ctx) error {
	sysStats, err := h.repo.GetSystemStats()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve system statistics")
	}

	txStats, err := h.repo.GetTransactionStats()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve transaction statistics")
	}

	return c.JSON(fiber.Map{
		"system":      sysStats,
		"transaction": txStats,
	})
}

// ListUsers retrieves all users with pagination
func (h *Handler) ListUsers(c *fiber.Ctx) error {
	var req models.ListedRequest
	req.FromContext(c)

	users, total, err := h.repo.ListAllUsers(req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve users")
	}

	// Convert to generic interface slice for ListedResponse
	records := make([]interface{}, len(users))
	for i, u := range users {
		records[i] = u
	}

	return c.JSON(models.ListedResponse{
		Records: records,
		Total:   int(total),
		Page:    req.Page,
		Limit:   req.Limit,
	})
}

// UpdateUserStatus handles user activation/suspension
func (h *Handler) UpdateUserStatus(c *fiber.Ctx) error {
	userIDStr := c.Params("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")
	}

	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.repo.UpdateUserStatus(userID, req.IsActive); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update user status")
	}

	return c.JSON(fiber.Map{"message": "User status updated successfully"})
}

// GetAuditLogs retrieves system activity logs
func (h *Handler) GetAuditLogs(c *fiber.Ctx) error {
	var req models.ListedRequest
	req.FromContext(c)

	logs, total, err := h.repo.GetAuditLogs(req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve audit logs")
	}

	// Convert to generic interface slice
	records := make([]interface{}, len(logs))
	for i, l := range logs {
		records[i] = l
	}

	return c.JSON(models.ListedResponse{
		Records: records,
		Total:   int(total),
		Page:    req.Page,
		Limit:   req.Limit,
	})
}
