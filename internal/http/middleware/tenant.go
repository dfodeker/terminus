// internal/http/middleware/tenant.go
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/dfodeker/storeos/internal/domain"
	"github.com/dfodeker/storeos/internal/ports"
	"github.com/google/uuid"
)

const (
	ShopIDKey contextKey = "shop_id"
	ShopKey   contextKey = "shop"
)

type TenantMiddleware struct {
	shopRepo   ports.ShopRepository
	rootDomain string
}

func NewTenantMiddleware(shopRepo ports.ShopRepository, rootDomain string) *TenantMiddleware {
	return &TenantMiddleware{
		shopRepo:   shopRepo,
		rootDomain: rootDomain,
	}
}

func (m *TenantMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		host := r.Host

		// Try custom domain first
		shop, err := m.shopRepo.GetByCustomDomain(ctx, host)
		if err != nil {
			// Try subdomain
			subdomain := m.extractSubdomain(host)
			if subdomain == "" {
				http.Error(w, "tenant not found", http.StatusNotFound)
				return
			}

			shop, err = m.shopRepo.GetBySubdomain(ctx, subdomain)
			if err != nil {
				http.Error(w, "shop not found", http.StatusNotFound)
				return
			}
		}

		// Check shop status
		if !shop.IsActive() {
			http.Error(w, "shop is not active", http.StatusForbidden)
			return
		}

		// Add to context
		ctx = context.WithValue(ctx, ShopIDKey, shop.Id)
		ctx = context.WithValue(ctx, ShopKey, shop)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *TenantMiddleware) extractSubdomain(host string) string {
	host = strings.Split(host, ":")[0] // Remove port

	if !strings.HasSuffix(host, "."+m.rootDomain) {
		return ""
	}

	subdomain := strings.TrimSuffix(host, "."+m.rootDomain)
	if strings.Contains(subdomain, ".") {
		return "" // Nested subdomain not allowed
	}

	return subdomain
}

// ShopIDFromContext extracts the shop ID from context
func ShopIDFromContext(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(ShopIDKey).(uuid.UUID)
	return id
}

// ShopFromContext extracts the shop from context
func ShopFromContext(ctx context.Context) *domain.Shop {
	shop, _ := ctx.Value(ShopKey).(*domain.Shop)
	return shop
}
