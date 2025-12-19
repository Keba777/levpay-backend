package models

type Database struct {
	Host            string
	Port            int
	User            string
	Pass            string
	Name            string
	Schema          string
	SkipAutoMigrate bool
}

type App struct {
	Host     string
	Port     int
	Listen   string
	Shutdown int
	Service  string
	Url      string
}

type Security struct {
	Complecity     int
	Secret         string
	RefreshExpiries int
	AccessExpiries int
	ForgotExpiries int
}

type RMQ struct {
	Host     string
	Port     int
	User     string
	Pass     string
	Queue    string
	Exchange string
}

type SendGrid struct {
	APIKey string `json:"apikey"`
	URL    string `json:"url"`
}

type MSG struct {
	From     string
	FromName string
	Provider string
	SendGrid SendGrid
}

type Minio struct {
	Host     string
	Port     int
	User     string
	Pass     string
	Bucket   string
	SSL      bool
	Endpoint string
}

type Redis struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type Payments struct {
	TelebirrKey string
	ChapaKey    string
	BankAPIKey  string
	WebhookSecret string
	PriceBasicMonthly string
	PricePremiumMonthly string
	PriceEnterpriseMonthly string
	PriceBasicYearly string
	PricePremiumYearly string
	PriceEnterpriseYearly string
}

type Config struct {
	DB       Database
	App      App
	Security Security
	RMQ      RMQ
	MSG      MSG
	Minio    Minio
	Redis    Redis
	Payments Payments // LevPay-specific payment integrations
}