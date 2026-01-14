package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type OrderHandler struct {
	// service would be injected here
}

func NewOrderHandler() *OrderHandler {
	return &OrderHandler{}
}

func (h *OrderHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{orderID}", h.GetByID)
	r.Put("/{orderID}", h.Update)

	return r
}

func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}

func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}

func (h *OrderHandler) Update(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}
