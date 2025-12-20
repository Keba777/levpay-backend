package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

// BcryptHasher handles password hashing and verification
type BcryptHasher struct {
	complexity int
}

// NewBcryptHasher creates a new bcrypt hasher with specified complexity
func NewBcryptHasher(complexity int) *BcryptHasher {
	return &BcryptHasher{
		complexity: complexity,
	}
}

// HashPassword hashes a plaintext password using bcrypt
func (b *BcryptHasher) HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), b.complexity)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	// Encode to base64 for storage
	result := base64.StdEncoding.EncodeToString(hashed)
	return result, nil
}

// CheckPasswordHash compares a plaintext password with a hashed password
func (b *BcryptHasher) CheckPasswordHash(password, hash string) bool {
	decodedHash, err := base64.StdEncoding.DecodeString(hash)
	if err != nil {
		return false
	}
	err = bcrypt.CompareHashAndPassword(decodedHash, []byte(password))
	return err == nil
}

// PWD is the global password hasher instance
var PWD *BcryptHasher

// InitPasswordUtils initializes the password utilities with complexity
func InitPasswordUtils(complexity int) {
	PWD = NewBcryptHasher(complexity)
}

// GeneratePassword generates a random alphanumeric password
func GeneratePassword(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		nBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			panic(err)
		}
		result[i] = charset[nBig.Int64()]
	}
	return string(result)
}
