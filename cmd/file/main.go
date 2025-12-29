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
	"github.com/Keba777/levpay-backend/internal/storage"
	"github.com/Keba777/levpay-backend/internal/utils"
	"github.com/Keba777/levpay-backend/router"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	config.InitConfig()
	database.Connect()
	storage.InitMinio()

	logger := utils.GetLogger("file")
	logger.Info("Running database AutoMigrate...")
	if !config.CFG.DB.SkipAutoMigrate {
		if err := database.AutoMigrate(); err != nil {
			logger.ErrorWithErr("AutoMigrate failed", err)
			panic(fmt.Sprintf("AutoMigrate failed: %v", err))
		}
	}
	logger.Info("AutoMigrate completed successfully")

	app := fiber.New(fiber.Config{
		Network: "tcp",
	})

	// CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"status":  "ok",
			"service": config.CFG.App.Service,
		})
	})
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	// Setup API Routes
	api := app.Group("/api")
	router.SetupFileRoutes(api, database.DB)

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
		logger.ErrorWithErr("Failed to start file service", err)
		panic(err)
	}

	serviceShutdown.Wait()
	logger.Info("File service shutdown completed")
}
