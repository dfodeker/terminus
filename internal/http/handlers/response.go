package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/dfodeker/storeos/internal/http/middleware"
)

// ErrorResponse is the standard error response format
type ErrorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"request_id,omitempty"`
}

// respondJSON writes a JSON response with the given status code
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode JSON response", "error", err)
	}
}

// respondError writes an error response to the client
// Use this for client-safe error messages (4xx errors typically)
func respondError(w http.ResponseWriter, r *http.Request, status int, message string) {
	reqID := middleware.GetRequestID(r.Context())

	resp := ErrorResponse{
		Error:     message,
		RequestID: reqID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("failed to encode error response", "error", err)
	}
}

// respondInternalError logs the actual error and returns a generic message to the client
// Use this for 5xx errors where you don't want to leak internal details
func respondInternalError(w http.ResponseWriter, r *http.Request, message string, err error) {
	reqID := middleware.GetRequestID(r.Context())

	// Log the actual error with full context
	slog.Error(message,
		"error", err,
		"request_id", reqID,
		"method", r.Method,
		"path", r.URL.Path,
	)

	resp := ErrorResponse{
		Error:     message,
		RequestID: reqID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	if encErr := json.NewEncoder(w).Encode(resp); encErr != nil {
		slog.Error("failed to encode error response", "error", encErr)
	}
}
