package main

import (
	"context"
	"log"
	"net/http"

	"github.com/dfodeker/terminus/internal/auth"
	"github.com/google/uuid"
)

type ctxKey int

const userKey ctxKey = iota

func userFromContext(ctx context.Context) (uuid.UUID, bool) {
	u, ok := ctx.Value(userKey).(uuid.UUID)
	return u, ok
}

// we'll need to add to this later
func (cfg *apiConfig) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bearerToken, err := auth.GetBearerToken(r.Header)
		log.Printf("value here : %s", bearerToken)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Authentication credentials are missing or invalid", err)
			return
		}
		user, err := auth.ValidateJWT(bearerToken, cfg.signingKey)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Authentication credentials are invalid.", err)
			return
		}
		log.Printf("valid User: %s", user)

		ctx := context.WithValue(r.Context(), userKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
