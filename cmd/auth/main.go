package main

import (
	"context"
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
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	config.InitConfig()
	database.Connect()

	// AutoMigrate only auth-related models for this service
	if !config.CFG.DB.SkipAutoMigrate {
		database.DB.AutoMigrate(
			&models.User{},
			&models.Session{},
		)
	}

	// Initialize JWT and password utilities
	utils.InitJWT(os.Getenv("JWT_SECRET"))
	utils.InitPasswordUtils(12)

	app := fiber.New(fiber.Config{
		Network: "tcp",
	})

	// CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

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
		logger := utils.GetLogger("auth")
		logger.Info("Graceful shutdown initiated")
		serviceShutdown.Add(1)
		defer serviceShutdown.Done()
		_ = app.ShutdownWithContext(ctx)
	}()

	if err := app.Listen(config.CFG.App.Listen); err != nil {
		logger := utils.GetLogger("auth")
		logger.ErrorWithErr("Failed to start auth service", err)
		panic(err)
	}

	serviceShutdown.Wait()

	logger := utils.GetLogger("auth")
	logger.Info("Auth service shutdown completed")
}
