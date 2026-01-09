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
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type TenantStoreResponse struct {
	ID              uuid.UUID  `json:"id"`
	TenantID        *uuid.UUID `json:"tenant_id,omitempty"`
	Name            string     `json:"name"`
	Handle          string     `json:"handle"`
	Address         string     `json:"address"`
	Status          string     `json:"status"`
	DefaultCurrency string     `json:"default_currency"`
	Timezone        string     `json:"timezone"`
	Plan            string     `json:"plan"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type StoreCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        uuid.UUID `json:"id"`
}

var storeCursorCodec = CursorCodec[StoreCursor]{
	Validate: func(c StoreCursor) error {
		if c.CreatedAt.IsZero() || c.ID == uuid.Nil {
			return errors.New("invalid cursor: missing required fields")
		}
		return nil
	},
}

const (
	defaultStoreLimit = 50
	maxStoreLimit     = 100
)

func (cfg *apiConfig) handlerTenantStoresCreate(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	tenantParam := chi.URLParam(r, "tenantID")

	slog.InfoContext(r.Context(), "store creation request received",
		"request_id", reqID,
		"tenant_param", tenantParam,
	)

	user, ok := userFromContext(r.Context())
	if !ok {
		slog.WarnContext(r.Context(), "store creation failed: no authenticated user",
			"request_id", reqID,
		)
		respondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	tenantID, err := uuid.Parse(tenantParam)
	if err != nil {
		slog.WarnContext(r.Context(), "store creation failed: invalid tenant ID format",
			"request_id", reqID,
			"user_id", user,
			"tenant_param", tenantParam,
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

	type parameters struct {
		Name   string `json:"name"`
		Handle string `json:"handle"`
		Plan   string `json:"plan"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		slog.WarnContext(r.Context(), "store creation failed: invalid request body",
			"request_id", reqID,
			"user_id", user,
			"tenant_id", tenantID,
			"error", err,
		)
		respondWithError(w, http.StatusBadRequest, "Please provide a valid request body", err)
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

	plan := params.Plan
	if plan == "" {
		plan = "free"
	}

	slog.DebugContext(r.Context(), "store creation: creating store in database",
		"request_id", reqID,
		"user_id", user,
		"tenant_id", tenantID,
		"store_name", params.Name,
		"store_handle", params.Handle,
	)

	store, err := cfg.db.CreateStoreForTenant(r.Context(), database.CreateStoreForTenantParams{
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
		"store_handle", store.Handle,
	)

	var tenantIDPtr *uuid.UUID
	if store.TenantID.Valid {
		tenantIDPtr = &store.TenantID.UUID
	}

	respondWithJSON(w, http.StatusCreated, TenantStoreResponse{
		ID:              store.ID,
		TenantID:        tenantIDPtr,
		Name:            store.Name,
		Handle:          store.Handle,
		Address:         store.Address,
		Status:          store.Status,
		DefaultCurrency: store.DefaultCurrency,
		Timezone:        store.Timezone,
		Plan:            store.Plan,
		CreatedAt:       store.CreatedAt,
		UpdatedAt:       store.UpdatedAt,
	})
}

func (cfg *apiConfig) handlerTenantStoresList(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	tenantParam := chi.URLParam(r, "tenantID")

	slog.InfoContext(r.Context(), "tenant stores list request received",
		"request_id", reqID,
		"tenant_param", tenantParam,
	)

	user, ok := userFromContext(r.Context())
	if !ok {
		slog.WarnContext(r.Context(), "tenant stores list failed: no authenticated user",
			"request_id", reqID,
		)
		respondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	tenantID, err := uuid.Parse(tenantParam)
	if err != nil {
		slog.WarnContext(r.Context(), "tenant stores list failed: invalid tenant ID format",
			"request_id", reqID,
			"user_id", user,
			"tenant_param", tenantParam,
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
			slog.WarnContext(r.Context(), "tenant stores list denied: user not a member of tenant",
				"request_id", reqID,
				"user_id", user,
				"tenant_id", tenantID,
			)
			respondWithError(w, http.StatusForbidden, "You are not a member of this tenant", nil)
			return
		}
		slog.ErrorContext(r.Context(), "tenant stores list failed: error checking tenant membership",
			"request_id", reqID,
			"user_id", user,
			"tenant_id", tenantID,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to verify tenant membership", err)
		return
	}

	if tenantUser.Status != "active" {
		slog.WarnContext(r.Context(), "tenant stores list denied: user membership not active",
			"request_id", reqID,
			"user_id", user,
			"tenant_id", tenantID,
			"membership_status", tenantUser.Status,
		)
		respondWithError(w, http.StatusForbidden, "Your membership in this tenant is not active", nil)
		return
	}

	pageParams, err := ParsePageParams(r, defaultStoreLimit, maxStoreLimit)
	if err != nil {
		slog.WarnContext(r.Context(), "tenant stores list failed: invalid pagination parameters",
			"request_id", reqID,
			"user_id", user,
			"tenant_id", tenantID,
			"error", err,
		)
		respondWithError(w, http.StatusBadRequest, "Invalid pagination parameters", err)
		return
	}

	limit := pageParams.Limit
	limitPlusOne := int32(pageParams.Limit + 1)

	cursorCreatedAt, cursorID, hasCursor, err := decodeStoreCursor(pageParams.Cursor)
	if err != nil {
		slog.WarnContext(r.Context(), "tenant stores list failed: invalid cursor format",
			"request_id", reqID,
			"user_id", user,
			"tenant_id", tenantID,
			"error", err,
		)
		respondWithError(w, http.StatusBadRequest, "Invalid cursor", err)
		return
	}

	slog.DebugContext(r.Context(), "tenant stores list: fetching stores from database",
		"request_id", reqID,
		"user_id", user,
		"tenant_id", tenantID,
		"limit", limit,
		"has_cursor", hasCursor,
	)

	rows, err := cfg.db.GetStoresByTenantIDPaginated(r.Context(), database.GetStoresByTenantIDPaginatedParams{
		TenantID: uuid.NullUUID{UUID: tenantID, Valid: true},
		Column2:  hasCursor,
		Column3:  cursorCreatedAt,
		Column4:  cursorID,
		Limit:    limitPlusOne,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "tenant stores list failed: database query error",
			"request_id", reqID,
			"user_id", user,
			"tenant_id", tenantID,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve stores", err)
		return
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	var nextCursor string
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		nextCursor, err = storeCursorCodec.Encode(StoreCursor{
			CreatedAt: last.CreatedAt,
			ID:        last.ID,
		})
		if err != nil {
			slog.ErrorContext(r.Context(), "tenant stores list failed: cursor encoding error",
				"request_id", reqID,
				"user_id", user,
				"tenant_id", tenantID,
				"error", err,
			)
			respondWithError(w, http.StatusInternalServerError, "Unable to build pagination cursor", err)
			return
		}
	}

	response := make([]TenantStoreResponse, 0, len(rows))
	for _, store := range rows {
		var tenantIDPtr *uuid.UUID
		if store.TenantID.Valid {
			tenantIDPtr = &store.TenantID.UUID
		}
		response = append(response, TenantStoreResponse{
			ID:              store.ID,
			TenantID:        tenantIDPtr,
			Name:            store.Name,
			Handle:          store.Handle,
			Address:         store.Address,
			Status:          store.Status,
			DefaultCurrency: store.DefaultCurrency,
			Timezone:        store.Timezone,
			Plan:            store.Plan,
			CreatedAt:       store.CreatedAt,
			UpdatedAt:       store.UpdatedAt,
		})
	}

	slog.InfoContext(r.Context(), "tenant stores list successful",
		"request_id", reqID,
		"user_id", user,
		"tenant_id", tenantID,
		"store_count", len(response),
		"has_more", hasMore,
	)

	respondWithJSON(w, http.StatusOK, map[string]any{
		"data": response,
		"page": map[string]any{
			"limit":       limit,
			"has_more":    hasMore,
			"next_cursor": nextCursor,
		},
	})
}

func decodeStoreCursor(cursor string) (time.Time, uuid.UUID, bool, error) {
	cur, ok, err := storeCursorCodec.Decode(cursor)
	if err != nil {
		return time.Time{}, uuid.UUID{}, false, err
	}
	if !ok {
		return time.Time{}, uuid.UUID{}, false, nil
	}
	return cur.CreatedAt, cur.ID, true, nil
}
