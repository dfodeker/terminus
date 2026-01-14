// internal/http/handlers/auth.go
package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dfodeker/storeos/internal/domain"
	"github.com/dfodeker/storeos/internal/ports"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo     ports.UserRepository
	authConfig   AuthConfig
	tokenService TokenService
}

type AuthConfig struct {
	JWTSecret          string
	JWTIssuer          string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	BcryptCost         int
	PasswordMinLength  int
}

type TokenService interface {
	GenerateAccessToken(userID string) (string, error)
	GenerateRefreshToken() (string, error)
	ValidateRefreshToken(token string) (string, error)
}

func NewAuthHandler(userRepo ports.UserRepository, config AuthConfig, tokenSvc TokenService) *AuthHandler {
	return &AuthHandler{
		userRepo:     userRepo,
		authConfig:   config,
		tokenService: tokenSvc,
	}
}

// RegisterRequest represents the user registration request
type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	User         UserResponse `json:"user"`
}

// UserResponse represents the user info in auth responses
type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate email
	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "email is required")
		return
	}

	// Validate password
	if len(req.Password) < h.authConfig.PasswordMinLength {
		writeError(w, http.StatusBadRequest, "password too short")
		return
	}

	// Check if user already exists
	_, err := h.userRepo.GetByEmail(r.Context(), req.Email)
	if err == nil {
		writeError(w, http.StatusConflict, "user already exists")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), h.authConfig.BcryptCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to process password")
		return
	}

	// Create user
	user := &domain.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Status:       "active",
		Locale:       "en",
		Timezone:     "UTC",
	}

	if err := h.userRepo.Create(r.Context(), user); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	// Generate tokens
	accessToken, err := h.tokenService.GenerateAccessToken(user.ID.String())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	refreshToken, err := h.tokenService.GenerateRefreshToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate refresh token")
		return
	}

	// Return response
	resp := AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(h.authConfig.AccessTokenTTL.Seconds()),
		User: UserResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		},
	}

	writeJSON(w, http.StatusCreated, resp)
}

// Login handles user authentication
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	// Get user
	user, err := h.userRepo.GetByEmail(r.Context(), req.Email)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Check user status
	if user.Status != "active" {
		writeError(w, http.StatusForbidden, "account is not active")
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Update last login
	h.userRepo.UpdateLastLogin(r.Context(), user.ID)

	// Generate tokens
	accessToken, err := h.tokenService.GenerateAccessToken(user.ID.String())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	refreshToken, err := h.tokenService.GenerateRefreshToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate refresh token")
		return
	}

	// Return response
	resp := AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(h.authConfig.AccessTokenTTL.Seconds()),
		User: UserResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		},
	}

	writeJSON(w, http.StatusOK, resp)
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	// Validate refresh token and get user ID
	userID, err := h.tokenService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	// Generate new access token
	accessToken, err := h.tokenService.GenerateAccessToken(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   int(h.authConfig.AccessTokenTTL.Seconds()),
	})
}

// Helper functions - use common helpers from other handlers
