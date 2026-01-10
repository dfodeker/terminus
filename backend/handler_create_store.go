package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/dfodeker/terminus/internal/database"
	"github.com/dfodeker/terminus/middleware"
	"github.com/google/uuid"
)

type StoreResponse struct {
	ID        uuid.UUID  `json:"id"`
	GID       *int64     `json:"gid,omitempty"`
	TenantID  *uuid.UUID `json:"tenant_id,omitempty"`
	Name      string     `json:"name"`
	Handle    string     `json:"handle"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (cfg *apiConfig) handlerCreateStore(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	slog.InfoContext(r.Context(), "creating resource: stores", "request_id", reqID)

	user, ok := userFromContext(r.Context())
	if !ok {
		slog.WarnContext(r.Context(), "store creation failed: no authenticated user",
			"request_id", reqID,
		)
		respondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	type parameters struct {
		TenantID string `json:"tenant_id"`
		Name     string `json:"name"`
		Handle   string `json:"handle"`
		Plan     string `json:"plan"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		slog.WarnContext(r.Context(), "store creation failed: invalid request body",
			"request_id", reqID,
			"user_id", user,
			"error", err,
		)
		respondWithError(w, http.StatusBadRequest, "Please provide a valid request body", err)
		return
	}

	// Validate required fields
	if params.TenantID == "" {
		respondWithError(w, http.StatusBadRequest, "Tenant ID is required", nil)
		return
	}
	if params.Name == "" {
		respondWithError(w, http.StatusBadRequest, "Store name is required", nil)
		return
	}
	if params.Handle == "" {
		respondWithError(w, http.StatusBadRequest, "Store handle is required", nil)
		return
	}

	tenantID, err := uuid.Parse(params.TenantID)
	if err != nil {
		slog.WarnContext(r.Context(), "store creation failed: invalid tenant ID format",
			"request_id", reqID,
			"user_id", user,
			"tenant_id", params.TenantID,
			"error", err,
		)
		respondWithError(w, http.StatusBadRequest, "Invalid tenant ID format", err)
		return
	}

	// Verify user is a member of the tenant
	tenantUser, err := cfg.db.GetTenantUser(r.Context(), database.GetTenantUserParams{
		TenantID: tenantID,
		UserID:   user,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.WarnContext(r.Context(), "store creation denied: user not a member of tenant",
				"request_id", reqID,
				"user_id", user,
				"tenant_id", tenantID,
			)
			respondWithError(w, http.StatusForbidden, "You are not a member of this tenant", nil)
			return
		}
		slog.ErrorContext(r.Context(), "store creation failed: error checking tenant membership",
			"request_id", reqID,
			"user_id", user,
			"tenant_id", tenantID,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to verify tenant membership", err)
		return
	}

	if tenantUser.Status != "active" {
		slog.WarnContext(r.Context(), "store creation denied: user membership not active",
			"request_id", reqID,
			"user_id", user,
			"tenant_id", tenantID,
			"membership_status", tenantUser.Status,
		)
		respondWithError(w, http.StatusForbidden, "Your membership in this tenant is not active", nil)
		return
	}

	plan := params.Plan
	if plan == "" {
		plan = "free"
	}

	// Generate GID for the store
	storeGID := cfg.gidGen.Generate()

	slog.DebugContext(r.Context(), "store creation: creating store in database",
		"request_id", reqID,
		"user_id", user,
		"tenant_id", tenantID,
		"store_name", params.Name,
		"store_handle", params.Handle,
	)

	store, err := cfg.db.CreateStoreForTenant(r.Context(), database.CreateStoreForTenantParams{
		Gid:      sql.NullInt64{Int64: int64(storeGID), Valid: true},
		Name:     params.Name,
		Handle:   params.Handle,
		Plan:     plan,
		TenantID: uuid.NullUUID{UUID: tenantID, Valid: true},
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "store creation failed: database error",
			"request_id", reqID,
			"user_id", user,
			"tenant_id", tenantID,
			"error", err,
		)
		respondWithError(w, http.StatusBadRequest, "Unable to create store", err)
		return
	}

	slog.InfoContext(r.Context(), "store created successfully",
		"request_id", reqID,
		"user_id", user,
		"tenant_id", tenantID,
		"store_id", store.ID,
		"store_gid", storeGID,
		"store_handle", store.Handle,
	)

	response := StoreResponse{
		ID:        store.ID,
		Name:      store.Name,
		Handle:    store.Handle,
		CreatedAt: store.CreatedAt,
		UpdatedAt: store.UpdatedAt,
	}
	if store.Gid.Valid {
		response.GID = &store.Gid.Int64
	}
	if store.TenantID.Valid {
		response.TenantID = &store.TenantID.UUID
	}

	respondWithJSON(w, http.StatusCreated, response)
}
