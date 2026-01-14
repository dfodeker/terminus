// internal/domain/api_credential.go
package domain

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type APICredential struct {
	ID           uuid.UUID
	ShopID       uuid.UUID
	Name         string
	KeyPrefix    string // "sk_live_abc123" - shown to user
	KeyHash      string // bcrypt hash of full key
	Scopes       []Scope
	Status       CredentialStatus
	LastUsedAt   *time.Time
	RequestCount int64
	CreatedAt    time.Time
	ExpiresAt    *time.Time
}

type CredentialStatus string

const (
	CredentialStatusActive  CredentialStatus = "active"
	CredentialStatusRevoked CredentialStatus = "revoked"
)

// GenerateAPIKey creates a new API key
// Returns the full key (only shown once) and the credential
func GenerateAPIKey(shopID uuid.UUID, name string, scopes []Scope, isLive bool) (fullKey string, cred *APICredential, err error) {
	// Generate 32 random bytes
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", nil, fmt.Errorf("generate random: %w", err)
	}

	// Create key format: sk_live_abc123xyz...
	// or sk_test_abc123xyz... for test mode
	prefix := "sk_live_"
	if !isLive {
		prefix = "sk_test_"
	}

	randomPart := base64.RawURLEncoding.EncodeToString(randomBytes)
	fullKey = prefix + randomPart

	// Hash for storage
	hash, err := bcrypt.GenerateFromPassword([]byte(fullKey), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, fmt.Errorf("hash key: %w", err)
	}

	// Key prefix for identification (first 16 chars)
	keyPrefix := fullKey[:16]

	cred = &APICredential{
		ID:        uuid.New(),
		ShopID:    shopID,
		Name:      name,
		KeyPrefix: keyPrefix,
		KeyHash:   string(hash),
		Scopes:    scopes,
		Status:    CredentialStatusActive,
		CreatedAt: time.Now(),
	}

	return fullKey, cred, nil
}

// ValidateAPIKey checks if a key is valid
func ValidateAPIKey(providedKey string, stored *APICredential) error {
	if stored.Status != CredentialStatusActive {
		return ErrCredentialRevoked
	}

	if stored.ExpiresAt != nil && stored.ExpiresAt.Before(time.Now()) {
		return ErrCredentialExpired
	}

	if err := bcrypt.CompareHashAndPassword([]byte(stored.KeyHash), []byte(providedKey)); err != nil {
		return ErrInvalidAPIKey
	}

	return nil
}

// ExtractKeyPrefix gets the prefix from a full key for lookup
func ExtractKeyPrefix(fullKey string) string {
	if len(fullKey) < 16 {
		return ""
	}
	return fullKey[:16]
}
