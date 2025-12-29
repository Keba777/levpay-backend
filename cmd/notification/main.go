package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/Keba777/levpay-backend/feature/notification"
	"github.com/Keba777/levpay-backend/internal/config"
	"github.com/Keba777/levpay-backend/internal/database"
	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/Keba777/levpay-backend/internal/rabbitmq"
	"github.com/Keba777/levpay-backend/internal/utils"
	"github.com/Keba777/levpay-backend/router"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	amqp "github.com/rabbitmq/amqp091-go"
)

var svc *notification.Service

func processMessages(messages <-chan amqp.Delivery) {
	logger := utils.GetLogger("notification")
	for message := range messages {
		var msg models.Message
		if err := json.Unmarshal(message.Body, &msg); err != nil {
			logger.ErrorWithErr("Failed to unmarshal message", err)
			message.Ack(false)
			continue
		}

		if err := svc.SendEmail(msg); err != nil {
			logger.ErrorWithErr("Failed to send email", err)
		} else {
			logger.Info("Email sent successfully", utils.Field{Key: "to", Value: msg.To})
		}

		message.Ack(false)
	}
}

func main() {
	config.InitConfig()
	rabbitmq.InitRabbitMQ(config.CFG)
	defer rabbitmq.RMQ.Close()
	database.Connect()

	logger := utils.GetLogger("notification")
	logger.Info("Running database AutoMigrate...")
	if !config.CFG.DB.SkipAutoMigrate {
		if err := database.AutoMigrate(); err != nil {
			logger.ErrorWithErr("AutoMigrate failed", err)
			panic(fmt.Sprintf("AutoMigrate failed: %v", err))
		}
	}
	logger.Info("AutoMigrate completed successfully")

	// Initialize Service
	svc = notification.NewService(config.CFG)

	// Start HTTP server
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
	router.SetupNotificationRoutes(api, database.DB)

	// Start RabbitMQ consumer in goroutine
	go func() {
		err := rabbitmq.RMQ.Consume(processMessages)
		if err != nil {
			logger.ErrorWithErr("Failed to consume", err)
		}
	}()

	// Graceful shutdown
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
		logger.ErrorWithErr("Failed to start notification service", err)
		panic(err)
	}

	serviceShutdown.Wait()
	logger.Info("Notification service shutdown completed")
}
