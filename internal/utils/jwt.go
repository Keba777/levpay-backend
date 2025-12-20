package utils

import (
	"fmt"
	"time"

	"github.com/Keba777/levpay-backend/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTUtils handles JWT token generation and validation
type JWTUtils struct {
	secret string
}

// GenerateToken creates a new JWT token
func (j *JWTUtils) GenerateToken(userID uuid.UUID, role string, expirySeconds int, tokenType int8) (string, error) {
	claims := jwt.MapClaims{
		"token_type": tokenType,
		"exp":        time.Now().Add(time.Duration(expirySeconds) * time.Second).Unix(),
		"user_id":    userID.String(),
		"role":       role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(j.secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signed, nil
}

// DecodeToken validates and decodes a JWT token
func (j *JWTUtils) DecodeToken(tokenString string, expectedType int8) (models.DecodedToken, error) {
	var decodedToken models.DecodedToken

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(j.secret), nil
	})

	if err != nil {
		return decodedToken, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return decodedToken, fmt.Errorf("invalid token claims or token expired")
	}

	// Extract user_id
	if userIDStr, ok := claims["user_id"].(string); ok {
		parsedUserID, err := uuid.Parse(userIDStr)
		if err != nil {
			return decodedToken, fmt.Errorf("invalid user_id format: %w", err)
		}
		decodedToken.UserID = parsedUserID
	} else {
		return decodedToken, fmt.Errorf("user_id missing or not a string")
	}

	// Extract expiration
	if exp, ok := claims["exp"].(float64); ok {
		decodedToken.Expiries = int(exp)
	} else {
		return decodedToken, fmt.Errorf("exp missing or invalid")
	}

	// Extract and validate token_type
	claimedToken, ok := claims["token_type"].(float64)
	if !ok {
		return decodedToken, fmt.Errorf("token_type missing or invalid")
	}

	// Validate token type matches expected
	if int8(claimedToken) != expectedType {
		return decodedToken, fmt.Errorf("invalid token type: expected %d, got %d", expectedType, int8(claimedToken))
	}

	decodedToken.Type = int8(claimedToken)

	// Extract role
	if role, ok := claims["role"].(string); ok {
		decodedToken.Role = role
	}

	return decodedToken, nil
}

// JWT is the global JWT utility instance
var JWT *JWTUtils

// InitJWT initializes the JWT utilities with a secret
func InitJWT(secret string) {
	JWT = &JWTUtils{
		secret: secret,
	}
}
