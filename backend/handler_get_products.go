package main

import (
	"log/slog"
	"net/http"

	"github.com/dfodeker/terminus/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerListProducts(w http.ResponseWriter, r *http.Request) {
	storeParam := chi.URLParam(r, "store")

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
	response := []Product{}

	products, err := cfg.db.GetProductsByStore(r.Context(), storeID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "unable to get stores", err)
		return
	}
	for _, product := range products {
		response = append(response, Product{
			Id:          product.ID,
			Title:       product.Name,
			Description: product.Description.String,
			Handle:      product.Handle,
			CreatedAt:   product.CreatedAt,
			UpdatedAt:   product.UpdatedAt,
		})

	}
	respondWithJSON(w, http.StatusOK, response)
}
