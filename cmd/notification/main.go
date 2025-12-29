package main

import (
	"encoding/json"
	"fmt"

	"github.com/Keba777/levpay-backend/feature/notification"
	"github.com/Keba777/levpay-backend/internal/config"
	"github.com/Keba777/levpay-backend/internal/database"
	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/Keba777/levpay-backend/internal/rabbitmq"
	"github.com/Keba777/levpay-backend/internal/utils"

	amqp "github.com/rabbitmq/amqp091-go"
)

var svc *notification.Service

func processMessages(messages <-chan amqp.Delivery) {
	logger := utils.GetLogger("notification")
	for message := range messages {
		var msg models.Message
		if err := json.Unmarshal(message.Body, &msg); err != nil {
			logger.ErrorWithErr("Failed to unmarshal message", err)
			message.Ack(false) // Ack to remove bad message or Nack to retry?
			// If strictly unmarshal error, it's likely bad format, so Ack to discard is safer than infinite loop.
			continue
		}

		if err := svc.SendEmail(msg); err != nil {
			logger.ErrorWithErr("Failed to send email", err)
			// Decide on retry logic. handling basic failure for now.
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

	// Consume messages
	err := rabbitmq.RMQ.Consume(processMessages)
	if err != nil {
		logger.ErrorWithErr("Failed to consume", err)
		panic(err)
	}
}
