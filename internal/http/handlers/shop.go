package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/dfodeker/storeos/internal/application/shop"
	"github.com/dfodeker/storeos/internal/domain"
	"github.com/dfodeker/storeos/internal/ports"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ShopHandler struct {
	service *shop.Service
}

func NewShopHandler(service *shop.Service) *ShopHandler {
	return &ShopHandler{service: service}
}

func (h *ShopHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{shopID}", h.GetByID)
	r.Put("/{shopID}", h.Update)
	r.Delete("/{shopID}", h.Delete)

	return r
}

// Request/Response types

type CreateShopRequest struct {
	OrganizationID string `json:"organization_id" validate:"required"`
	Name           string `json:"name" validate:"required,min=1,max=255"`
	Handle         string `json:"handle" validate:"required,min=1,max=100"`
	Subdomain      string `json:"subdomain" validate:"required,min=1,max=100"`
	CustomDomain   string `json:"custom_domain,omitempty"`
	Currency       string `json:"currency" validate:"required,len=3"`
	Locale         string `json:"locale,omitempty"`
	Timezone       string `json:"timezone,omitempty"`
	ShopOwner      string `json:"shop_owner,omitempty"`
	Email          string `json:"email" validate:"required,email"`
	Phone          string `json:"phone,omitempty"`
}

type UpdateShopRequest struct {
	Name         *string `json:"name,omitempty"`
	CustomDomain *string `json:"custom_domain,omitempty"`
	Currency     *string `json:"currency,omitempty"`
	Timezone     *string `json:"timezone,omitempty"`
	ShopOwner    *string `json:"shop_owner,omitempty"`
	Email        *string `json:"email,omitempty"`
	Phone        *string `json:"phone,omitempty"`
}

type ShopResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Handle       string  `json:"handle,omitempty"`
	Subdomain    string  `json:"subdomain"`
	CustomDomain string  `json:"custom_domain,omitempty"`
	Currency     string  `json:"currency"`
	Timezone     string  `json:"timezone,omitempty"`
	ShopOwner    string  `json:"shop_owner,omitempty"`
	Email        string  `json:"email"`
	Phone        string  `json:"phone,omitempty"`
	CreatedAt    *string `json:"created_at,omitempty"`
	UpdatedAt    *string `json:"updated_at,omitempty"`
}

type ListShopsResponse struct {
	Shops      []ShopResponse `json:"shops"`
	NextCursor string         `json:"next_cursor,omitempty"`
}

func toShopResponse(s *domain.Shop) ShopResponse {
	resp := ShopResponse{
		ID:           strconv.FormatUint(s.Id, 10),
		Name:         s.Name,
		Subdomain:    s.MyShopifyDomain,
		CustomDomain: s.Domain,
		Currency:     s.Currency,
		Timezone:     s.Timezone,
		ShopOwner:    s.ShopOwner,
		Email:        s.Email,
		Phone:        s.Phone,
	}
	if s.CreatedAt != nil {
		t := s.CreatedAt.Format("2006-01-02T15:04:05Z")
		resp.CreatedAt = &t
	}
	if s.UpdatedAt != nil {
		t := s.UpdatedAt.Format("2006-01-02T15:04:05Z")
		resp.UpdatedAt = &t
	}
	return resp
}

func (h *ShopHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateShopRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Parse organization ID
	orgID, err := uuid.Parse(req.OrganizationID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization_id")
		return
	}

	s, err := h.service.Create(r.Context(), shop.CreateShopCmd{
		OrganizationID: orgID,
		Name:           req.Name,
		Handle:         req.Handle,
		Subdomain:      req.Subdomain,
		CustomDomain:   req.CustomDomain,
		Currency:       req.Currency,
		Locale:         req.Locale,
		Timezone:       req.Timezone,
		ShopOwner:      req.ShopOwner,
		Email:          req.Email,
		Phone:          req.Phone,
		Source:         "api",
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toShopResponse(s))
}

func (h *ShopHandler) List(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()

	filter := ports.ShopFilter{
		Limit: 50, // default
	}

	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			filter.Limit = limit
		}
	}

	if cursor := query.Get("cursor"); cursor != "" {
		filter.Cursor = cursor
	}

	if status := query.Get("status"); status != "" {
		filter.Status = &status
	}

	if orgIDStr := query.Get("organization_id"); orgIDStr != "" {
		if orgID, err := uuid.Parse(orgIDStr); err == nil {
			filter.OrganizationID = &orgID
		}
	}

	shops, nextCursor, err := h.service.List(r.Context(), filter)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response := ListShopsResponse{
		Shops:      make([]ShopResponse, len(shops)),
		NextCursor: nextCursor,
	}
	for i, s := range shops {
		shopPtr := &s
		response.Shops[i] = toShopResponse(shopPtr)
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *ShopHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	shopIDStr := chi.URLParam(r, "shopID")

	// Try parsing as GID first, then as UUID
	var shopID uuid.UUID
	var err error

	gid, gidErr := domain.ParseGID(shopIDStr)
	if gidErr == nil && gid.Type == domain.GIDTypeShop {
		shopID = gid.ID
	} else {
		shopID, err = uuid.Parse(shopIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid shop ID")
			return
		}
	}

	s, err := h.service.GetByID(r.Context(), shopID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toShopResponse(s))
}

func (h *ShopHandler) Update(w http.ResponseWriter, r *http.Request) {
	shopIDStr := chi.URLParam(r, "shopID")

	var shopID uuid.UUID
	var err error

	gid, gidErr := domain.ParseGID(shopIDStr)
	if gidErr == nil && gid.Type == domain.GIDTypeShop {
		shopID = gid.ID
	} else {
		shopID, err = uuid.Parse(shopIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid shop ID")
			return
		}
	}

	var req UpdateShopRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	s, err := h.service.Update(r.Context(), shop.UpdateShopCmd{
		ID:           shopID,
		Name:         req.Name,
		CustomDomain: req.CustomDomain,
		Currency:     req.Currency,
		Timezone:     req.Timezone,
		ShopOwner:    req.ShopOwner,
		Email:        req.Email,
		Phone:        req.Phone,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toShopResponse(s))
}

func (h *ShopHandler) Delete(w http.ResponseWriter, r *http.Request) {
	shopIDStr := chi.URLParam(r, "shopID")

	var shopID uuid.UUID
	var err error

	gid, gidErr := domain.ParseGID(shopIDStr)
	if gidErr == nil && gid.Type == domain.GIDTypeShop {
		shopID = gid.ID
	} else {
		shopID, err = uuid.Parse(shopIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid shop ID")
			return
		}
	}

	if err := h.service.Delete(r.Context(), shopID); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleServiceError maps service errors to HTTP responses
func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case domain.IsNotFound(err):
		writeError(w, http.StatusNotFound, err.Error())
	case domain.IsValidation(err):
		writeError(w, http.StatusBadRequest, err.Error())
	case domain.IsConflict(err):
		writeError(w, http.StatusConflict, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
