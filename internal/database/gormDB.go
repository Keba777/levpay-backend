package database

import (
	"fmt"
	"time"

	"github.com/Keba777/levpay-backend/internal/config"
	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/Keba777/levpay-backend/internal/utils"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

func Connect() {
	c := config.CFG
	logger := utils.GetLogger("database")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		c.DB.Host, c.DB.User, c.DB.Pass, c.DB.Name, c.DB.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   c.DB.Schema + ".", // e.g., auth.users, wallet.wallets
			SingularTable: false,
		},
		PrepareStmt: true,
	})
	if err != nil {
		logger.Error("Failed to connect to database", utils.Field{Key: "error", Value: err})
	}

	DB = db
	sqlDB, _ := DB.DB()
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(50)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	logger.Info("Database connected",
		utils.Field{Key: "schema", Value: c.DB.Schema},
		utils.Field{Key: "service", Value: c.App.Service},
	)

	if !c.DB.SkipAutoMigrate {
		logger.Info("Running AutoMigrate for service", utils.Field{Key: "service", Value: c.App.Service})
		if err := AutoMigrate(); err != nil {
			logger.Error("AutoMigrate failed", utils.Field{Key: "error", Value: err})
		}
		logger.Info("AutoMigrate completed successfully")
	}
}

func AutoMigrate() error {
	return DB.AutoMigrate(
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
	)
}
