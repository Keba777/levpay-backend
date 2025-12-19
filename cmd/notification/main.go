package main

import (
	"fmt"

	"github.com/Keba777/levpay-backend/internal/config"
	"github.com/Keba777/levpay-backend/internal/database"
	"github.com/Keba777/levpay-backend/internal/rabbitmq"
	"github.com/Keba777/levpay-backend/internal/utils"

	amqp "github.com/rabbitmq/amqp091-go"
)

func processMessages(messages <-chan amqp.Delivery) {
	for message := range messages {
		// Process notification messages here later
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

	// Placeholder for message consumption
	err := rabbitmq.RMQ.Consume(processMessages)
	if err != nil {
		logger.ErrorWithErr("Failed to consume", err)
		panic(err)
	}
}