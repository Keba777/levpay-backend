package main

import (
	"fmt"

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

	// Placeholder for cron jobs
	logger.Info("Cron service started. Press Ctrl+C to stop.")
	select {}
}