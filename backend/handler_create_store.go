package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/dfodeker/terminus/internal/database"
	"github.com/dfodeker/terminus/middleware"
	"github.com/google/uuid"
)

type StoreResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Handle    string    `json:"handle"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (cfg *apiConfig) handlerCreateStore(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	slog.InfoContext(r.Context(), "creating resource : stores", "request_id", reqID)
	user, ok := userFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}
	_ = user // use it, log it, authorize, etc.
	log.Printf("user id: %s", user)
	type parameters struct {
		Name   string `json:"name"`
		Handle string `json:"handle"`
		Plan   string `json:"plan"`
	}
	//we wont be using created by right now
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error Decoding Params: %s", err)
		respondWithError(w, 400, "Please provide a valid request body", err)
		return
	}
	store, err := cfg.db.CreateStore(r.Context(), database.CreateStoreParams{
		Name:   params.Name,
		Handle: params.Handle,
		Plan:   params.Plan,
	})
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to create store", err)
		return
	}
	log.Printf("Creaetd by: %s", user)
	respondWithJSON(w, http.StatusCreated, StoreResponse{
		ID:        store.ID,
		Name:      store.Name,
		Handle:    store.Handle,
		CreatedAt: store.CreatedAt,
		UpdatedAt: store.UpdatedAt,
	})

}
