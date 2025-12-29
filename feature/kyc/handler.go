package kyc

import (
	"fmt"
	"strings"
	"time"

	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/Keba777/levpay-backend/internal/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Handler handles KYC HTTP requests
type Handler struct {
	repo *Repository
}

// NewHandler creates a new KYC handler
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

// UploadDocument handles uploading a KYC document
func (h *Handler) UploadDocument(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	// Parse basic fields
	docType := c.FormValue("type")
	if docType == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Document type is required")
	}

	// Handle File Upload
	file, err := c.FormFile("document")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Document file is required")
	}

	src, err := file.Open()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Failed to open file")
	}
	defer src.Close()

	// Generate object name
	// kyc/{userID}/{type}/{timestamp}.ext
	ext := "jpg"
	contentType := file.Header.Get("Content-Type")
	if strings.Contains(contentType, "png") {
		ext = "png"
	} else if strings.Contains(contentType, "pdf") {
		ext = "pdf"
	} else if strings.Contains(contentType, "jpeg") {
		ext = "jpeg"
	} else {
		return fiber.NewError(fiber.StatusBadRequest, "Unsupported file type. Only JPG, PNG, PDF allowed.")
	}

	objectName := fmt.Sprintf("kyc/%s/%s/%d.%s", userID.String(), docType, time.Now().Unix(), ext)

	// Upload to MinIO
	// Note regarding bucket: plan said "kyc-documents", need to ensure config uses that or we override?
	// The storage util currently uses config.CFG.Minio.Bucket which defaults to "kyc-files"
	// (based on config.go reading). So "kyc-files" is fine.

	// We use the simpler UploadFile which assumes the configured bucket.
	fileURL, err := storage.UploadFile(objectName, src, file.Size, contentType)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to upload document")
	}

	// Create DB Record
	doc := &models.KYCDocument{
		UserID:     userID,
		Type:       docType,
		FilePath:   fileURL,
		Status:     models.KYCStatusPending,
		UploadedAt: time.Now(),
	}

	if err := h.repo.CreateDocument(doc); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to record document submission")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":  "Document uploaded successfully",
		"document": doc,
	})
}

// GetStatus returns the user's KYC documents and current status
func (h *Handler) GetStatus(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	docs, err := h.repo.GetDocumentsByUserID(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve documents")
	}

	// Determine overall status
	// Simple rule: if any approved, partial? If all necessary?
	// For now, let's just return list.
	// Or maybe derive an overall status logic:
	// - If any Rejected -> Rejected
	// - If any Pending -> Pending
	// - Else -> Approved (if docs exist)

	overall := models.KYCStatusPending
	hasRejected := false
	hasPending := false

	if len(docs) == 0 {
		overall = "not_started"
	} else {
		for _, d := range docs {
			if d.Status == models.KYCStatusRejected {
				hasRejected = true
			}
			if d.Status == models.KYCStatusPending {
				hasPending = true
			}
		}

		if hasRejected {
			overall = models.KYCStatusRejected
		} else if hasPending {
			overall = models.KYCStatusPending
		} else {
			overall = models.KYCStatusApproved
		}
	}

	return c.JSON(models.KYCStatusResponse{
		OverallStatus: overall,
		Documents:     docs,
	})
}

// Admin: ListPending
func (h *Handler) ListPending(c *fiber.Ctx) error {
	// Role check usually done by middleware
	docs, err := h.repo.GetPendingDocuments()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to list pending documents")
	}
	return c.JSON(docs)
}

// Admin: ReviewDocument
func (h *Handler) ReviewDocument(c *fiber.Ctx) error {
	docIDStr := c.Params("id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid document ID")
	}

	var req models.ReviewKYCRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.Status != models.KYCStatusApproved && req.Status != models.KYCStatusRejected {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid status. Must be 'approved' or 'rejected'")
	}

	if err := h.repo.UpdateDocumentStatus(docID, req.Status, req.Notes); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update document status")
	}

	return c.JSON(fiber.Map{"message": "Document reviewed successfully"})
}
