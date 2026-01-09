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

type VariantResponse struct {
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

type VariantCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        uuid.UUID `json:"id"`
}

var variantCursorCodec = CursorCodec[VariantCursor]{
	Validate: func(c VariantCursor) error {
		if c.CreatedAt.IsZero() || c.ID == uuid.Nil {
			return errors.New("invalid cursor: missing required fields")
		}
		return nil
	},
}

// handlerTenantVariantCreate creates a variant for a product
func (cfg *apiConfig) handlerTenantVariantCreate(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	tenantParam := chi.URLParam(r, "tenantID")
	storeParam := chi.URLParam(r, "storeID")
	productParam := chi.URLParam(r, "productID")

	slog.InfoContext(r.Context(), "tenant variant creation request received",
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

	// Verify user has permission to create products (variants are part of products)
	hasPermission, err := cfg.db.CheckUserHasPermission(r.Context(), database.CheckUserHasPermissionParams{
		TenantID: tenantID,
		UserID:   user,
		Key:      "products:create",
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to verify permissions", err)
		return
	}

	if !hasPermission {
		respondWithError(w, http.StatusForbidden, "You do not have permission to create variants", nil)
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

	// Verify product exists and belongs to store
	_, err = cfg.db.GetProductByID(r.Context(), database.GetProductByIDParams{
		ID:      productID,
		StoreID: storeID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Product not found", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Unable to verify product", err)
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
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Please provide a valid request body", err)
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
		TenantID:       tenantID,
		StoreID:        storeID,
		ProductID:      productID,
		Sku:            sku,
		Barcode:        barcode,
		Title:          title,
		PriceCents:     params.PriceCents,
		CompareAtCents: compareAtCents,
		OptionValues:   optionValues,
		Status:         status,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "tenant variant creation failed: database error",
			"request_id", reqID,
			"error", err,
		)
		respondWithError(w, http.StatusBadRequest, "Unable to create variant", err)
		return
	}

	slog.InfoContext(r.Context(), "tenant variant created successfully",
		"request_id", reqID,
		"user_id", user,
		"tenant_id", tenantID,
		"store_id", storeID,
		"product_id", productID,
		"variant_id", variant.ID,
	)

	var skuPtr, barcodePtr *string
	var compareAtPtr *int32
	if variant.Sku.Valid {
		skuPtr = &variant.Sku.String
	}
	if variant.Barcode.Valid {
		barcodePtr = &variant.Barcode.String
	}
	if variant.CompareAtCents.Valid {
		compareAtPtr = &variant.CompareAtCents.Int32
	}

	respondWithJSON(w, http.StatusCreated, VariantResponse{
		ID:             variant.ID,
		TenantID:       variant.TenantID,
		StoreID:        variant.StoreID,
		ProductID:      variant.ProductID,
		SKU:            skuPtr,
		Barcode:        barcodePtr,
		Title:          variant.Title,
		PriceCents:     variant.PriceCents,
		CompareAtCents: compareAtPtr,
		OptionValues:   variant.OptionValues,
		Status:         variant.Status,
		CreatedAt:      variant.CreatedAt,
		UpdatedAt:      variant.UpdatedAt,
	})
}

// handlerTenantVariantsList lists variants for a product
func (cfg *apiConfig) handlerTenantVariantsList(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	tenantParam := chi.URLParam(r, "tenantID")
	storeParam := chi.URLParam(r, "storeID")
	productParam := chi.URLParam(r, "productID")

	slog.InfoContext(r.Context(), "tenant variants list request received",
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
		respondWithError(w, http.StatusForbidden, "You do not have permission to view variants", nil)
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
		nextCursor, err = variantCursorCodec.Encode(VariantCursor{
			CreatedAt: last.CreatedAt,
			ID:        last.ID,
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Unable to build pagination cursor", err)
			return
		}
	}

	response := make([]VariantResponse, 0, len(rows))
	for _, variant := range rows {
		var skuPtr, barcodePtr *string
		var compareAtPtr *int32
		if variant.Sku.Valid {
			skuPtr = &variant.Sku.String
		}
		if variant.Barcode.Valid {
			barcodePtr = &variant.Barcode.String
		}
		if variant.CompareAtCents.Valid {
			compareAtPtr = &variant.CompareAtCents.Int32
		}

		response = append(response, VariantResponse{
			ID:             variant.ID,
			TenantID:       variant.TenantID,
			StoreID:        variant.StoreID,
			ProductID:      variant.ProductID,
			SKU:            skuPtr,
			Barcode:        barcodePtr,
			Title:          variant.Title,
			PriceCents:     variant.PriceCents,
			CompareAtCents: compareAtPtr,
			OptionValues:   variant.OptionValues,
			Status:         variant.Status,
			CreatedAt:      variant.CreatedAt,
			UpdatedAt:      variant.UpdatedAt,
		})
	}

	slog.InfoContext(r.Context(), "tenant variants list successful",
		"request_id", reqID,
		"user_id", user,
		"tenant_id", tenantID,
		"product_id", productID,
		"variant_count", len(response),
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

// handlerTenantVariantUpdate updates a variant
func (cfg *apiConfig) handlerTenantVariantUpdate(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	tenantParam := chi.URLParam(r, "tenantID")
	storeParam := chi.URLParam(r, "storeID")
	productParam := chi.URLParam(r, "productID")
	variantParam := chi.URLParam(r, "variantID")

	slog.InfoContext(r.Context(), "tenant variant update request received",
		"request_id", reqID,
		"tenant_param", tenantParam,
		"store_param", storeParam,
		"product_param", productParam,
		"variant_param", variantParam,
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

	variantID, err := uuid.Parse(variantParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid variant ID format", err)
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
		respondWithError(w, http.StatusForbidden, "You do not have permission to edit variants", nil)
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

	// Get existing variant
	existingVariant, err := cfg.db.GetProductVariantByID(r.Context(), variantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Variant not found", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve variant", err)
		return
	}

	// Verify variant belongs to the correct product
	if existingVariant.ProductID != productID {
		respondWithError(w, http.StatusNotFound, "Variant not found for this product", nil)
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
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Please provide a valid request body", err)
		return
	}

	// Use existing values if not provided
	sku := existingVariant.Sku
	if params.SKU != nil {
		sku = sql.NullString{String: *params.SKU, Valid: true}
	}

	barcode := existingVariant.Barcode
	if params.Barcode != nil {
		barcode = sql.NullString{String: *params.Barcode, Valid: true}
	}

	title := existingVariant.Title
	if params.Title != nil {
		title = *params.Title
	}

	priceCents := existingVariant.PriceCents
	if params.PriceCents != nil {
		priceCents = *params.PriceCents
	}

	compareAtCents := existingVariant.CompareAtCents
	if params.CompareAtCents != nil {
		compareAtCents = sql.NullInt32{Int32: *params.CompareAtCents, Valid: true}
	}

	optionValues := existingVariant.OptionValues
	if params.OptionValues != nil {
		optionValues = *params.OptionValues
	}

	status := existingVariant.Status
	if params.Status != nil {
		status = *params.Status
	}

	variant, err := cfg.db.UpdateProductVariant(r.Context(), database.UpdateProductVariantParams{
		ID:             variantID,
		ProductID:      productID,
		Sku:            sku,
		Barcode:        barcode,
		Title:          title,
		PriceCents:     priceCents,
		CompareAtCents: compareAtCents,
		OptionValues:   optionValues,
		Status:         status,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "tenant variant update failed: database error",
			"request_id", reqID,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to update variant", err)
		return
	}

	slog.InfoContext(r.Context(), "tenant variant updated successfully",
		"request_id", reqID,
		"user_id", user,
		"tenant_id", tenantID,
		"variant_id", variant.ID,
	)

	var skuPtr, barcodePtr *string
	var compareAtPtr *int32
	if variant.Sku.Valid {
		skuPtr = &variant.Sku.String
	}
	if variant.Barcode.Valid {
		barcodePtr = &variant.Barcode.String
	}
	if variant.CompareAtCents.Valid {
		compareAtPtr = &variant.CompareAtCents.Int32
	}

	respondWithJSON(w, http.StatusOK, VariantResponse{
		ID:             variant.ID,
		TenantID:       variant.TenantID,
		StoreID:        variant.StoreID,
		ProductID:      variant.ProductID,
		SKU:            skuPtr,
		Barcode:        barcodePtr,
		Title:          variant.Title,
		PriceCents:     variant.PriceCents,
		CompareAtCents: compareAtPtr,
		OptionValues:   variant.OptionValues,
		Status:         variant.Status,
		CreatedAt:      variant.CreatedAt,
		UpdatedAt:      variant.UpdatedAt,
	})
}

// handlerTenantVariantDelete deletes a variant
func (cfg *apiConfig) handlerTenantVariantDelete(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	tenantParam := chi.URLParam(r, "tenantID")
	storeParam := chi.URLParam(r, "storeID")
	productParam := chi.URLParam(r, "productID")
	variantParam := chi.URLParam(r, "variantID")

	slog.InfoContext(r.Context(), "tenant variant delete request received",
		"request_id", reqID,
		"tenant_param", tenantParam,
		"store_param", storeParam,
		"product_param", productParam,
		"variant_param", variantParam,
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

	variantID, err := uuid.Parse(variantParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid variant ID format", err)
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
		respondWithError(w, http.StatusForbidden, "You do not have permission to delete variants", nil)
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

	err = cfg.db.DeleteProductVariant(r.Context(), database.DeleteProductVariantParams{
		ID:        variantID,
		ProductID: productID,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "tenant variant delete failed: database error",
			"request_id", reqID,
			"error", err,
		)
		respondWithError(w, http.StatusInternalServerError, "Unable to delete variant", err)
		return
	}

	slog.InfoContext(r.Context(), "tenant variant deleted successfully",
		"request_id", reqID,
		"user_id", user,
		"tenant_id", tenantID,
		"variant_id", variantID,
	)

	respondWithJSON(w, http.StatusOK, map[string]any{
		"message":    "Variant deleted successfully",
		"variant_id": variantID,
	})
}

func decodeVariantCursor(cursor string) (time.Time, uuid.UUID, bool, error) {
	cur, ok, err := variantCursorCodec.Decode(cursor)
	if err != nil {
		return time.Time{}, uuid.UUID{}, false, err
	}
	if !ok {
		return time.Time{}, uuid.UUID{}, false, nil
	}
	return cur.CreatedAt, cur.ID, true, nil
}
