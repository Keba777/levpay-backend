package notification

import (
	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Handler handles notification HTTP requests
type Handler struct {
	repo *Repository
}

// NewHandler creates a new notification handler
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

// GetNotifications retrieves user's notification history
func (h *Handler) GetNotifications(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var req models.ListedRequest
	req.FromContext(c)

	notifications, total, err := h.repo.GetUserNotifications(userID, req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve notifications")
	}

	// Convert to responses
	var responses []models.NotificationResponse
	for _, notif := range notifications {
		responses = append(responses, notif.ToResponse())
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

// MarkAsRead marks a notification as read
func (h *Handler) MarkAsRead(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	notifIDStr := c.Params("id")
	notifID, err := uuid.Parse(notifIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid notification ID")
	}

	// Verify ownership
	notification, err := h.repo.GetNotificationByID(notifID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Notification not found")
	}

	if notification.UserID != userID {
		return fiber.NewError(fiber.StatusForbidden, "Access denied")
	}

	if err := h.repo.MarkAsRead(notifID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to mark as read")
	}

	return c.JSON(fiber.Map{"message": "Notification marked as read"})
}

// GetUnreadCount returns the count of unread notifications
func (h *Handler) GetUnreadCount(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	count, err := h.repo.GetUnreadCount(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get unread count")
	}

	return c.JSON(fiber.Map{"unread_count": count})
}
