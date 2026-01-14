package dataloader

// internal/graphql/dataloader/middleware.go

import (
	"context"
	"net/http"
)

type contextKey string

const loadersKey contextKey = "dataloaders"

// Middleware injects fresh dataloaders into each request
func Middleware(loaders *Loaders) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create fresh loaders for each request (important for caching per-request)
			ctx := context.WithValue(r.Context(), loadersKey, loaders)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// FromContext retrieves dataloaders from context
func FromContext(ctx context.Context) *Loaders {
	return ctx.Value(loadersKey).(*Loaders)
}
