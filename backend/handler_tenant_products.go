package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/dfodeker/terminus/internal/database"
	"github.com/dfodeker/terminus/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type TenantProductResponse struct {
	ID               uuid.UUID `json:"id"`
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

// handlerTenantProductCreate creates a product within a tenant's store
func (cfg *apiConfig) handlerTenantProductCreate(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	tenantParam := chi.URLParam(r, "tenantID")
	storeParam := chi.URLParam(r, "storeID")

	slog.InfoContext(r.Context(), "tenant product creation request received",
		"request_id", reqID,
		"tenant_param", tenantParam,
		"store_param", storeParam,
	)

	user, ok := userFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	tenantID, err := uuid.Parse(tenantParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid tenant ID format", err)
		return
	}

	storeID, err := uuid.Parse(storeParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid store ID format", err)
		return
	}

	// Verify user has permission to create products
	hasPermission, err := cfg.db.CheckUserHasPermission(r.Context(), database.CheckUserHasPermissionParams{
		TenantID: tenantID,
		UserID:   user,
		Key:      "products:create",
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "tenant product creation failed: error checking permissions",
			"request_id", reqID,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to verify permissions", err)
		return
	}

	if !hasPermission {
		slog.WarnContext(r.Context(), "tenant product creation denied: insufficient permissions",
			"request_id", reqID,
			"user_id", user,
			"tenant_id", tenantID,
		)
		respondWithError(w, http.StatusForbidden, "You do not have permission to create products", nil)
		return
	}

	// Verify store belongs to tenant
	store, err := cfg.db.GetStoreByTenantAndID(r.Context(), database.GetStoreByTenantAndIDParams{
		TenantID: uuid.NullUUID{UUID: tenantID, Valid: true},
		ID:       storeID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Store not found in this tenant", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Unable to verify store", err)
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
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Please provide a valid request body", err)
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
		slog.ErrorContext(r.Context(), "tenant product creation failed: database error",
			"request_id", reqID,
			"user_id", user,
			"tenant_id", tenantID,
			"store_id", storeID,
			"error", err,
		)
		respondWithError(w, http.StatusBadRequest, "Unable to create product", err)
		return
	}

	slog.InfoContext(r.Context(), "tenant product created successfully",
		"request_id", reqID,
		"user_id", user,
		"tenant_id", tenantID,
		"store_id", storeID,
		"product_id", product.ID,
	)

	var desc, skuPtr, tagsPtr *string
	if product.Description.Valid {
		desc = &product.Description.String
	}
	if product.Sku.Valid {
		skuPtr = &product.Sku.String
	}
	if product.Tags.Valid {
		tagsPtr = &product.Tags.String
	}

	respondWithJSON(w, http.StatusCreated, TenantProductResponse{
		ID:               product.ID,
		StoreID:          product.StoreID,
		Handle:           product.Handle,
		Name:             product.Name,
		Description:      desc,
		InventoryTracked: product.InventoryTracked,
		SKU:              skuPtr,
		Tags:             tagsPtr,
		Status:           product.Status,
		CreatedAt:        product.CreatedAt,
		UpdatedAt:        product.UpdatedAt,
	})
}

// handlerTenantProductGet retrieves a single product
func (cfg *apiConfig) handlerTenantProductGet(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	tenantParam := chi.URLParam(r, "tenantID")
	storeParam := chi.URLParam(r, "storeID")
	productParam := chi.URLParam(r, "productID")

	slog.InfoContext(r.Context(), "tenant product get request received",
		"request_id", reqID,
		"tenant_param", tenantParam,
		"store_param", storeParam,
		"product_param", productParam,
	)

	user, ok := userFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	tenantID, err := uuid.Parse(tenantParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid tenant ID format", err)
		return
	}

	storeID, err := uuid.Parse(storeParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid store ID format", err)
		return
	}

	productID, err := uuid.Parse(productParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product ID format", err)
		return
	}

	// Verify user has permission to view products
	hasPermission, err := cfg.db.CheckUserHasPermission(r.Context(), database.CheckUserHasPermissionParams{
		TenantID: tenantID,
		UserID:   user,
		Key:      "products:view",
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to verify permissions", err)
		return
	}

	if !hasPermission {
		respondWithError(w, http.StatusForbidden, "You do not have permission to view products", nil)
		return
	}

	// Verify store belongs to tenant
	_, err = cfg.db.GetStoreByTenantAndID(r.Context(), database.GetStoreByTenantAndIDParams{
		TenantID: uuid.NullUUID{UUID: tenantID, Valid: true},
		ID:       storeID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Store not found in this tenant", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Unable to verify store", err)
		return
	}

	product, err := cfg.db.GetProductByID(r.Context(), database.GetProductByIDParams{
		ID:      productID,
		StoreID: storeID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Product not found", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve product", err)
		return
	}

	var desc, skuPtr, tagsPtr *string
	if product.Description.Valid {
		desc = &product.Description.String
	}
	if product.Sku.Valid {
		skuPtr = &product.Sku.String
	}
	if product.Tags.Valid {
		tagsPtr = &product.Tags.String
	}

	respondWithJSON(w, http.StatusOK, TenantProductResponse{
		ID:               product.ID,
		StoreID:          product.StoreID,
		Handle:           product.Handle,
		Name:             product.Name,
		Description:      desc,
		InventoryTracked: product.InventoryTracked,
		SKU:              skuPtr,
		Tags:             tagsPtr,
		Status:           product.Status,
		CreatedAt:        product.CreatedAt,
		UpdatedAt:        product.UpdatedAt,
	})
}

// handlerTenantProductUpdate updates a product
func (cfg *apiConfig) handlerTenantProductUpdate(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	tenantParam := chi.URLParam(r, "tenantID")
	storeParam := chi.URLParam(r, "storeID")
	productParam := chi.URLParam(r, "productID")

	slog.InfoContext(r.Context(), "tenant product update request received",
		"request_id", reqID,
		"tenant_param", tenantParam,
		"store_param", storeParam,
		"product_param", productParam,
	)

	user, ok := userFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	tenantID, err := uuid.Parse(tenantParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid tenant ID format", err)
		return
	}

	storeID, err := uuid.Parse(storeParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid store ID format", err)
		return
	}

	productID, err := uuid.Parse(productParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product ID format", err)
		return
	}

	// Verify user has permission to edit products
	hasPermission, err := cfg.db.CheckUserHasPermission(r.Context(), database.CheckUserHasPermissionParams{
		TenantID: tenantID,
		UserID:   user,
		Key:      "products:edit",
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to verify permissions", err)
		return
	}

	if !hasPermission {
		respondWithError(w, http.StatusForbidden, "You do not have permission to edit products", nil)
		return
	}

	// Verify store belongs to tenant
	_, err = cfg.db.GetStoreByTenantAndID(r.Context(), database.GetStoreByTenantAndIDParams{
		TenantID: uuid.NullUUID{UUID: tenantID, Valid: true},
		ID:       storeID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Store not found in this tenant", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Unable to verify store", err)
		return
	}

	// Get existing product first
	existingProduct, err := cfg.db.GetProductByID(r.Context(), database.GetProductByIDParams{
		ID:      productID,
		StoreID: storeID,
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
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Please provide a valid request body", err)
		return
	}

	// Use existing values if not provided
	name := existingProduct.Name
	if params.Name != nil {
		name = *params.Name
	}

	handle := existingProduct.Handle
	if params.Handle != nil {
		handle = *params.Handle
	}

	description := existingProduct.Description
	if params.Description != nil {
		description = sql.NullString{String: *params.Description, Valid: true}
	}

	inventoryTracked := existingProduct.InventoryTracked
	if params.InventoryTracked != nil {
		inventoryTracked = *params.InventoryTracked
	}

	sku := existingProduct.Sku
	if params.SKU != nil {
		sku = sql.NullString{String: *params.SKU, Valid: true}
	}

	tags := existingProduct.Tags
	if params.Tags != nil {
		tags = sql.NullString{String: *params.Tags, Valid: true}
	}

	status := existingProduct.Status
	if params.Status != nil {
		status = *params.Status
	}

	product, err := cfg.db.UpdateProduct(r.Context(), database.UpdateProductParams{
		ID:               productID,
		StoreID:          storeID,
		Handle:           handle,
		Name:             name,
		Description:      description,
		InventoryTracked: inventoryTracked,
		Sku:              sku,
		Tags:             tags,
		Status:           status,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "tenant product update failed: database error",
			"request_id", reqID,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to update product", err)
		return
	}

	slog.InfoContext(r.Context(), "tenant product updated successfully",
		"request_id", reqID,
		"user_id", user,
		"tenant_id", tenantID,
		"store_id", storeID,
		"product_id", product.ID,
	)

	var desc, skuPtr, tagsPtr *string
	if product.Description.Valid {
		desc = &product.Description.String
	}
	if product.Sku.Valid {
		skuPtr = &product.Sku.String
	}
	if product.Tags.Valid {
		tagsPtr = &product.Tags.String
	}

	respondWithJSON(w, http.StatusOK, TenantProductResponse{
		ID:               product.ID,
		StoreID:          product.StoreID,
		Handle:           product.Handle,
		Name:             product.Name,
		Description:      desc,
		InventoryTracked: product.InventoryTracked,
		SKU:              skuPtr,
		Tags:             tagsPtr,
		Status:           product.Status,
		CreatedAt:        product.CreatedAt,
		UpdatedAt:        product.UpdatedAt,
	})
}

// handlerTenantProductDelete deletes a product
func (cfg *apiConfig) handlerTenantProductDelete(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	tenantParam := chi.URLParam(r, "tenantID")
	storeParam := chi.URLParam(r, "storeID")
	productParam := chi.URLParam(r, "productID")

	slog.InfoContext(r.Context(), "tenant product delete request received",
		"request_id", reqID,
		"tenant_param", tenantParam,
		"store_param", storeParam,
		"product_param", productParam,
	)

	user, ok := userFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	tenantID, err := uuid.Parse(tenantParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid tenant ID format", err)
		return
	}

	storeID, err := uuid.Parse(storeParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid store ID format", err)
		return
	}

	productID, err := uuid.Parse(productParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product ID format", err)
		return
	}

	// Verify user has permission to delete products
	hasPermission, err := cfg.db.CheckUserHasPermission(r.Context(), database.CheckUserHasPermissionParams{
		TenantID: tenantID,
		UserID:   user,
		Key:      "products:delete",
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to verify permissions", err)
		return
	}

	if !hasPermission {
		respondWithError(w, http.StatusForbidden, "You do not have permission to delete products", nil)
		return
	}

	// Verify store belongs to tenant
	_, err = cfg.db.GetStoreByTenantAndID(r.Context(), database.GetStoreByTenantAndIDParams{
		TenantID: uuid.NullUUID{UUID: tenantID, Valid: true},
		ID:       storeID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Store not found in this tenant", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Unable to verify store", err)
		return
	}

	// Delete the product (variants will be deleted via CASCADE)
	deletedProduct, err := cfg.db.DeleteProduct(r.Context(), database.DeleteProductParams{
		ID:      productID,
		StoreID: storeID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Product not found", nil)
			return
		}
		slog.ErrorContext(r.Context(), "tenant product delete failed: database error",
			"request_id", reqID,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to delete product", err)
		return
	}

	slog.InfoContext(r.Context(), "tenant product deleted successfully",
		"request_id", reqID,
		"user_id", user,
		"tenant_id", tenantID,
		"store_id", storeID,
		"product_id", deletedProduct.ID,
	)

	respondWithJSON(w, http.StatusOK, map[string]any{
		"message":    "Product deleted successfully",
		"product_id": deletedProduct.ID,
	})
}

// handlerTenantProductsList lists products in a store with pagination
func (cfg *apiConfig) handlerTenantProductsList(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	tenantParam := chi.URLParam(r, "tenantID")
	storeParam := chi.URLParam(r, "storeID")

	slog.InfoContext(r.Context(), "tenant products list request received",
		"request_id", reqID,
		"tenant_param", tenantParam,
		"store_param", storeParam,
	)

	user, ok := userFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	tenantID, err := uuid.Parse(tenantParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid tenant ID format", err)
		return
	}

	storeID, err := uuid.Parse(storeParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid store ID format", err)
		return
	}

	// Verify user has permission to view products
	hasPermission, err := cfg.db.CheckUserHasPermission(r.Context(), database.CheckUserHasPermissionParams{
		TenantID: tenantID,
		UserID:   user,
		Key:      "products:view",
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to verify permissions", err)
		return
	}

	if !hasPermission {
		respondWithError(w, http.StatusForbidden, "You do not have permission to view products", nil)
		return
	}

	// Verify store belongs to tenant
	_, err = cfg.db.GetStoreByTenantAndID(r.Context(), database.GetStoreByTenantAndIDParams{
		TenantID: uuid.NullUUID{UUID: tenantID, Valid: true},
		ID:       storeID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Store not found in this tenant", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Unable to verify store", err)
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
		StoreID: storeID,
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
		nextCursor, err = productCursorCodec.Encode(ProductCursor{
			CreatedAt: last.CreatedAt,
			ID:        last.ID,
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Unable to build pagination cursor", err)
			return
		}
	}

	response := make([]TenantProductResponse, 0, len(rows))
	for _, product := range rows {
		var desc, skuPtr, tagsPtr *string
		if product.Description.Valid {
			desc = &product.Description.String
		}
		if product.Sku.Valid {
			skuPtr = &product.Sku.String
		}
		if product.Tags.Valid {
			tagsPtr = &product.Tags.String
		}

		response = append(response, TenantProductResponse{
			ID:               product.ID,
			StoreID:          product.StoreID,
			Handle:           product.Handle,
			Name:             product.Name,
			Description:      desc,
			InventoryTracked: product.InventoryTracked,
			SKU:              skuPtr,
			Tags:             tagsPtr,
			Status:           product.Status,
			CreatedAt:        product.CreatedAt,
			UpdatedAt:        product.UpdatedAt,
		})
	}

	slog.InfoContext(r.Context(), "tenant products list successful",
		"request_id", reqID,
		"user_id", user,
		"tenant_id", tenantID,
		"store_id", storeID,
		"product_count", len(response),
		"has_more", hasMore,
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
