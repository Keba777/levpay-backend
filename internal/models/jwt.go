package models

import "github.com/google/uuid"

// JWT token type constants
const (
	JWTAccess  int8 = 0
	JWTRefresh int8 = 1
	JWTForgot  int8 = 2
)

// DecodedToken represents the claims in a JWT token
type DecodedToken struct {
	UserID   uuid.UUID `json:"user_id"`
	Expiries int       `json:"exp"`
	Type     int8      `json:"token_type"`
	Role     string    `json:"role,omitempty"` // user, merchant, admin
}
