package wallet

import (
	"fmt"
	"time"

	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository handles wallet-related database operations
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new wallet repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// CreateWallet creates a new wallet for a user
func (r *Repository) CreateWallet(userID uuid.UUID, currency string) (*models.Wallet, error) {
	wallet := &models.Wallet{
		UserID:      userID,
		Balance:     0,
		Currency:    currency,
		Locked:      false,
		LastUpdated: time.Now(),
	}

	if err := r.db.Create(wallet).Error; err != nil {
		return nil, err
	}

	return wallet, nil
}

// GetWalletByUserID retrieves a user's wallet
func (r *Repository) GetWalletByUserID(userID uuid.UUID) (*models.Wallet, error) {
	var wallet models.Wallet
	if err := r.db.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		return nil, err
	}
	return &wallet, nil
}

// GetOrCreateWallet gets existing wallet or creates one if it doesn't exist
func (r *Repository) GetOrCreateWallet(userID uuid.UUID, currency string) (*models.Wallet, error) {
	wallet, err := r.GetWalletByUserID(userID)
	if err == gorm.ErrRecordNotFound {
		return r.CreateWallet(userID, currency)
	}
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

// UpdateBalance atomically updates the wallet balance
// amount can be positive (credit) or negative (debit)
func (r *Repository) UpdateBalance(userID uuid.UUID, amount float64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var wallet models.Wallet

		// Lock the row for update using raw SQL
		if err := tx.Raw("SELECT * FROM wallets WHERE user_id = ? FOR UPDATE", userID).
			Scan(&wallet).Error; err != nil {
			return err
		}

		// Check if wallet is locked
		if wallet.Locked {
			return fmt.Errorf("wallet is locked")
		}

		// Calculate new balance
		newBalance := wallet.Balance + amount

		// Prevent negative balance for debits
		if newBalance < 0 {
			return fmt.Errorf("insufficient balance")
		}

		// Update balance and timestamp
		if err := tx.Model(&wallet).Updates(map[string]interface{}{
			"balance":      newBalance,
			"last_updated": time.Now(),
		}).Error; err != nil {
			return err
		}

		return nil
	})
}

// LockWallet locks a wallet (e.g., for security reasons)
func (r *Repository) LockWallet(userID uuid.UUID) error {
	return r.db.Model(&models.Wallet{}).
		Where("user_id = ?", userID).
		Update("locked", true).Error
}

// UnlockWallet unlocks a wallet
func (r *Repository) UnlockWallet(userID uuid.UUID) error {
	return r.db.Model(&models.Wallet{}).
		Where("user_id = ?", userID).
		Update("locked", false).Error
}

// GetBalance returns the current balance for a user
func (r *Repository) GetBalance(userID uuid.UUID) (float64, error) {
	wallet, err := r.GetWalletByUserID(userID)
	if err != nil {
		return 0, err
	}
	return wallet.Balance, nil
}
