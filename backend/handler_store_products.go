package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/dfodeker/terminus/internal/database"
	"github.com/dfodeker/terminus/internal/gid"
	"github.com/dfodeker/terminus/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ProductResponse struct {
	ID               uuid.UUID `json:"id"`
	GID              string    `json:"gid,omitempty"`
	StoreID          uuid.UUID `json:"store_id"`
	Handle           string    `json:"handle"`
	Name             string    `json:"name"`
	Description      *string   `json:"description,omitempty"`
	InventoryTracked bool      `json:"inventory_tracked"`
	SKU              *string   `json:"sku,omitempty"`
	Tags             *string   `json:"tags,omitempty"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// getStoreAndVerifyAccess looks up store by handle and verifies user has the required permission
func (cfg *apiConfig) getStoreAndVerifyAccess(r *http.Request, storeHandle string, userID uuid.UUID, permissionKey string) (database.Store, error) {
	store, err := cfg.db.GetStoreByHandle(r.Context(), storeHandle)
	if err != nil {
		return database.Store{}, err
	}

	// Store must have a tenant
	if !store.TenantID.Valid {
		return database.Store{}, errors.New("store has no tenant association")
	}

	// Check user has permission in this tenant
	hasPermission, err := cfg.db.CheckUserHasPermission(r.Context(), database.CheckUserHasPermissionParams{
		TenantID: store.TenantID.UUID,
		UserID:   userID,
		Key:      permissionKey,
	})
	if err != nil {
		return database.Store{}, err
	}

	if !hasPermission {
		return database.Store{}, errors.New("permission denied")
	}

	return store, nil
}

// handlerStoreProductCreate creates a product in a store (store-handle scoped)
func (cfg *apiConfig) handlerStoreProductCreate(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	storeHandle := chi.URLParam(r, "storeHandle")

	slog.InfoContext(r.Context(), "product creation request",
		"request_id", reqID,
		"store_handle", storeHandle,
	)

	user, ok := userFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	store, err := cfg.getStoreAndVerifyAccess(r, storeHandle, user, "products:create")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Store not found", nil)
			return
		}
		if err.Error() == "permission denied" {
			slog.WarnContext(r.Context(), "product creation denied: insufficient permissions",
				"request_id", reqID,
				"user_id", user,
				"store_handle", storeHandle,
			)
			respondWithError(w, http.StatusForbidden, "You do not have permission to create products in this store", nil)
			return
		}
		slog.ErrorContext(r.Context(), "product creation failed: error verifying access",
			"request_id", reqID,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to verify access", err)
		return
	}

	type parameters struct {
		Name             string  `json:"name"`
		Handle           string  `json:"handle"`
		Description      *string `json:"description"`
		InventoryTracked bool    `json:"inventory_tracked"`
		SKU              *string `json:"sku"`
		Tags             *string `json:"tags"`
		Status           string  `json:"status"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if params.Name == "" {
		respondWithError(w, http.StatusBadRequest, "Product name is required", nil)
		return
	}
	if params.Handle == "" {
		respondWithError(w, http.StatusBadRequest, "Product handle is required", nil)
		return
	}

	status := params.Status
	if status == "" {
		status = "active"
	}

	description := sql.NullString{}
	if params.Description != nil {
		description = sql.NullString{String: *params.Description, Valid: true}
	}

	sku := sql.NullString{}
	if params.SKU != nil {
		sku = sql.NullString{String: *params.SKU, Valid: true}
	}

	tags := sql.NullString{}
	if params.Tags != nil {
		tags = sql.NullString{String: *params.Tags, Valid: true}
	}

	product, err := cfg.db.CreateProduct(r.Context(), database.CreateProductParams{
		StoreID:          store.ID,
		Handle:           params.Handle,
		Name:             params.Name,
		Description:      description,
		InventoryTracked: params.InventoryTracked,
		Sku:              sku,
		Tags:             tags,
		Status:           status,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "product creation failed: database error",
			"request_id", reqID,
			"store_id", store.ID,
			"error", err,
		)
		respondWithError(w, http.StatusBadRequest, "Unable to create product", err)
		return
	}

	slog.InfoContext(r.Context(), "product created successfully",
		"request_id", reqID,
		"user_id", user,
		"store_handle", storeHandle,
		"product_id", product.ID,
	)

	respondWithJSON(w, http.StatusCreated, toProductResponse(product))
}

