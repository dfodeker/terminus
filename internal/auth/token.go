// internal/auth/token.go
package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// TokenService handles JWT token generation and validation
type TokenService struct {
	secretKey   []byte
	issuer      string
	accessTTL   time.Duration
	refreshTTL  time.Duration
}

// Claims represents the JWT claims
type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

// NewTokenService creates a new token service
func NewTokenService(secret, issuer string, accessTTL, refreshTTL time.Duration) *TokenService {
	return &TokenService{
		secretKey:  []byte(secret),
		issuer:     issuer,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// GenerateAccessToken creates a new JWT access token
func (s *TokenService) GenerateAccessToken(userID string) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTTL)),
		},
		UserID: userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// GenerateRefreshToken creates a random refresh token
func (s *TokenService) GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// ValidateAccessToken validates an access token and returns the claims
func (s *TokenService) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token
// In a real implementation, this would check against the database
func (s *TokenService) ValidateRefreshToken(token string) (string, error) {
	// For now, this is a placeholder
	// A real implementation would:
	// 1. Look up the token hash in the database
	// 2. Verify it hasn't been revoked
	// 3. Return the associated user ID
	return "", ErrInvalidToken
}
