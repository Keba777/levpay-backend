package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Keba777/levpay-backend/feature/cron"
	"github.com/Keba777/levpay-backend/internal/config"
	"github.com/Keba777/levpay-backend/internal/database"
	"github.com/Keba777/levpay-backend/internal/utils"
)

func main() {
	config.InitConfig()
	database.Connect()

	logger := utils.GetLogger("cron")
	logger.Info("Running database AutoMigrate...")
	if !config.CFG.DB.SkipAutoMigrate {
		if err := database.AutoMigrate(); err != nil {
			logger.ErrorWithErr("AutoMigrate failed", err)
			panic(fmt.Sprintf("AutoMigrate failed: %v", err))
		}
	}
	logger.Info("AutoMigrate completed successfully")

	// Initialize and start cron scheduler
	scheduler := cron.NewScheduler(database.DB)
	if err := scheduler.Start(); err != nil {
		logger.ErrorWithErr("Failed to start cron scheduler", err)
		panic(err)
	}
	logger.Info("Cron scheduler started successfully")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down cron service...")
	scheduler.Stop()
	logger.Info("Cron service shutdown completed")
}
