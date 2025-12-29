package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Keba777/levpay-backend/internal/config"
	"github.com/Keba777/levpay-backend/internal/database"
	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/Keba777/levpay-backend/internal/utils"
	"github.com/google/uuid"
)

func main() {
	// Parse flags
	fullName := flag.String("name", "", "User full name")
	email := flag.String("email", "", "User email")
	password := flag.String("password", "", "User password")
	role := flag.String("role", "user", "User role (user, merchant, admin)")
	flag.Parse()

	if *email == "" || *password == "" {
		fmt.Println("Usage: go run cmd/tools/register.go -name \"First Last\" -email \"email@test.com\" -password \"pass123\" -role \"admin|merchant|user\"")
		os.Exit(1)
	}

	// Init config and DB
	config.InitConfig()
	database.Connect()

	// Init password utils
	utils.InitPasswordUtils(12)

	// Hash password
	hashedPassword, err := utils.PWD.HashPassword(*password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Split name
	names := strings.Split(*fullName, " ")
	firstName := names[0]
	lastName := ""
	if len(names) > 1 {
		lastName = strings.Join(names[1:], " ")
	}

	prefs := "{}"

	// Create user
	user := &models.User{
		ID:           uuid.New(),
		FirstName:    firstName,
		LastName:     lastName,
		Email:        *email,
		PasswordHash: &hashedPassword,
		Role:         *role,
		KYCStatus:    "verified", // Auto-verify for manual registration
		TwoFAEnabled: false,
		Preferences:  &prefs,
	}

	if err := database.DB.Create(user).Error; err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	// Create wallet
	wallet := &models.Wallet{
		UserID:   user.ID,
		Balance:  0,
		Currency: "ETB",
		Locked:   false,
	}

	if err := database.DB.Create(wallet).Error; err != nil {
		log.Printf("[Warning] Failed to create wallet for user: %v", err)
	}

	fmt.Printf("\nSuccessfully registered user:\n")
	fmt.Printf("  ID:       %s\n", user.ID)
	fmt.Printf("  Name:     %s %s\n", user.FirstName, user.LastName)
	fmt.Printf("  Email:    %s\n", user.Email)
	fmt.Printf("  Role:     %s\n", user.Role)
	fmt.Printf("\nYou can now login at http://localhost:4003/api/auth/login\n")
}
