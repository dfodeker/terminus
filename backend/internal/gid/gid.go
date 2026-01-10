package gid

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	// Scheme is the URI scheme for GIDs
	Scheme = "gid"
	// Namespace is the application namespace
	Namespace = "mystoreos"
	// Prefix is the full GID prefix
	Prefix = Scheme + "://" + Namespace + "/"
)

// GID represents a Global Identifier
type GID struct {
	Type EntityType
	ID   uint64
}

// String returns the GID in canonical form: gid://mystoreos/EntityType/id
func (g GID) String() string {
	return fmt.Sprintf("%s%s/%d", Prefix, g.Type, g.ID)
}

// Base64 returns URL-safe base64 encoded GID (for compact URLs)
func (g GID) Base64() string {
	return base64.RawURLEncoding.EncodeToString([]byte(g.String()))
}

// IsZero returns true if the GID is zero-valued
func (g GID) IsZero() bool {
	return g.ID == 0 && g.Type == ""
}

// Parse parses a GID string into its components.
// Expected format: gid://mystoreos/EntityType/id
func Parse(s string) (GID, error) {
	if !strings.HasPrefix(s, Prefix) {
		return GID{}, errors.New("invalid GID: missing prefix gid://mystoreos/")
	}

	remainder := strings.TrimPrefix(s, Prefix)
	parts := strings.SplitN(remainder, "/", 2)
	if len(parts) != 2 {
		return GID{}, errors.New("invalid GID: expected format gid://mystoreos/EntityType/id")
	}

	entityType := EntityType(parts[0])
	if !entityType.IsValid() {
		return GID{}, fmt.Errorf("invalid GID: unknown entity type %q", parts[0])
	}

	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return GID{}, fmt.Errorf("invalid GID: invalid ID %q: %w", parts[1], err)
	}

	return GID{Type: entityType, ID: id}, nil
}

// ParseBase64 parses a base64-encoded GID
func ParseBase64(s string) (GID, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return GID{}, fmt.Errorf("invalid GID: base64 decode failed: %w", err)
	}
	return Parse(string(decoded))
}

// MustParse parses a GID or panics (for tests/init)
func MustParse(s string) GID {
	g, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return g
}

// New creates a new GID with the given type and ID
func New(entityType EntityType, id uint64) GID {
	return GID{Type: entityType, ID: id}
}

// ProductGID creates a Product GID
func ProductGID(id uint64) GID {
	return New(EntityProduct, id)
}

// ProductVariantGID creates a ProductVariant GID
func ProductVariantGID(id uint64) GID {
	return New(EntityProductVariant, id)
}

// StoreGID creates a Store GID
func StoreGID(id uint64) GID {
	return New(EntityStore, id)
}

// TenantGID creates a Tenant GID
func TenantGID(id uint64) GID {
	return New(EntityTenant, id)
}

// UserGID creates a User GID
func UserGID(id uint64) GID {
	return New(EntityUser, id)
}

// RoleGID creates a Role GID
func RoleGID(id uint64) GID {
	return New(EntityRole, id)
}

// PermissionGID creates a Permission GID
func PermissionGID(id uint64) GID {
	return New(EntityPermission, id)
}

// CustomDomainGID creates a CustomDomain GID
func CustomDomainGID(id uint64) GID {
	return New(EntityCustomDomain, id)
}
