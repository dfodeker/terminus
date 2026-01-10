package middleware

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/dfodeker/terminus/internal/database"
	"github.com/google/uuid"
)

type storeCtxKey struct{}
type tenantCtxKey struct{}

// ResolvedStore contains the store information resolved from the subdomain or custom domain
type ResolvedStore struct {
	ID       uuid.UUID
	GID      sql.NullInt64
	Handle   string
	Name     string
	TenantID uuid.NullUUID
}

// GetResolvedStore retrieves the resolved store from the context
func GetResolvedStore(ctx context.Context) (ResolvedStore, bool) {
	store, ok := ctx.Value(storeCtxKey{}).(ResolvedStore)
	return store, ok
}

// GetResolvedTenantID retrieves the resolved tenant ID from the context
func GetResolvedTenantID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(tenantCtxKey{}).(uuid.UUID)
	return id, ok
}

// StoreResolverConfig configures the store resolver middleware
type StoreResolverConfig struct {
	DB *database.Queries
}

// StoreResolver resolves the store from subdomain or custom domain
// and adds it to the request context
func StoreResolver(cfg StoreResolverConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			info, ok := GetSubdomainInfo(r.Context())
			if !ok {
				// No subdomain info, pass through
				next.ServeHTTP(w, r)
				return
			}

			// Only resolve store for store subdomains or custom domains
			if info.DomainType != DomainTypeStore && info.DomainType != DomainTypeCustom {
				next.ServeHTTP(w, r)
				return
			}

			// Skip localhost/development requests - they shouldn't trigger store resolution
			host := strings.Split(info.FullHost, ":")[0] // strip port
			if host == "localhost" || host == "127.0.0.1" || host == "0.0.0.0" {
				log.Println("local host")
				next.ServeHTTP(w, r)
				return
			}

			var store database.Store
			var err error

			if info.IsCustom {
				// Look up by custom domain
				store, err = cfg.DB.GetStoreByCustomDomain(r.Context(), info.FullHost)
			} else {
				// Look up by handle (subdomain)
				store, err = cfg.DB.GetStoreByHandle(r.Context(), info.Subdomain)
			}

			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "Store not found", http.StatusNotFound)
					return
				}
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			resolved := ResolvedStore{
				ID:       store.ID,
				GID:      store.Gid,
				Handle:   store.Handle,
				Name:     store.Name,
				TenantID: store.TenantID,
			}

			ctx := context.WithValue(r.Context(), storeCtxKey{}, resolved)
			if store.TenantID.Valid {
				ctx = context.WithValue(ctx, tenantCtxKey{}, store.TenantID.UUID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
