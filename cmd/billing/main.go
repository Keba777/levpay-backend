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
	"github.com/Keba777/levpay-backend/internal/utils"

	"github.com/gofiber/fiber/v2"
)

func main() {
	config.InitConfig()
	database.Connect()

	logger := utils.GetLogger("billing")
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

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"status":  "ok",
			"service": config.CFG.App.Service,
		})
	})
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

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
		logger.ErrorWithErr("Failed to start billing service", err)
		panic(err)
	}

	serviceShutdown.Wait()
	logger.Info("Billing service shutdown completed")
}