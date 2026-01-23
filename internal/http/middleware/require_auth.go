// internal/http/middleware/require_auth.go
package middleware

import (
	"context"
	"net/http"

	"github.com/dfodeker/storeos/internal/domain"
	"github.com/dfodeker/storeos/internal/ports"
	"github.com/google/uuid"
)

const (
	UserIDKey contextKey = "user_id"
	UserKey   contextKey = "user"
)

// UserContext contains user authentication info for a request
type UserContext struct {
	UserID uuid.UUID
	Email  string
}

type RequireAuthMiddleware struct {
	tokenSecret string
	userRepo    ports.UserRepository
}

func NewRequireAuthMiddleware(tokenSecret string, userRepo ports.UserRepository) *RequireAuthMiddleware {
	return &RequireAuthMiddleware{
		tokenSecret: tokenSecret,
		userRepo:    userRepo,
	}
}

func (m *RequireAuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract bearer token
		token, err := domain.GetBearerToken(r.Header)
		if err != nil {
			writeUserAuthError(w, err.Error())
			return
		}

		// Validate JWT and extract user ID
		userID, err := domain.ValidateJWT(token, m.tokenSecret)
		if err != nil {
			writeUserAuthError(w, "invalid or expired token")
			return
		}

		// Optionally fetch user from repository
		var userCtx *UserContext
		if m.userRepo != nil {
			user, err := m.userRepo.GetByID(ctx, userID)
			if err != nil {
				writeUserAuthError(w, "user not found")
				return
			}

			if user.Status != "active" {
				writeUserAuthError(w, "user account is not active")
				return
			}

			userCtx = &UserContext{
				UserID: user.ID,
				Email:  user.Email,
			}
		} else {
			userCtx = &UserContext{
				UserID: userID,
			}
		}

		// Add to context
		ctx = context.WithValue(ctx, UserIDKey, userCtx.UserID)
		ctx = context.WithValue(ctx, UserKey, userCtx)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UserIDFromContext extracts the user ID from context
func UserIDFromContext(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(UserIDKey).(uuid.UUID)
	return id
}

// UserFromContext extracts the user context from context
func UserFromContext(ctx context.Context) *UserContext {
	user, _ := ctx.Value(UserKey).(*UserContext)
	return user
}

func writeUserAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("WWW-Authenticate", `Bearer realm="user"`)
	http.Error(w, message, http.StatusUnauthorized)
}

// RequireAuth is a simple middleware function for when you don't need user lookup
func RequireAuth(tokenSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			token, err := domain.GetBearerToken(r.Header)
			if err != nil {
				writeUserAuthError(w, err.Error())
				return
			}

			userID, err := domain.ValidateJWT(token, tokenSecret)
			if err != nil {
				writeUserAuthError(w, "invalid or expired token")
				return
			}

			ctx = context.WithValue(ctx, UserIDKey, userID)
			ctx = context.WithValue(ctx, UserKey, &UserContext{UserID: userID})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
