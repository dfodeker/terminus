package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/dfodeker/storeos/internal/application/product"
	"github.com/dfodeker/storeos/internal/domain"
	"github.com/go-chi/chi/v5"
)

type ProductHandler struct {
	service *product.Service
}

func NewProductHandler(service *product.Service) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{productID}", h.GetByID)
	r.Put("/{productID}", h.Update)
	r.Delete("/{productID}", h.Delete)

	return r
}

type CreateProductRequest struct {
	Title       string `json:"title" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"max=5000"`
	PriceCents  int64  `json:"price_cents" validate:"required,min=0"`
	SKU         string `json:"sku" validate:"required"`
	Inventory   int    `json:"inventory" validate:"min=0"`
}

type ProductResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Handle      string `json:"handle"`
	PriceCents  int64  `json:"price_cents"`
	SKU         string `json:"sku"`
	Inventory   int    `json:"inventory"`
	Status      string `json:"status"`
}

func toProductResponse(p *domain.Product) ProductResponse {
	return ProductResponse{
		ID:          string(p.ID[:]),
		Title:       p.Title,
		Description: p.Description,
		Handle:      p.Handle,
		PriceCents:  int64(p.PriceCents),
		SKU:         p.SKU,
		Inventory:   p.Inventory,
		Status:      string(p.Status),
	}
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := shopIDFromContext(ctx)

	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	p, err := h.service.Create(ctx, product.CreateProductCmd{
		ShopID:      shopID,
		Title:       req.Title,
		Description: req.Description,
		PriceCents:  req.PriceCents,
		SKU:         req.SKU,
		Inventory:   req.Inventory,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create product")
		return
	}

	writeJSON(w, http.StatusCreated, toProductResponse(p))
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement list products
	writeError(w, http.StatusNotImplemented, "not implemented")
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement get product by ID
	writeError(w, http.StatusNotImplemented, "not implemented")
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement update product
	writeError(w, http.StatusNotImplemented, "not implemented")
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement delete product
	writeError(w, http.StatusNotImplemented, "not implemented")
}

func (h *ProductHandler) ListVariants(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement list variants
	writeError(w, http.StatusNotImplemented, "not implemented")
}

func (h *ProductHandler) CreateVariant(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement create variant
	writeError(w, http.StatusNotImplemented, "not implemented")
}

// GetShop returns info about the current shop
func GetShop(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement get shop info
	writeError(w, http.StatusNotImplemented, "not implemented")
}

// Helper functions

type contextKey string

const shopIDContextKey contextKey = "shopID"

func shopIDFromContext(ctx context.Context) domain.ShopID {
	if shopID, ok := ctx.Value(shopIDContextKey).(domain.ShopID); ok {
		return shopID
	}
	return domain.ShopID{}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
