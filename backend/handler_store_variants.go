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

type StoreVariantResponse struct {
	ID             uuid.UUID       `json:"id"`
	TenantID       uuid.UUID       `json:"tenant_id"`
	StoreID        uuid.UUID       `json:"store_id"`
	ProductID      uuid.UUID       `json:"product_id"`
	SKU            *string         `json:"sku,omitempty"`
	Barcode        *string         `json:"barcode,omitempty"`
	Title          string          `json:"title"`
	PriceCents     int32           `json:"price_cents"`
	CompareAtCents *int32          `json:"compare_at_cents,omitempty"`
	OptionValues   json.RawMessage `json:"option_values"`
	Status         string          `json:"status"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// getProductAndVerifyAccess looks up product by ID and verifies user has the required permission
func (cfg *apiConfig) getProductAndVerifyAccess(r *http.Request, productID uuid.UUID, userID uuid.UUID, permissionKey string) (database.Product, database.Store, error) {
	// Get product first
	product, err := cfg.db.GetProductByIDOnly(r.Context(), productID)
	if err != nil {
		return database.Product{}, database.Store{}, err
	}

	// Get the store to find tenant
	store, err := cfg.db.GetStoreByTenantAndID(r.Context(), database.GetStoreByTenantAndIDParams{
		ID:       product.StoreID,
		TenantID: uuid.NullUUID{Valid: false}, // We don't know tenant yet, will check after
	})
	if err != nil {
		// Try getting store without tenant filter
		stores, err := cfg.db.GetStores(r.Context())
		if err != nil {
			return database.Product{}, database.Store{}, err
		}
		for _, s := range stores {
			if s.ID == product.StoreID {
				store = s
				break
			}
		}
		if store.ID == uuid.Nil {
			return database.Product{}, database.Store{}, sql.ErrNoRows
		}
	}

	if !store.TenantID.Valid {
		return database.Product{}, database.Store{}, errors.New("store has no tenant association")
	}

	// Check user has permission in this tenant
	hasPermission, err := cfg.db.CheckUserHasPermission(r.Context(), database.CheckUserHasPermissionParams{
		TenantID: store.TenantID.UUID,
		UserID:   userID,
		Key:      permissionKey,
	})
	if err != nil {
		return database.Product{}, database.Store{}, err
	}

	if !hasPermission {
		return database.Product{}, database.Store{}, errors.New("permission denied")
	}

	return product, store, nil
}

// getVariantAndVerifyAccess looks up variant and verifies user has the required permission
func (cfg *apiConfig) getVariantAndVerifyAccess(r *http.Request, variantID uuid.UUID, userID uuid.UUID, permissionKey string) (database.ProductVariant, error) {
	variant, err := cfg.db.GetProductVariantByID(r.Context(), variantID)
	if err != nil {
		return database.ProductVariant{}, err
	}

	// Check user has permission in this tenant
	hasPermission, err := cfg.db.CheckUserHasPermission(r.Context(), database.CheckUserHasPermissionParams{
		TenantID: variant.TenantID,
		UserID:   userID,
		Key:      permissionKey,
	})
	if err != nil {
		return database.ProductVariant{}, err
	}

	if !hasPermission {
		return database.ProductVariant{}, errors.New("permission denied")
	}

	return variant, nil
}

// POST /api/v1/products/{productID}/variants
func (cfg *apiConfig) handlerProductVariantCreate(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	productParam := chi.URLParam(r, "productID")

	slog.InfoContext(r.Context(), "variant creation request",
		"request_id", reqID,
		"product_param", productParam,
	)

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

	product, store, err := cfg.getProductAndVerifyAccess(r, productID, user, "products:create")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Product not found", nil)
			return
		}
		if err.Error() == "permission denied" {
			respondWithError(w, http.StatusForbidden, "You do not have permission to create variants", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Unable to verify access", err)
		return
	}

	type parameters struct {
		SKU            *string         `json:"sku"`
		Barcode        *string         `json:"barcode"`
		Title          string          `json:"title"`
		PriceCents     int32           `json:"price_cents"`
		CompareAtCents *int32          `json:"compare_at_cents"`
		OptionValues   json.RawMessage `json:"option_values"`
		Status         string          `json:"status"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if params.PriceCents < 0 {
		respondWithError(w, http.StatusBadRequest, "Price must be non-negative", nil)
		return
	}

	title := params.Title
	if title == "" {
		title = "Default Title"
	}

	status := params.Status
	if status == "" {
		status = "active"
	}

	optionValues := params.OptionValues
	if optionValues == nil {
		optionValues = json.RawMessage("{}")
	}

	sku := sql.NullString{}
	if params.SKU != nil {
		sku = sql.NullString{String: *params.SKU, Valid: true}
	}

	barcode := sql.NullString{}
	if params.Barcode != nil {
		barcode = sql.NullString{String: *params.Barcode, Valid: true}
	}

	compareAtCents := sql.NullInt32{}
	if params.CompareAtCents != nil {
		compareAtCents = sql.NullInt32{Int32: *params.CompareAtCents, Valid: true}
	}

	variant, err := cfg.db.CreateProductVariant(r.Context(), database.CreateProductVariantParams{
		TenantID:       store.TenantID.UUID,
		StoreID:        store.ID,
		ProductID:      product.ID,
		Sku:            sku,
		Barcode:        barcode,
		Title:          title,
		PriceCents:     params.PriceCents,
		CompareAtCents: compareAtCents,
		OptionValues:   optionValues,
		Status:         status,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "variant creation failed",
			"request_id", reqID,
			"error", err,
		)
		respondWithError(w, http.StatusBadRequest, "Unable to create variant", err)
		return
	}

	slog.InfoContext(r.Context(), "variant created successfully",
		"request_id", reqID,
		"product_id", productID,
		"variant_id", variant.ID,
	)

	respondWithJSON(w, http.StatusCreated, toVariantResponse(variant))
}

