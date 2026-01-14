// internal/domain/gid.go
package domain

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// GID represents a Global ID in format: gid://storeOS/{type}/{id}
// Encoded as base64 for external use
type GID struct {
	Type string
	ID   uuid.UUID
}

const gidPrefix = "gid://storeOS/"

// GID Types - matches GraphQL type names
const (
	GIDTypeProduct         = "Product"
	GIDTypeProductVariant  = "ProductVariant"
	GIDTypeOrder           = "Order"
	GIDTypeOrderLineItem   = "OrderLineItem"
	GIDTypeCustomer        = "Customer"
	GIDTypeCustomerAddress = "CustomerAddress"
	GIDTypeShop            = "Shop"
	GIDTypeWebhook         = "Webhook"
	GIDTypeCollection      = "Collection"
	GIDTypeFulfillment     = "Fulfillment"
)

// NewGID creates a new GID
func NewGID(gidType string, id uuid.UUID) GID {
	return GID{Type: gidType, ID: id}
}

// String returns the raw GID format: gid://yourplatform/Product/uuid
func (g GID) String() string {
	return fmt.Sprintf("%s%s/%s", gidPrefix, g.Type, g.ID.String())
}

// Encode returns base64 encoded GID for external use
func (g GID) Encode() string {
	return base64.RawURLEncoding.EncodeToString([]byte(g.String()))
}

// ParseGID parses a base64 encoded GID
func ParseGID(encoded string) (GID, error) {
	// Try base64 decode first
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		// Maybe it's already raw format
		decoded = []byte(encoded)
	}

	raw := string(decoded)

	// Validate prefix
	if !strings.HasPrefix(raw, gidPrefix) {
		return GID{}, fmt.Errorf("invalid GID format: missing prefix")
	}

	// Extract type and ID
	remainder := strings.TrimPrefix(raw, gidPrefix)
	parts := strings.Split(remainder, "/")

	if len(parts) != 2 {
		return GID{}, fmt.Errorf("invalid GID format: expected type/id")
	}

	gidType := parts[0]
	idStr := parts[1]

	id, err := uuid.Parse(idStr)
	if err != nil {
		return GID{}, fmt.Errorf("invalid GID: bad UUID: %w", err)
	}

	return GID{Type: gidType, ID: id}, nil
}

// MustParseGID parses or panics - use only when you know input is valid
func MustParseGID(encoded string) GID {
	gid, err := ParseGID(encoded)
	if err != nil {
		panic(err)
	}
	return gid
}

// ParseGIDOfType parses and validates the type
func ParseGIDOfType(encoded string, expectedType string) (uuid.UUID, error) {
	gid, err := ParseGID(encoded)
	if err != nil {
		return uuid.Nil, err
	}

	if gid.Type != expectedType {
		return uuid.Nil, fmt.Errorf("invalid GID type: expected %s, got %s", expectedType, gid.Type)
	}

	return gid.ID, nil
}

// Helper functions for common types
func ProductGID(id uuid.UUID) GID {
	return NewGID(GIDTypeProduct, id)
}

func OrderGID(id uuid.UUID) GID {
	return NewGID(GIDTypeOrder, id)
}

func CustomerGID(id uuid.UUID) GID {
	return NewGID(GIDTypeCustomer, id)
}

// ParseProductID is a convenience function
func ParseProductID(gidStr string) (uuid.UUID, error) {
	return ParseGIDOfType(gidStr, GIDTypeProduct)
}

func ParseOrderID(gidStr string) (uuid.UUID, error) {
	return ParseGIDOfType(gidStr, GIDTypeOrder)
}

func ParseCustomerID(gidStr string) (uuid.UUID, error) {
	return ParseGIDOfType(gidStr, GIDTypeCustomer)
}
