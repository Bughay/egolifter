// internal/auth/jwt.go
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/Bughay/go-backend-layout/internal/model"

	"github.com/golang-jwt/jwt/v5"
)

// jwtCustomClaims embeds the standard RegisteredClaims and adds our custom fields.
type jwtCustomClaims struct {
	model.Claims
	jwt.RegisteredClaims
}

// Manager handles all JWT operations using a typed struct, not a global variable.
type Manager struct {
	secretKey     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewManager creates a new JWT Manager.
func NewManager(secret string, accessExpiryHours int, refreshExpiryHours int) *Manager {
	return &Manager{
		secretKey:     []byte(secret),
		accessExpiry:  time.Duration(accessExpiryHours) * time.Hour,
		refreshExpiry: time.Duration(refreshExpiryHours) * time.Hour,
	}
}

// Generate creates a signed JWT string for a given user.
func (m *Manager) Generate(userID int64, email, role string) (string, error) {
	claims := jwtCustomClaims{
		Claims: model.Claims{
			UserID: userID,
			Email:  email,
			Role:   role,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(m.accessExpiry) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "go-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", fmt.Errorf("auth: failed to sign token: %w", err)
	}
	return signedToken, nil
}

// Validate parses and validates a JWT string, returning the embedded claims.
func (m *Manager) Validate(tokenStr string) (*model.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwtCustomClaims{}, func(token *jwt.Token) (any, error) {
		// IMPORTANT: Always verify the signing method to prevent algorithm substitution attacks
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("auth: unexpected signing method: %v", token.Header["alg"])
		}
		return m.secretKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("auth: token has expired")
		}
		return nil, fmt.Errorf("auth: invalid token: %w", err)
	}

	claims, ok := token.Claims.(*jwtCustomClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("auth: token claims are invalid")
	}

	return &claims.Claims, nil
}

func (m *Manager) GenerateRefreshToken(userID int64, email, role string) (string, error) {
	claims := jwtCustomClaims{
		Claims: model.Claims{
			UserID: userID,
			Email:  email,
			Role:   role,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "go-api-refresh",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}
