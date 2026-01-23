// internal/http/middleware/scopes.go
package middleware

import (
	"net/http"

	"github.com/dfodeker/storeos/internal/domain"
)

// RequireScopes returns middleware that checks for required scopes
func RequireScopes(required ...domain.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := AuthFromContext(r.Context())
			if auth == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			for _, scope := range required {
				if !auth.Scopes.Has(scope) {
					w.Header().Set("X-Missing-Scope", string(scope))
					http.Error(w, "insufficient scope: "+string(scope), http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyScope checks if at least one scope is present
func RequireAnyScope(scopes ...domain.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := AuthFromContext(r.Context())
			if auth == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if !auth.Scopes.HasAny(scopes...) {
				http.Error(w, "insufficient scope", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
