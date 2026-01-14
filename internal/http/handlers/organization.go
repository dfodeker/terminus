package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/dfodeker/storeos/internal/application/organization"
	"github.com/dfodeker/storeos/internal/domain"
	"github.com/dfodeker/storeos/internal/ports"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type OrganizationHandler struct {
	service *organization.Service
}

func NewOrganizationHandler(service *organization.Service) *OrganizationHandler {
	return &OrganizationHandler{service: service}
}

func (h *OrganizationHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{orgID}", h.GetByID)
	r.Put("/{orgID}", h.Update)
	r.Delete("/{orgID}", h.Delete)

	// Nested resources
	r.Get("/{orgID}/shops", h.ListShops)
	r.Get("/{orgID}/can-create-shop", h.CanCreateShop)

	return r
}

// Request/Response types

type CreateOrganizationRequest struct {
	Name         string `json:"name" validate:"required,min=1,max=255"`
	Handle       string `json:"handle" validate:"required,min=1,max=100"`
	BillingEmail string `json:"billing_email" validate:"required,email"`
	Plan         string `json:"plan,omitempty"`
}

type UpdateOrganizationRequest struct {
	Name             *string          `json:"name,omitempty"`
	Handle           *string          `json:"handle,omitempty"`
	BillingEmail     *string          `json:"billing_email,omitempty"`
	Plan             *string          `json:"plan,omitempty"`
	StripeCustomerID *string          `json:"stripe_customer_id,omitempty"`
	Status           *string          `json:"status,omitempty"`
	MaxShops         *int32           `json:"max_shops,omitempty"`
	MaxMembers       *int32           `json:"max_members,omitempty"`
	Settings         *json.RawMessage `json:"settings,omitempty"`
}

type OrganizationResponse struct {
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	Handle           string           `json:"handle"`
	BillingEmail     string           `json:"billing_email"`
	Plan             string           `json:"plan"`
	Status           string           `json:"status"`
	MaxShops         int32            `json:"max_shops"`
	MaxMembers       int32            `json:"max_members"`
	StripeCustomerID string           `json:"stripe_customer_id,omitempty"`
	Settings         *json.RawMessage `json:"settings,omitempty"`
}

type ListOrganizationsResponse struct {
	Organizations []OrganizationResponse `json:"organizations"`
	NextCursor    string                 `json:"next_cursor,omitempty"`
}

type CanCreateShopResponse struct {
	Allowed      bool   `json:"allowed"`
	CurrentCount int    `json:"current_count"`
	MaxAllowed   int32  `json:"max_allowed"`
	Message      string `json:"message,omitempty"`
}

func toOrganizationResponse(o *domain.Organization) OrganizationResponse {
	resp := OrganizationResponse{
		ID:               uuid.UUID(o.ID).String(),
		Name:             o.Name,
		Handle:           o.Handle,
		BillingEmail:     o.BillingEmail,
		Plan:             o.Plan,
		Status:           o.Status,
		MaxShops:         o.MaxShops,
		MaxMembers:       o.MaxMembers,
		StripeCustomerID: o.StripeCustomerID,
	}
	if len(o.Settings) > 0 {
		settings := json.RawMessage(o.Settings)
		resp.Settings = &settings
	}
	return resp
}

func (h *OrganizationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	org, err := h.service.Create(r.Context(), organization.CreateOrganizationCmd{
		Name:         req.Name,
		Handle:       req.Handle,
		BillingEmail: req.BillingEmail,
		Plan:         req.Plan,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toOrganizationResponse(org))
}

func (h *OrganizationHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	filter := ports.OrganizationFilter{
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

	if plan := query.Get("plan"); plan != "" {
		filter.Plan = &plan
	}

	orgs, nextCursor, err := h.service.List(r.Context(), filter)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response := ListOrganizationsResponse{
		Organizations: make([]OrganizationResponse, len(orgs)),
		NextCursor:    nextCursor,
	}
	for i, o := range orgs {
		orgPtr := &o
		response.Organizations[i] = toOrganizationResponse(orgPtr)
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *OrganizationHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "orgID")

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization ID")
		return
	}

	org, err := h.service.GetByID(r.Context(), orgID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toOrganizationResponse(org))
}

func (h *OrganizationHandler) Update(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "orgID")

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization ID")
		return
	}

	var req UpdateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := organization.UpdateOrganizationCmd{
		ID:               orgID,
		Name:             req.Name,
		Handle:           req.Handle,
		BillingEmail:     req.BillingEmail,
		Plan:             req.Plan,
		StripeCustomerID: req.StripeCustomerID,
		Status:           req.Status,
		MaxShops:         req.MaxShops,
		MaxMembers:       req.MaxMembers,
	}
	if req.Settings != nil {
		cmd.Settings = *req.Settings
	}

	org, err := h.service.Update(r.Context(), cmd)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toOrganizationResponse(org))
}

func (h *OrganizationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "orgID")

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization ID")
		return
	}

	if err := h.service.Delete(r.Context(), orgID); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *OrganizationHandler) ListShops(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "orgID")

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization ID")
		return
	}

	query := r.URL.Query()
	filter := ports.ShopFilter{
		OrganizationID: &orgID,
		Limit:          50,
	}

	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			filter.Limit = limit
		}
	}

	if cursor := query.Get("cursor"); cursor != "" {
		filter.Cursor = cursor
	}

	shops, nextCursor, err := h.service.ListShops(r.Context(), orgID, filter)
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

func (h *OrganizationHandler) CanCreateShop(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "orgID")

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid organization ID")
		return
	}

	allowed, currentCount, maxAllowed, err := h.service.CanCreateShop(r.Context(), orgID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response := CanCreateShopResponse{
		Allowed:      allowed,
		CurrentCount: currentCount,
		MaxAllowed:   maxAllowed,
	}
	if !allowed {
		response.Message = "Shop limit reached for this organization. Please upgrade your plan."
	}

	writeJSON(w, http.StatusOK, response)
}
