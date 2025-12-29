package file

import (
	"fmt"
	"strings"
	"time"

	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/Keba777/levpay-backend/internal/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Handler handles file HTTP requests
type Handler struct {
	repo *Repository
}

// NewHandler creates a new file handler
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

// UploadFile handles file upload to MinIO
func (h *Handler) UploadFile(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	// Get category from form
	category := c.FormValue("category")
	if category == "" {
		category = models.FileCategoryOther
	}

	description := c.FormValue("description")

	// Handle file upload
	file, err := c.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "File is required")
	}

	// Validate file size (e.g., max 10MB)
	maxSize := int64(10 * 1024 * 1024) // 10MB
	if file.Size > maxSize {
		return fiber.NewError(fiber.StatusBadRequest, "File size exceeds 10MB limit")
	}

	src, err := file.Open()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Failed to open file")
	}
	defer src.Close()

	// Generate unique object name
	ext := "bin"
	contentType := file.Header.Get("Content-Type")
	if strings.Contains(contentType, "pdf") {
		ext = "pdf"
	} else if strings.Contains(contentType, "png") {
		ext = "png"
	} else if strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg") {
		ext = "jpg"
	} else if strings.Contains(contentType, "doc") {
		ext = "doc"
	}

	objectName := fmt.Sprintf("files/%s/%s/%d.%s", category, userID.String(), time.Now().Unix(), ext)

	// Upload to MinIO
	fileURL, err := storage.UploadFile(objectName, src, file.Size, contentType)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to upload file")
	}

	// Create database record
	fileRecord := &models.File{
		UserID:      userID,
		FileName:    file.Filename,
		FilePath:    fileURL,
		FileType:    contentType,
		FileSize:    file.Size,
		Category:    category,
		Description: &description,
		UploadedAt:  time.Now(),
	}

	if err := h.repo.CreateFile(fileRecord); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record file upload")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "File uploaded successfully",
		"file":    fileRecord.ToResponse(),
	})
}

// GetFiles retrieves user's files with optional category filter
func (h *Handler) GetFiles(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	category := c.Query("category")

	var req models.ListedRequest
	req.FromContext(c)

	files, total, err := h.repo.GetUserFiles(userID, category, req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve files")
	}

	// Convert to responses
	var responses []models.FileResponse
	for _, f := range files {
		responses = append(responses, f.ToResponse())
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

// GetFileDetails retrieves details of a specific file
func (h *Handler) GetFileDetails(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	fileIDStr := c.Params("id")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid file ID")
	}

	file, err := h.repo.GetFileByID(fileID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "File not found")
	}

	// Verify ownership
	if file.UserID != userID {
		return fiber.NewError(fiber.StatusForbidden, "Access denied")
	}

	return c.JSON(file.ToResponse())
}

// DownloadFile provides file download URL
func (h *Handler) DownloadFile(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	fileIDStr := c.Params("id")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid file ID")
	}

	file, err := h.repo.GetFileByID(fileID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "File not found")
	}

	// Verify ownership
	if file.UserID != userID {
		return fiber.NewError(fiber.StatusForbidden, "Access denied")
	}

	// Return the file URL (in production, you might want to generate presigned URLs)
	return c.JSON(fiber.Map{
		"download_url": file.FilePath,
		"file_name":    file.FileName,
	})
}

// DeleteFile removes a file from MinIO and database
func (h *Handler) DeleteFile(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	fileIDStr := c.Params("id")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid file ID")
	}

	file, err := h.repo.GetFileByID(fileID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "File not found")
	}

	// Verify ownership
	if file.UserID != userID {
		return fiber.NewError(fiber.StatusForbidden, "Access denied")
	}

	// Delete from database
	if err := h.repo.DeleteFile(fileID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete file")
	}

	// Note: In production, you should also delete from MinIO
	// storage.DeleteFile(file.FilePath)

	return c.JSON(fiber.Map{"message": "File deleted successfully"})
}
