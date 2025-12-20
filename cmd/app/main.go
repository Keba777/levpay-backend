package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/Keba777/levpay-backend/internal/config"
	"github.com/Keba777/levpay-backend/internal/database"
	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/Keba777/levpay-backend/internal/utils"
	"github.com/Keba777/levpay-backend/router"

	"github.com/gofiber/fiber/v2"
)

func main() {
	config.InitConfig()
	database.Connect()

	logger := utils.GetLogger("app")
	logger.Info("Running database AutoMigrate...")

	if !config.CFG.DB.SkipAutoMigrate {
		if err := database.DB.AutoMigrate(
			// Core user and authentication models
			&models.User{},
			&models.Session{},

			// Wallet and financial models
			&models.Wallet{},
			&models.Transaction{},
			&models.PaymentMethod{},
			&models.Invoice{},

			// KYC and verification models
			&models.KYCDocument{},

			// Communication models
			&models.Notification{},

			// Audit and security models
			&models.AuditLog{},
		); err != nil {
			logger.ErrorWithErr("AutoMigrate failed", err)
			panic(fmt.Sprintf("AutoMigrate failed: %v", err))
		}
	}

	logger.Info("AutoMigrate completed successfully")

	// Initialize utilities
	utils.InitJWT(os.Getenv("JWT_SECRET"))
	utils.InitPasswordUtils(12) // bcrypt complexity

	app := fiber.New(fiber.Config{
		Network: "tcp",
	})

	// Health check routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"status":  "ok",
			"service": config.CFG.App.Service,
		})
	})
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	// Setup auth routes
	router.SetupAuthRoutes(app, database.DB)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

	var serviceShutdown sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.CFG.App.Shutdown)*time.Second)
	defer cancel()

	go func() {
		<-done
		logger.Info("Graceful shutdown initiated")
		serviceShutdown.Add(1)
		defer serviceShutdown.Done()
		_ = app.ShutdownWithContext(ctx)
	}()

	if err := app.Listen(config.CFG.App.Listen); err != nil {
		logger.ErrorWithErr("Failed to start app service", err)
		panic(err)
	}

	serviceShutdown.Wait()
	logger.Info("App service shutdown completed")
}
