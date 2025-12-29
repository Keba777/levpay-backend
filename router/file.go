package router

import (
	"github.com/Keba777/levpay-backend/feature/file"
	"github.com/Keba777/levpay-backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupFileRoutes sets up routes for File service
func SetupFileRoutes(api fiber.Router, db *gorm.DB) {
	repo := file.NewRepository(db)
	handler := file.NewHandler(repo)

	fileGroup := api.Group("/files")

	// Apply JWT Middleware to all file routes
	fileGroup.Use(middleware.JWTMiddleware(db))

	// User Endpoints
	fileGroup.Post("/upload", handler.UploadFile)
	fileGroup.Get("/", handler.GetFiles)
	fileGroup.Get("/:id", handler.GetFileDetails)
	fileGroup.Get("/:id/download", handler.DownloadFile)
	fileGroup.Delete("/:id", handler.DeleteFile)
}
