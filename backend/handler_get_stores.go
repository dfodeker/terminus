package main

import (
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/dfodeker/terminus/internal/auth"
	"github.com/dfodeker/terminus/middleware"
	"github.com/google/uuid"
)

type Store struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Handle          string    `json:"handle"`
	Address         string    `json:"address"`
	Status          string    `json:"status"`
	DefaultCurrency string    `json:"default_currency"`
	Timezone        string    `json:"timezone"`
	Plan            string    `json:"plan"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

//we can extract this auth function we're about to write somewhere else

func (cfg *apiConfig) handlerGetStores(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	slog.InfoContext(r.Context(), "requesting resource : stores", "request_id", reqID)
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authentication credentials are missing or invalid", err)
		return
	}
	user, err := auth.ValidateJWT(bearerToken, cfg.signingKey)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authentication credentials are invalid.", err)
		return
	}
	log.Println(user)
	response := []Store{}

	stores, err := cfg.db.GetStores(r.Context())
	if err != nil {
		respondWithError(w, http.StatusNotFound, "unable to get stores", err)
	}
	for _, store := range stores {
		response = append(response, Store{
			ID:              store.ID,
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
	respondWithJSON(w, http.StatusOK, response)

}
