// internal/http/handlers/user_auth.go
package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/dfodeker/storeos/internal/domain"
	"github.com/dfodeker/storeos/internal/ports"
	"github.com/google/uuid"
)

type UserAuthHandler struct {
	userRepo        ports.UserRepository
	cache           ports.Cache
	jwtSecret       string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	bcryptCost      int
	minPasswordLen  int
	tokenCodeTTL    time.Duration
}

type UserAuthConfig struct {
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	BcryptCost      int
	MinPasswordLen  int
	TokenCodeTTL    time.Duration
}

func NewUserAuthHandler(
	userRepo ports.UserRepository,
	cache ports.Cache,
	cfg UserAuthConfig,
) *UserAuthHandler {
	return &UserAuthHandler{
		userRepo:        userRepo,
		cache:           cache,
		jwtSecret:       cfg.JWTSecret,
		accessTokenTTL:  cfg.AccessTokenTTL,
		refreshTokenTTL: cfg.RefreshTokenTTL,
		bcryptCost:      cfg.BcryptCost,
		minPasswordLen:  cfg.MinPasswordLen,
		tokenCodeTTL:    cfg.TokenCodeTTL,
	}
}

// Request/Response types
type UserRegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type ExchangeCodeRequest struct {
	Code string `json:"code"`
}

type AuthTokenResponse struct {
	AccessToken  string           `json:"access_token"`
	RefreshToken string           `json:"refresh_token"`
	TokenType    string           `json:"token_type"`
	ExpiresIn    int              `json:"expires_in"`
	User         AuthUserResponse `json:"user"`
	Code         string           `json:"code,omitempty"`
}

type AuthUserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// Register handles user registration
func (h *UserAuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req UserRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate email
	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "email is required")
		return
	}

	// Validate password length
	if len(req.Password) < h.minPasswordLen {
		writeError(w, http.StatusBadRequest, "password too short")
		return
	}

	log.Printf("user email: %s -> user password: %s", req.Email, req.Password)

	// Check if user already exists
	_, err := h.userRepo.GetByEmail(r.Context(), req.Email)
	if err == nil {
		writeError(w, http.StatusConflict, "user already exists")
		return
	}
	log.Print("user does not exist\n")
	// Hash password
	hashedPassword, err := domain.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to process password")
		return
	}

	log.Printf("hashed %s", hashedPassword)

	// Create user
	user := &domain.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Status:       string(domain.UserStatusActive),
		Locale:       "en",
		Timezone:     "UTC",
	}

	log.Println(user)
	if err := h.userRepo.Create(r.Context(), user); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	// Generate tokens
	accessToken, err := domain.MakeJWT(user.ID, h.jwtSecret, h.accessTokenTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	refreshToken, err := h.createRefreshToken(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate refresh token")
		return
	}

	// Generate short-lived code for cross-subdomain auth
	code, err := domain.MakeJWT(user.ID, h.jwtSecret, h.tokenCodeTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate auth code")
		return
	}

	resp := AuthTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(h.accessTokenTTL.Seconds()),
		User: AuthUserResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		},
		Code: code,
	}

	writeJSON(w, http.StatusCreated, resp)
}

// Login handles user authentication
func (h *UserAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	// Get user by email
	user, err := h.userRepo.GetByEmail(r.Context(), req.Email)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Check user is active
	if !user.IsActive() {
		writeError(w, http.StatusForbidden, "account is not active")
		return
	}

	// Verify password
	if err := domain.CheckPasswordHash(req.Password, user.PasswordHash); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Update last login
	_ = h.userRepo.UpdateLastLogin(r.Context(), user.ID)

	// Generate tokens
	accessToken, err := domain.MakeJWT(user.ID, h.jwtSecret, h.accessTokenTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	code, err := domain.MakeJWT(user.ID, h.jwtSecret, h.tokenCodeTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to connect token")
		return
	}

	refreshToken, err := h.createRefreshToken(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate refresh token")
		return
	}

	resp := AuthTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(h.accessTokenTTL.Seconds()),
		User: AuthUserResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		},
		Code: code,
	}

	writeJSON(w, http.StatusOK, resp)
}

// RefreshToken handles token refresh
func (h *UserAuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	// Validate refresh token and get user ID
	userID, err := h.validateRefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	// Get user to ensure they're still active
	user, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil || !user.IsActive() {
		writeError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	// Revoke old refresh token
	_ = h.revokeRefreshToken(r.Context(), req.RefreshToken)

	// Generate new tokens
	accessToken, err := domain.MakeJWT(user.ID, h.jwtSecret, h.accessTokenTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	newRefreshToken, err := h.createRefreshToken(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate refresh token")
		return
	}

	resp := AuthTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(h.accessTokenTTL.Seconds()),

		User: AuthUserResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		},
	}

	writeJSON(w, http.StatusOK, resp)
}

// Logout revokes the refresh token
func (h *UserAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RefreshToken != "" {
		_ = h.revokeRefreshToken(r.Context(), req.RefreshToken)
	}

	w.WriteHeader(http.StatusNoContent)
}

// Exchange exchanges a short-lived auth code for tokens
func (h *UserAuthHandler) Exchange(w http.ResponseWriter, r *http.Request) {
	var req ExchangeCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Code == "" {
		writeError(w, http.StatusBadRequest, "code is required")
		return
	}

	// Validate the code (it's a short-lived JWT)
	userID, err := domain.ValidateJWT(req.Code, h.jwtSecret)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired code")
		return
	}

	// Get user to ensure they exist and are active
	user, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid code")
		return
	}

	if !user.IsActive() {
		writeError(w, http.StatusForbidden, "account is not active")
		return
	}

	// Generate new tokens
	accessToken, err := domain.MakeJWT(user.ID, h.jwtSecret, h.accessTokenTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	refreshToken, err := h.createRefreshToken(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate refresh token")
		return
	}

	resp := AuthTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(h.accessTokenTTL.Seconds()),
		User: AuthUserResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		},
	}

	writeJSON(w, http.StatusOK, resp)
}

// Refresh token helpers using Redis cache
func (h *UserAuthHandler) createRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	token, err := domain.MakeRefreshToken()
	if err != nil {
		return "", err
	}

	// Store token hash -> userID mapping in cache
	tokenHash := hashRefreshToken(token)
	cacheKey := "refresh_token:" + tokenHash

	if err := h.cache.Set(ctx, cacheKey, userID.String(), h.refreshTokenTTL); err != nil {
		return "", err
	}

	return token, nil
}

func (h *UserAuthHandler) validateRefreshToken(ctx context.Context, token string) (uuid.UUID, error) {
	tokenHash := hashRefreshToken(token)
	cacheKey := "refresh_token:" + tokenHash

	value, err := h.cache.Get(ctx, cacheKey)
	if err != nil {
		return uuid.Nil, err
	}

	userIDStr, ok := value.(string)
	if !ok {
		return uuid.Nil, domain.ErrInvalidToken
	}

	return uuid.Parse(userIDStr)
}

func (h *UserAuthHandler) revokeRefreshToken(ctx context.Context, token string) error {
	tokenHash := hashRefreshToken(token)
	cacheKey := "refresh_token:" + tokenHash
	return h.cache.Delete(ctx, cacheKey)
}

func hashRefreshToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}
