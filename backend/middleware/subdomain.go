package middleware

import (
	"context"
	"net/http"
	"strings"
)

type subdomainKey struct{}

// DomainType represents the type of domain being accessed
type DomainType int

const (
	// DomainTypeAPI is for api.mystoreos.org
	DomainTypeAPI DomainType = iota
	// DomainTypeAdmin is for admin.mystoreos.org
	DomainTypeAdmin
	// DomainTypeStore is for {store-handle}.mystoreos.org
	DomainTypeStore
	// DomainTypeCustom is for custom domains like shop.example.com
	DomainTypeCustom
	// DomainTypeRoot is for the root domain mystoreos.org
	DomainTypeRoot
)

// SubdomainInfo contains information extracted from the request host
type SubdomainInfo struct {
	Subdomain  string     // The subdomain part (e.g., "myshop" from "myshop.mystoreos.org")
	DomainType DomainType // The type of domain
	FullHost   string     // The full host including port if present
	IsCustom   bool       // True if this is a custom domain (not *.mystoreos.org)
}

// SubdomainConfig configures the subdomain middleware
type SubdomainConfig struct {
	BaseDomain     string // e.g., "mystoreos.org"
	APISubdomain   string // e.g., "api"
	AdminSubdomain string // e.g., "admin"
}

// GetSubdomainInfo retrieves subdomain information from the context
func GetSubdomainInfo(ctx context.Context) (SubdomainInfo, bool) {
	info, ok := ctx.Value(subdomainKey{}).(SubdomainInfo)
	return info, ok
}

// Subdomain extracts subdomain information from the request host
// and adds it to the request context
func Subdomain(cfg SubdomainConfig) func(http.Handler) http.Handler {
	baseParts := strings.Split(cfg.BaseDomain, ".")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host := r.Host
			// Strip port if present
			if colonIdx := strings.LastIndex(host, ":"); colonIdx != -1 {
				host = host[:colonIdx]
			}

			info := SubdomainInfo{FullHost: r.Host}
			hostParts := strings.Split(host, ".")

			// Check if this is a subdomain of our base domain
			if len(hostParts) > len(baseParts) {
				suffix := strings.Join(hostParts[len(hostParts)-len(baseParts):], ".")
				if strings.EqualFold(suffix, cfg.BaseDomain) {
					// This is a subdomain of our base domain
					subdomain := strings.Join(hostParts[:len(hostParts)-len(baseParts)], ".")
					info.Subdomain = strings.ToLower(subdomain)

					switch strings.ToLower(subdomain) {
					case cfg.APISubdomain:
						info.DomainType = DomainTypeAPI
					case cfg.AdminSubdomain:
						info.DomainType = DomainTypeAdmin
					default:
						info.DomainType = DomainTypeStore
					}
				} else {
					// Custom domain
					info.DomainType = DomainTypeCustom
					info.IsCustom = true
				}
			} else if strings.EqualFold(host, cfg.BaseDomain) {
				// Root domain
				info.DomainType = DomainTypeRoot
			} else {
				// Custom domain
				info.DomainType = DomainTypeCustom
				info.IsCustom = true
			}

			ctx := context.WithValue(r.Context(), subdomainKey{}, info)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// String returns a human-readable string representation of the DomainType
func (d DomainType) String() string {
	switch d {
	case DomainTypeAPI:
		return "api"
	case DomainTypeAdmin:
		return "admin"
	case DomainTypeStore:
		return "store"
	case DomainTypeCustom:
		return "custom"
	case DomainTypeRoot:
		return "root"
	default:
		return "unknown"
	}
}