// GET /api/v1/products/{productID}/variants
func (cfg *apiConfig) handlerProductVariantsList(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
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

	_, _, err = cfg.getProductAndVerifyAccess(r, productID, user, "products:view")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Product not found", nil)
			return
		}
		if err.Error() == "permission denied" {
			respondWithError(w, http.StatusForbidden, "You do not have permission to view variants", nil)
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

	cursorCreatedAt, cursorID, hasCursor, err := decodeVariantCursor(pageParams.Cursor)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid cursor", err)
		return
	}

	rows, err := cfg.db.GetProductVariantsByProductIDPaginated(r.Context(), database.GetProductVariantsByProductIDPaginatedParams{
		ProductID: productID,
		Column2:   hasCursor,
		Column3:   cursorCreatedAt,
		Column4:   cursorID,
		Limit:     limitPlusOne,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve variants", err)
		return
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	var nextCursor string
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		nextCursor, _ = variantCursorCodec.Encode(VariantCursor{
			CreatedAt: last.CreatedAt,
			ID:        last.ID,
		})
	}

	response := make([]StoreVariantResponse, 0, len(rows))
	for _, v := range rows {
		response = append(response, toVariantResponse(v))
	}

	slog.InfoContext(r.Context(), "variants list successful",
		"request_id", reqID,
		"product_id", productID,
		"variant_count", len(response),
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

// GET /api/v1/variants/{variantID}
func (cfg *apiConfig) handlerVariantGet(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	variantParam := chi.URLParam(r, "variantID")

	user, ok := userFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	variantID, err := uuid.Parse(variantParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid variant ID format", err)
		return
	}

	variant, err := cfg.getVariantAndVerifyAccess(r, variantID, user, "products:view")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Variant not found", nil)
			return
		}
		if err.Error() == "permission denied" {
			respondWithError(w, http.StatusForbidden, "You do not have permission to view this variant", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Unable to verify access", err)
		return
	}

	slog.InfoContext(r.Context(), "variant retrieved",
		"request_id", reqID,
		"variant_id", variant.ID,
	)

	respondWithJSON(w, http.StatusOK, toVariantResponse(variant))
}

// PUT /api/v1/variants/{variantID}
func (cfg *apiConfig) handlerVariantUpdate(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	variantParam := chi.URLParam(r, "variantID")

	user, ok := userFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	variantID, err := uuid.Parse(variantParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid variant ID format", err)
		return
	}

	existing, err := cfg.getVariantAndVerifyAccess(r, variantID, user, "products:edit")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Variant not found", nil)
			return
		}
		if err.Error() == "permission denied" {
			respondWithError(w, http.StatusForbidden, "You do not have permission to edit this variant", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Unable to verify access", err)
		return
	}

	type parameters struct {
		SKU            *string          `json:"sku"`
		Barcode        *string          `json:"barcode"`
		Title          *string          `json:"title"`
		PriceCents     *int32           `json:"price_cents"`
		CompareAtCents *int32           `json:"compare_at_cents"`
		OptionValues   *json.RawMessage `json:"option_values"`
		Status         *string          `json:"status"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Merge with existing values
	sku := existing.Sku
	if params.SKU != nil {
		sku = sql.NullString{String: *params.SKU, Valid: true}
	}

	barcode := existing.Barcode
	if params.Barcode != nil {
		barcode = sql.NullString{String: *params.Barcode, Valid: true}
	}

	title := existing.Title
	if params.Title != nil {
		title = *params.Title
	}

	priceCents := existing.PriceCents
	if params.PriceCents != nil {
		priceCents = *params.PriceCents
	}

	compareAtCents := existing.CompareAtCents
	if params.CompareAtCents != nil {
		compareAtCents = sql.NullInt32{Int32: *params.CompareAtCents, Valid: true}
	}

	optionValues := existing.OptionValues
	if params.OptionValues != nil {
		optionValues = *params.OptionValues
	}

	status := existing.Status
	if params.Status != nil {
		status = *params.Status
	}

	variant, err := cfg.db.UpdateProductVariant(r.Context(), database.UpdateProductVariantParams{
		ID:             variantID,
		ProductID:      existing.ProductID,
		Sku:            sku,
		Barcode:        barcode,
		Title:          title,
		PriceCents:     priceCents,
		CompareAtCents: compareAtCents,
		OptionValues:   optionValues,
		Status:         status,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "variant update failed",
			"request_id", reqID,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to update variant", err)
		return
	}

	slog.InfoContext(r.Context(), "variant updated successfully",
		"request_id", reqID,
		"variant_id", variant.ID,
	)

	respondWithJSON(w, http.StatusOK, toVariantResponse(variant))
}

// DELETE /api/v1/variants/{variantID}
func (cfg *apiConfig) handlerVariantDelete(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	variantParam := chi.URLParam(r, "variantID")

	user, ok := userFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	variantID, err := uuid.Parse(variantParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid variant ID format", err)
		return
	}

	existing, err := cfg.getVariantAndVerifyAccess(r, variantID, user, "products:delete")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Variant not found", nil)
			return
		}
		if err.Error() == "permission denied" {
			respondWithError(w, http.StatusForbidden, "You do not have permission to delete this variant", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Unable to verify access", err)
		return
	}

	err = cfg.db.DeleteProductVariant(r.Context(), database.DeleteProductVariantParams{
		ID:        variantID,
		ProductID: existing.ProductID,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "variant delete failed",
			"request_id", reqID,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to delete variant", err)
		return
	}

	slog.InfoContext(r.Context(), "variant deleted successfully",
		"request_id", reqID,
		"variant_id", variantID,
	)

	respondWithJSON(w, http.StatusOK, map[string]any{
		"message":    "Variant deleted successfully",
		"variant_id": variantID,
	})
}

func toVariantResponse(v database.ProductVariant) StoreVariantResponse {
	var sku, barcode *string
	var compareAt *int32

	if v.Sku.Valid {
		sku = &v.Sku.String
	}
	if v.Barcode.Valid {
		barcode = &v.Barcode.String
	}
	if v.CompareAtCents.Valid {
		compareAt = &v.CompareAtCents.Int32
	}

	return StoreVariantResponse{
		ID:             v.ID,
		TenantID:       v.TenantID,
		StoreID:        v.StoreID,
		ProductID:      v.ProductID,
		SKU:            sku,
		Barcode:        barcode,
		Title:          v.Title,
		PriceCents:     v.PriceCents,
		CompareAtCents: compareAt,
		OptionValues:   v.OptionValues,
		Status:         v.Status,
		CreatedAt:      v.CreatedAt,
		UpdatedAt:      v.UpdatedAt,
	}
}
