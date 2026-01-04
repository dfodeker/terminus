package main

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/dfodeker/terminus/internal/database"
	"github.com/dfodeker/terminus/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ProductCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        uuid.UUID `json:"id"`
}

var productCursorCodec = CursorCodec[ProductCursor]{
	Validate: func(c ProductCursor) error {
		if c.CreatedAt.IsZero() || c.ID == uuid.Nil {
			return errors.New("missing fields")
		}
		return nil
	},
}

var cursorCreatedAt sql.NullTime
var cursorID uuid.NullUUID

func cursorInfo(cursor string) (time.Time, uuid.UUID, bool, error) {
	cur, ok, err := productCursorCodec.Decode(cursor)
	if err != nil {
		return time.Time{}, uuid.UUID{}, false, err
	}
	if !ok {
		return time.Time{}, uuid.UUID{}, false, nil
	}
	return cur.CreatedAt, cur.ID, true, nil
}

var defaultLimit int = 50
var maxLimit int = 100

func (cfg *apiConfig) handlerListProducts(w http.ResponseWriter, r *http.Request) {
	storeParam := chi.URLParam(r, "store")

	pageParams, err := ParsePageParams(r, defaultLimit, maxLimit)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid cursor", err)
	}
	limit := pageParams.Limit
	limitPlusOne := pageParams.Limit + 1

	cursorCreatedAt, cursorID, hasCursor, err := cursorInfo(pageParams.Cursor)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid cursor", err)
		return
	}

	storeID, err := uuid.Parse(storeParam)
	if err != nil {
		respondWithError(w, http.StatusForbidden, "Invalid UUID format", err)
		return
	}
	reqID := middleware.GetRequestID(r.Context())
	slog.InfoContext(r.Context(), "creating resource : stores", "request_id", reqID)
	user, ok := userFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}
	_ = user

	rows, err := cfg.db.GetProductsByStorePaginated(
		r.Context(),
		database.GetProductsByStorePaginatedParams{
			StoreID: storeID,
			Column2: hasCursor, //has cursor
			// only meaningful if hasCursor=true, but must be provided
			Column3: cursorCreatedAt, //Cursor Time
			Column4: cursorID,        //Cursor ID
			Limit:   int32(limitPlusOne),
		},
	)
	if err != nil {
		respondWithError(w, http.StatusServiceUnavailable, "Unable to retrieve products", err)
		return
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	nextCursor := ""
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		nextCursor, err = productCursorCodec.Encode(ProductCursor{
			CreatedAt: last.CreatedAt,
			ID:        last.ID,
		})
		if err != nil {
			respondWithError(w, http.StatusServiceUnavailable, "unable to build cursor", err)
			return
		}
	}

	response := make([]Product, 0, len(rows))
	for _, product := range rows {
		response = append(response, Product{
			Id:          product.ID,
			Title:       product.Name,
			Description: product.Description.String,
			Handle:      product.Handle,
			CreatedAt:   product.CreatedAt,
			UpdatedAt:   product.UpdatedAt,
		})
	}

	respondWithJSON(w, http.StatusOK, map[string]any{
		"data": response,
		"page": map[string]any{
			"limit":       limit,
			"has_more":    hasMore,
			"next_cursor": nextCursor,
		},
	})
}
