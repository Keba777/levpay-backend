package database

import (
	"fmt"
	"time"

	"github.com/Keba777/levpay-backend/internal/config"
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
	// switch config.CFG.App.Service {
	// case "app":
	// 	return DB.AutoMigrate(
	// 		&models.SystemConfig{},
	// 		&models.AuditLog{},
	// 	)
	// case "auth":
	// 	return DB.AutoMigrate(
	// 		&models.User{},
	// 		&models.Session{},
	// 		&models.PasswordReset{},
	// 		&models.TwoFactorSecret{},
	// 	)
	// case "user":
	// 	return DB.AutoMigrate(&models.UserProfile{})
	// case "wallet":
	// 	return DB.AutoMigrate(&models.Wallet{}, &models.WalletTransaction{})
	// case "transaction":
	// 	return DB.AutoMigrate(&models.Transaction{})
	// case "kyc":
	// 	return DB.AutoMigrate(&models.KYCSubmission{}, &models.KYCDocument{})
	// case "file":
	// 	return DB.AutoMigrate(&models.UploadedFile{})
	// case "notification":
	// 	return DB.AutoMigrate(&models.Notification{})
	// case "billing":
	// 	return DB.AutoMigrate(&models.Invoice{}, &models.Payment{})
	// case "admin":
	// 	return DB.AutoMigrate(&models.AdminUser{}, &models.Permission{})
	// default:
	// 	return nil
	// }
	return nil
}