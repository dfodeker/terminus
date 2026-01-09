package main

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/dfodeker/terminus/internal/database"
	"github.com/dfodeker/terminus/middleware"
	"github.com/google/uuid"
)

type TenantCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        uuid.UUID `json:"id"`
}

var tenantCursorCodec = CursorCodec[TenantCursor]{
	Validate: func(c TenantCursor) error {
		if c.CreatedAt.IsZero() || c.ID == uuid.Nil {
			return errors.New("invalid cursor: missing required fields")
		}
		return nil
	},
}

const (
	defaultTenantLimit = 50
	maxTenantLimit     = 100
)

func (cfg *apiConfig) handlerTenantsList(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())

	slog.InfoContext(r.Context(), "tenant list request received",
		"request_id", reqID,
		"method", r.Method,
		"path", r.URL.Path,
	)

	user, ok := userFromContext(r.Context())
	if !ok {
		slog.WarnContext(r.Context(), "tenant list failed: no authenticated user in context",
			"request_id", reqID,
		)
		respondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	slog.DebugContext(r.Context(), "tenant list: user authenticated",
		"request_id", reqID,
		"user_id", user,
	)

	pageParams, err := ParsePageParams(r, defaultTenantLimit, maxTenantLimit)
	if err != nil {
		slog.WarnContext(r.Context(), "tenant list failed: invalid pagination parameters",
			"request_id", reqID,
			"user_id", user,
			"error", err,
		)
		respondWithError(w, http.StatusBadRequest, "Invalid pagination parameters", err)
		return
	}

	limit := pageParams.Limit
	limitPlusOne := int32(pageParams.Limit + 1)

	cursorCreatedAt, cursorID, hasCursor, err := decodeTenantCursor(pageParams.Cursor)
	if err != nil {
		slog.WarnContext(r.Context(), "tenant list failed: invalid cursor format",
			"request_id", reqID,
			"user_id", user,
			"cursor", pageParams.Cursor,
			"error", err,
		)
		respondWithError(w, http.StatusBadRequest, "Invalid cursor", err)
		return
	}

	slog.DebugContext(r.Context(), "tenant list: fetching tenants from database",
		"request_id", reqID,
		"user_id", user,
		"limit", limit,
		"has_cursor", hasCursor,
	)

	rows, err := cfg.db.GetTenantsByUserIDPaginated(r.Context(), database.GetTenantsByUserIDPaginatedParams{
		UserID:  user,
		Column2: hasCursor,
		Column3: cursorCreatedAt,
		Column4: cursorID,
		Limit:   limitPlusOne,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "tenant list failed: database query error",
			"request_id", reqID,
			"user_id", user,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve tenants", err)
		return
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	var nextCursor string
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		nextCursor, err = tenantCursorCodec.Encode(TenantCursor{
			CreatedAt: last.CreatedAt,
			ID:        last.ID,
		})
		if err != nil {
			slog.ErrorContext(r.Context(), "tenant list failed: cursor encoding error",
				"request_id", reqID,
				"user_id", user,
				"error", err,
			)
			respondWithError(w, http.StatusInternalServerError, "Unable to build pagination cursor", err)
			return
		}
	}

	response := make([]TenantResponse, 0, len(rows))
	for _, tenant := range rows {
		response = append(response, TenantResponse{
			ID:        tenant.ID,
			Name:      tenant.Name,
			Status:    tenant.Status,
			CreatedAt: tenant.CreatedAt,
			UpdatedAt: tenant.UpdatedAt,
		})
	}

	slog.InfoContext(r.Context(), "tenant list successful",
		"request_id", reqID,
		"user_id", user,
		"tenant_count", len(response),
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

func decodeTenantCursor(cursor string) (time.Time, uuid.UUID, bool, error) {
	cur, ok, err := tenantCursorCodec.Decode(cursor)
	if err != nil {
		return time.Time{}, uuid.UUID{}, false, err
	}
	if !ok {
		return time.Time{}, uuid.UUID{}, false, nil
	}
	return cur.CreatedAt, cur.ID, true, nil
}
