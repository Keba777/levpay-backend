package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/Keba777/levpay-backend/internal/utils"
)

var CFG *models.Config

func InitConfig() {
	logger := utils.GetLogger("config")

	CFG = &models.Config{
		DB: models.Database{
			Host:            getEnvString("DB_HOST", "db"),
			Port:            getEnvInt("DB_PORT", 5432),
			User:            getEnvString("DB_USER", "postgres"),
			Pass:            getEnvString("DB_PASS", "postgres"),
			Name:            getEnvString("DB_NAME", "levpay"),
			Schema:          getEnvString("DB_SCHEMA", "public"), // overridden per service via APP_SERVICE
			SkipAutoMigrate: getEnvString("DB_SKIP_AUTO_MIGRATE", "false") == "true",
		},
		App: models.App{
			Host:     getEnvString("APP_HOST", "[::]"),
			Port:     getEnvInt("APP_PORT", 5000),
			Listen:   fmt.Sprintf("%s:%d", getEnvString("APP_HOST", "[::]"), getEnvInt("APP_PORT", 5000)),
			Shutdown: getEnvInt("APP_SHUTDOWN", 30),
			Service:  os.Getenv("APP_SERVICE"), // critical: app, auth, wallet, etc.
			Url:      os.Getenv("APP_URL"),
		},
		Security: models.Security{
			Complecity:     getEnvInt("SECURITY_COMPLECITY", 14),
			Secret:         getEnvString("SECURITY_SECRET", "your-super-secret-jwt-key-here"),
			AccessExpiries: getEnvInt("SECURITY_ACCESS_EXPIRIES", 30*60),     // 30 mins
			RefreshExpiries: getEnvInt("SECURITY_REFRESH_EXPIRIES", 7*24*60*60), // 7 days
			ForgotExpiries: getEnvInt("SECURITY_FORGOT_EXPIRIES", 45*60),     // 45 mins
		},
		RMQ: models.RMQ{
			Host:     getEnvString("RMQ_HOST", "rabbitmq"),
			Port:     getEnvInt("RMQ_PORT", 5672),
			User:     getEnvString("RMQ_USER", "rmquser"),
			Pass:     getEnvString("RMQ_PASS", "rmqpassword"),
			Queue:    getEnvString("RMQ_QUEUE", "general"),
			Exchange: getEnvString("RMQ_EXCHANGE", ""),
		},
		MSG: models.MSG{
			From:     getEnvString("MSG_FROM", ""),
			FromName: getEnvString("MSG_FROM_NAME", ""),
			Provider: getEnvString("MSG_PROVIDER", ""),
			SendGrid: models.SendGrid{
				APIKey: getEnvString("MSG_SENDGRID_APIKEY", ""),
				URL:    getEnvString("MSG_SENDGRID_URL", ""),
			},
		},
		Minio: models.Minio{
			Host:     getEnvString("MINIO_HOST", "minio"),
			Port:     getEnvInt("MINIO_PORT", 9000),
			User:     getEnvString("MINIO_USER", "minioadmin"),
			Pass:     getEnvString("MINIO_PASS", "minioadmin"),
			Bucket:   getEnvString("MINIO_BUCKET", "kyc-files"),
			SSL:      getEnvInt("MINIO_SSL", 0) == 1,
			Endpoint: fmt.Sprintf("%s:%d", getEnvString("MINIO_HOST", "minio"), getEnvInt("MINIO_PORT", 9000)),
		},
		Redis: models.Redis{
			Host:     getEnvString("REDIS_HOST", "redis"),
			Port:     getEnvInt("REDIS_PORT", 6379),
			Password: getEnvString("REDIS_PASSWORD", "redispassword"),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		Payments: models.Payments{
			TelebirrKey: getEnvString("TELEBIRR_KEY", ""),
			ChapaKey:    getEnvString("CHAPA_KEY", ""),
			BankAPIKey:  getEnvString("BANK_API_KEY", ""),
			WebhookSecret: getEnvString("PAYMENT_WEBHOOK_SECRET", ""),
			PriceBasicMonthly: getEnvString("PRICE_BASIC_MONTHLY", ""),
			PricePremiumMonthly: getEnvString("PRICE_PREMIUM_MONTHLY", ""),
			PriceEnterpriseMonthly: getEnvString("PRICE_ENTERPRISE_MONTHLY", ""),
			PriceBasicYearly: getEnvString("PRICE_BASIC_YEARLY", ""),
			PricePremiumYearly: getEnvString("PRICE_PREMIUM_YEARLY", ""),
			PriceEnterpriseYearly: getEnvString("PRICE_ENTERPRISE_YEARLY", ""),
		},
	}

	// Auto-set schema based on service name
	schemaMap := map[string]string{
		"app":          "app",
		"auth":         "auth",
		"user":         "user",
		"wallet":       "wallet",
		"transaction":  "transaction",
		"kyc":          "kyc",
		"file":         "file",
		"notification": "notification",
		"billing":      "billing",
		"admin":        "admin",
	}
	if schema, exists := schemaMap[CFG.App.Service]; exists {
		CFG.DB.Schema = schema
	}

	logger.Info("Configuration loaded successfully",
		utils.Field{Key: "service", Value: CFG.App.Service},
		utils.Field{Key: "schema", Value: CFG.DB.Schema},
		utils.Field{Key: "port", Value: CFG.App.Listen},
	)
}

func getEnvString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}