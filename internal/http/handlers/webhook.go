package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type WebhookHandler struct {
	// service would be injected here
}

func NewWebhookHandler() *WebhookHandler {
	return &WebhookHandler{}
}

func (h *WebhookHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Delete("/{webhookID}", h.Delete)

	return r
}

func (h *WebhookHandler) List(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}

func (h *WebhookHandler) Create(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}

func (h *WebhookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}
