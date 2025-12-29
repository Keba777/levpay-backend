package user

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/Keba777/levpay-backend/internal/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Handler handles user-related HTTP requests
type Handler struct {
	repo *Repository
}

// NewHandler creates a new user handler
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// Helper to get userID from locals (set by JWT middleware)
func getUserID(c *fiber.Ctx) (uuid.UUID, error) {
	user, ok := c.Locals("user").(models.User)
	if !ok {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid user session")
	}
	return user.ID, nil
}

// GetProfile returns the current user's profile
func (h *Handler) GetProfile(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	user, err := h.repo.GetUserByID(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	return c.JSON(user.PrepareResponse())
}

// UpdateProfile updates the current user's profile
func (h *Handler) UpdateProfile(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var req models.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	updates := make(map[string]interface{})
	if req.FirstName != "" {
		updates["first_name"] = req.FirstName
	}
	if req.LastName != "" {
		updates["last_name"] = req.LastName
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Address != nil {
		updates["address"] = *req.Address
	}
	if req.Username != nil {
		// Basic validation for username uniqueness could be added here or rely on DB constraint
		updates["username"] = *req.Username
	}
	// Handle Avatar File Upload
	file, err := c.FormFile("avatar")
	if err == nil {
		// File present, upload it
		src, err := file.Open()
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Failed to open avatar file")
		}
		defer src.Close()

		// Generate unique object name
		ext := "jpg" // specific extraction logic can be added or rely on mimetype
		if strings.Contains(file.Header.Get("Content-Type"), "png") {
			ext = "png"
		} else if strings.Contains(file.Header.Get("Content-Type"), "jpeg") {
			ext = "jpeg"
		}

		objectName := fmt.Sprintf("avatars/%s.%s", userID.String(), ext)

		// Upload to MinIO
		url, err := storage.UploadFile(objectName, src, file.Size, file.Header.Get("Content-Type"))
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to upload avatar")
		}

		updates["avatar_url"] = url
	} else if req.AvatarURL != nil {
		// Fallback to URL if provided explicitly
		updates["avatar_url"] = *req.AvatarURL
	}

	if len(updates) == 0 {
		return c.JSON(fiber.Map{"message": "No changes detected"})
	}

	if err := h.repo.UpdateUser(userID, updates); err != nil {
		// Check for unique constraint violation (simplified)
		if strings.Contains(err.Error(), "duplicate key") {
			return fiber.NewError(fiber.StatusConflict, "Username or Phone already exists")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update profile")
	}

	// Fetch updated user to return
	updatedUser, err := h.repo.GetUserByID(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve updated profile")
	}

	return c.JSON(updatedUser.PrepareResponse())
}

// GetSettings returns the current user's preferences
func (h *Handler) GetSettings(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	user, err := h.repo.GetUserByID(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	// Preferences is a JSON string, try to parse it or return as is
	// Since we defined it as string in model, we can return it.
	// But it might be cleaner to unmarshal it if we had a struct,
	// for now returning the raw JSON string inside a wrapper or map is fine.

	// Actually, let's try to return it as a proper JSON object if possible,
	// but the simple way is just returning the string field if that's what frontend expects.
	// The UserResponse already includes it.
	// Let's create a specific response for settings if needed, or just return the UserResponse object which has it.
	// The implementation plan said `GetSettings` returns user settings.

	return c.JSON(fiber.Map{
		"preferences": user.Preferences,
	})
}

// UpdateSettings updates the current user's preferences
func (h *Handler) UpdateSettings(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var req models.UpdateSettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Fetch existing prefs to merge? Or just overwrite?
	// A simple approach is to read, merge, write.
	user, err := h.repo.GetUserByID(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	// Decode existing
	prefs := make(map[string]interface{})
	if user.Preferences != nil && *user.Preferences != "" {
		json.Unmarshal([]byte(*user.Preferences), &prefs)
	}

	// Apply updates
	if req.Currency != "" {
		prefs["currency"] = req.Currency
	}
	if req.Language != "" {
		prefs["language"] = req.Language
	}
	if req.Notifications != nil {
		prefs["notifications"] = *req.Notifications
	}

	// Serialize back
	newPrefsBytes, err := json.Marshal(prefs)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to serialize preferences")
	}
	newPrefsStr := string(newPrefsBytes)

	if err := h.repo.UpdateUser(userID, map[string]interface{}{"preferences": newPrefsStr}); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update settings")
	}

	return c.JSON(fiber.Map{
		"message":     "Settings updated",
		"preferences": newPrefsStr,
	})
}

// ListUsers (Admin only)
func (h *Handler) ListUsers(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	caller, err := h.repo.GetUserByID(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "User not found")
	}
	if caller.Role != "admin" {
		return fiber.NewError(fiber.StatusForbidden, "Admins only")
	}

	var req models.ListedRequest
	req.FromContext(c)

	users, total, err := h.repo.ListUsers(req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to list users")
	}

	// Sanitize
	var sanitized []models.UserResponse
	for _, u := range users {
		sanitized = append(sanitized, u.PrepareResponse())
	}

	return c.JSON(models.ListedResponse{
		Records: func() []interface{} {
			result := make([]interface{}, len(sanitized))
			for i, v := range sanitized {
				result[i] = v
			}
			return result
		}(),
		Total: int(total),
		Page:  req.Page,
		Limit: req.Limit,
	})
}

// SearchUsers allows users to find recipients (Limited data)
func (h *Handler) SearchUsers(c *fiber.Ctx) error {
	queryStr := c.Query("query")
	if queryStr == "" {
		return c.JSON(fiber.Map{"records": []interface{}{}, "total": 0})
	}

	req := models.ListedRequest{
		Keywords: queryStr,
		Limit:    10,
		Page:     1,
	}

	users, total, err := h.repo.ListUsers(req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to search users")
	}

	// Sanitize heavily for privacy - only show name, username, and avatar
	var sanitized []fiber.Map
	for _, u := range users {
		sanitized = append(sanitized, fiber.Map{
			"id":         u.ID,
			"first_name": u.FirstName,
			"last_name":  u.LastName,
			"username":   u.Username,
			"email":      u.Email, // Email is needed for the transfer endpoint
			"avatar_url": u.AvatarURL,
		})
	}

	return c.JSON(fiber.Map{
		"records": sanitized,
		"total":   total,
	})
}

// UpdateKYC (Admin only)
func (h *Handler) UpdateKYC(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	caller, err := h.repo.GetUserByID(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "User not found")
	}
	if caller.Role != "admin" {
		return fiber.NewError(fiber.StatusForbidden, "Admins only")
	}

	targetIDStr := c.Params("id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid target user ID")
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid body")
	}

	if err := h.repo.UpdateKYCStatus(targetID, body.Status); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update KYC status")
	}

	return c.JSON(fiber.Map{"message": "KYC status updated"})
}