// handlerStoreProductsList lists products in a store with pagination
func (cfg *apiConfig) handlerStoreProductsList(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	storeHandle := chi.URLParam(r, "storeHandle")

	slog.InfoContext(r.Context(), "products list request",
		"request_id", reqID,
		"store_handle", storeHandle,
	)

	user, ok := userFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	store, err := cfg.getStoreAndVerifyAccess(r, storeHandle, user, "products:view")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Store not found", nil)
			return
		}
		if err.Error() == "permission denied" {
			respondWithError(w, http.StatusForbidden, "You do not have permission to view products in this store", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Unable to verify access", err)
		return
	}

	pageParams, err := ParsePageParams(r, 50, 100)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid pagination parameters", err)
		return
	}

	limit := pageParams.Limit
	limitPlusOne := int32(pageParams.Limit + 1)

	cursorCreatedAt, cursorID, hasCursor, err := cursorInfo(pageParams.Cursor)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid cursor", err)
		return
	}

	rows, err := cfg.db.GetProductsByStorePaginated(r.Context(), database.GetProductsByStorePaginatedParams{
		StoreID: store.ID,
		Column2: hasCursor,
		Column3: cursorCreatedAt,
		Column4: cursorID,
		Limit:   limitPlusOne,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve products", err)
		return
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	var nextCursor string
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		nextCursor, _ = productCursorCodec.Encode(ProductCursor{
			CreatedAt: last.CreatedAt,
			ID:        last.ID,
		})
	}

	response := make([]ProductResponse, 0, len(rows))
	for _, p := range rows {
		response = append(response, toProductResponseFromPaginatedRow(p))
	}

	slog.InfoContext(r.Context(), "products list successful",
		"request_id", reqID,
		"store_handle", storeHandle,
		"product_count", len(response),
	)

	respondWithJSON(w, http.StatusOK, map[string]any{
		"data": response,
		"page": map[string]any{
			"limit":       limit,
			"has_more":    hasMore,
			"next_cursor": nextCursor,
		},
	})
}

// handlerStoreProductGet retrieves a single product
func (cfg *apiConfig) handlerStoreProductGet(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	storeHandle := chi.URLParam(r, "storeHandle")
	productParam := chi.URLParam(r, "productID")

	user, ok := userFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	productID, err := uuid.Parse(productParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product ID format", err)
		return
	}

	store, err := cfg.getStoreAndVerifyAccess(r, storeHandle, user, "products:view")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Store not found", nil)
			return
		}
		if err.Error() == "permission denied" {
			respondWithError(w, http.StatusForbidden, "You do not have permission to view products in this store", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Unable to verify access", err)
		return
	}

	product, err := cfg.db.GetProductByID(r.Context(), database.GetProductByIDParams{
		ID:      productID,
		StoreID: store.ID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Product not found", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve product", err)
		return
	}

	slog.InfoContext(r.Context(), "product retrieved",
		"request_id", reqID,
		"store_handle", storeHandle,
		"product_id", product.ID,
	)

	respondWithJSON(w, http.StatusOK, toProductResponse(product))
}

