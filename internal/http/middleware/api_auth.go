// internal/http/middleware/api_auth.go
package middleware

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/dfodeker/storeos/internal/domain"
	"github.com/dfodeker/storeos/internal/ports"
	"github.com/google/uuid"
)

type contextKey string

const (
	AuthContextKey contextKey = "auth"
)

// AuthContext contains all auth info for a request
type AuthContext struct {
	ShopID         uuid.UUID
	Shop           *domain.Shop
	AuthType       AuthType
	Scopes         domain.ScopeSet
	CredentialID   *uuid.UUID // For API keys
	InstallationID *uuid.UUID // For OAuth tokens
}

type AuthType string

const (
	AuthTypeAPIKey AuthType = "api_key"
	AuthTypeOAuth  AuthType = "oauth"
)

type APIAuthMiddleware struct {
	credentialRepo   ports.APICredentialRepository
	installationRepo ports.OAuthInstallationRepository
	shopRepo         ports.ShopRepository
	cache            ports.Cache
}

func NewAPIAuthMiddleware(
	credRepo ports.APICredentialRepository,
	installRepo ports.OAuthInstallationRepository,
	shopRepo ports.ShopRepository,
	cache ports.Cache,
) *APIAuthMiddleware {
	return &APIAuthMiddleware{
		credentialRepo:   credRepo,
		installationRepo: installRepo,
		shopRepo:         shopRepo,
		cache:            cache,
	}
}

func (m *APIAuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract auth header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeAuthError(w, "missing authorization header")
			return
		}

		var authCtx *AuthContext
		var err error

		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")

			if strings.HasPrefix(token, "sk_") {
				// API Key authentication
				authCtx, err = m.authenticateAPIKey(ctx, r, token)
			} else {
				// OAuth token authentication
				authCtx, err = m.authenticateOAuthToken(ctx, r, token)
			}
		} else if strings.HasPrefix(authHeader, "Basic ") {
			// Basic auth (API key as username, empty password)
			authCtx, err = m.authenticateBasicAuth(ctx, r, authHeader)
		} else {
			writeAuthError(w, "invalid authorization format")
			return
		}

		if err != nil {
			writeAuthError(w, err.Error())
			return
		}

		// Add auth context
		ctx = context.WithValue(ctx, AuthContextKey, authCtx)
		ctx = context.WithValue(ctx, ShopIDKey, authCtx.ShopID)
		ctx = context.WithValue(ctx, ShopKey, authCtx.Shop)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *APIAuthMiddleware) authenticateAPIKey(ctx context.Context, r *http.Request, apiKey string) (*AuthContext, error) {
	// Extract prefix for lookup
	prefix := domain.ExtractKeyPrefix(apiKey)
	if prefix == "" {
		return nil, domain.ErrInvalidAPIKey
	}

	// Check cache first
	cacheKey := "api_cred:" + prefix
	if cached, err := m.cache.Get(ctx, cacheKey); err == nil {
		cred := cached.(*domain.APICredential)
		if err := domain.ValidateAPIKey(apiKey, cred); err != nil {
			return nil, err
		}
		return m.buildAuthContext(ctx, cred)
	}

	// Lookup credential
	cred, err := m.credentialRepo.GetByPrefix(ctx, prefix)
	if err != nil {
		return nil, domain.ErrInvalidAPIKey
	}

	// Validate key
	if err := domain.ValidateAPIKey(apiKey, cred); err != nil {
		return nil, err
	}

	// Cache for 5 minutes
	m.cache.Set(ctx, cacheKey, cred, 5*time.Minute)

	// Update last used (async)
	go m.credentialRepo.UpdateLastUsed(context.Background(), cred.ID)

	return m.buildAuthContext(ctx, cred)
}

func (m *APIAuthMiddleware) buildAuthContext(ctx context.Context, cred *domain.APICredential) (*AuthContext, error) {
	shop, err := m.shopRepo.GetByID(ctx, cred.ShopID)
	if err != nil {
		return nil, err
	}

	return &AuthContext{
		ShopID:       cred.ShopID,
		Shop:         shop,
		AuthType:     AuthTypeAPIKey,
		Scopes:       domain.NewScopeSet(cred.Scopes),
		CredentialID: &cred.ID,
	}, nil
}

func (m *APIAuthMiddleware) authenticateOAuthToken(ctx context.Context, r *http.Request, token string) (*AuthContext, error) {
	// Similar flow for OAuth tokens
	// Hash token, lookup installation, validate, return context
	// ...
	return nil, nil
}

func (m *APIAuthMiddleware) authenticateBasicAuth(ctx context.Context, r *http.Request, authHeader string) (*AuthContext, error) {
	// Basic auth expects API key as username with empty password
	// Format: Basic base64(apikey:)
	encoded := strings.TrimPrefix(authHeader, "Basic ")
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, domain.ErrInvalidAPIKey
	}

	// Split by colon - API key is username, password should be empty
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return nil, domain.ErrInvalidAPIKey
	}

	apiKey := parts[0]
	if apiKey == "" {
		return nil, domain.ErrInvalidAPIKey
	}

	return m.authenticateAPIKey(ctx, r, apiKey)
}

// Helper to get auth from context
func AuthFromContext(ctx context.Context) *AuthContext {
	auth, _ := ctx.Value(AuthContextKey).(*AuthContext)
	return auth
}

func writeAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("WWW-Authenticate", `Bearer realm="api"`)
	http.Error(w, message, http.StatusUnauthorized)
}