// handlerStoreProductUpdate updates a product
func (cfg *apiConfig) handlerStoreProductUpdate(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	storeHandle := chi.URLParam(r, "storeHandle")
	productParam := chi.URLParam(r, "productID")

	user, ok := userFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	productID, err := uuid.Parse(productParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product ID format", err)
		return
	}

	store, err := cfg.getStoreAndVerifyAccess(r, storeHandle, user, "products:edit")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Store not found", nil)
			return
		}
		if err.Error() == "permission denied" {
			respondWithError(w, http.StatusForbidden, "You do not have permission to edit products in this store", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Unable to verify access", err)
		return
	}

	// Get existing product
	existing, err := cfg.db.GetProductByID(r.Context(), database.GetProductByIDParams{
		ID:      productID,
		StoreID: store.ID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Product not found", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve product", err)
		return
	}

	type parameters struct {
		Name             *string `json:"name"`
		Handle           *string `json:"handle"`
		Description      *string `json:"description"`
		InventoryTracked *bool   `json:"inventory_tracked"`
		SKU              *string `json:"sku"`
		Tags             *string `json:"tags"`
		Status           *string `json:"status"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Merge with existing values
	name := existing.Name
	if params.Name != nil {
		name = *params.Name
	}

	handle := existing.Handle
	if params.Handle != nil {
		handle = *params.Handle
	}

	description := existing.Description
	if params.Description != nil {
		description = sql.NullString{String: *params.Description, Valid: true}
	}

	inventoryTracked := existing.InventoryTracked
	if params.InventoryTracked != nil {
		inventoryTracked = *params.InventoryTracked
	}

	sku := existing.Sku
	if params.SKU != nil {
		sku = sql.NullString{String: *params.SKU, Valid: true}
	}

	tags := existing.Tags
	if params.Tags != nil {
		tags = sql.NullString{String: *params.Tags, Valid: true}
	}

	status := existing.Status
	if params.Status != nil {
		status = *params.Status
	}

	product, err := cfg.db.UpdateProduct(r.Context(), database.UpdateProductParams{
		ID:               productID,
		StoreID:          store.ID,
		Handle:           handle,
		Name:             name,
		Description:      description,
		InventoryTracked: inventoryTracked,
		Sku:              sku,
		Tags:             tags,
		Status:           status,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "product update failed",
			"request_id", reqID,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to update product", err)
		return
	}

	slog.InfoContext(r.Context(), "product updated successfully",
		"request_id", reqID,
		"store_handle", storeHandle,
		"product_id", product.ID,
	)

	respondWithJSON(w, http.StatusOK, toProductResponse(product))
}

// handlerStoreProductDelete deletes a product
func (cfg *apiConfig) handlerStoreProductDelete(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	storeHandle := chi.URLParam(r, "storeHandle")
	productParam := chi.URLParam(r, "productID")

	user, ok := userFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	productID, err := uuid.Parse(productParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product ID format", err)
		return
	}

	store, err := cfg.getStoreAndVerifyAccess(r, storeHandle, user, "products:delete")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Store not found", nil)
			return
		}
		if err.Error() == "permission denied" {
			respondWithError(w, http.StatusForbidden, "You do not have permission to delete products in this store", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Unable to verify access", err)
		return
	}

	deleted, err := cfg.db.DeleteProduct(r.Context(), database.DeleteProductParams{
		ID:      productID,
		StoreID: store.ID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Product not found", nil)
			return
		}
		slog.ErrorContext(r.Context(), "product delete failed",
			"request_id", reqID,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to delete product", err)
		return
	}

	slog.InfoContext(r.Context(), "product deleted successfully",
		"request_id", reqID,
		"store_handle", storeHandle,
		"product_id", deleted.ID,
	)

	respondWithJSON(w, http.StatusOK, map[string]any{
		"message":    "Product deleted successfully",
		"product_id": deleted.ID,
	})
}

// Helper functions to convert database models to response types
func toProductResponse(p database.Product) ProductResponse {
	var desc, sku, tags *string
	var gidStr string
	if p.Description.Valid {
		desc = &p.Description.String
	}
	if p.Sku.Valid {
		sku = &p.Sku.String
	}
	if p.Tags.Valid {
		tags = &p.Tags.String
	}
	if p.Gid.Valid {
		gidStr = gid.ProductGID(uint64(p.Gid.Int64)).String()
	}

	return ProductResponse{
		ID:               p.ID,
		GID:              gidStr,
		StoreID:          p.StoreID,
		Handle:           p.Handle,
		Name:             p.Name,
		Description:      desc,
		InventoryTracked: p.InventoryTracked,
		SKU:              sku,
		Tags:             tags,
		Status:           p.Status,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
	}
}

// toProductResponseFromPaginatedRow converts a paginated row to ProductResponse
func toProductResponseFromPaginatedRow(p database.GetProductsByStorePaginatedRow) ProductResponse {
	var desc, sku, tags *string
	var gidStr string
	if p.Description.Valid {
		desc = &p.Description.String
	}
	if p.Sku.Valid {
		sku = &p.Sku.String
	}
	if p.Tags.Valid {
		tags = &p.Tags.String
	}
	if p.Gid.Valid {
		gidStr = gid.ProductGID(uint64(p.Gid.Int64)).String()
	}

	return ProductResponse{
		ID:               p.ID,
		GID:              gidStr,
		StoreID:          p.StoreID,
		Handle:           p.Handle,
		Name:             p.Name,
		Description:      desc,
		InventoryTracked: p.InventoryTracked,
		SKU:              sku,
		Tags:             tags,
		Status:           p.Status,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
	}
}

