package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/dfodeker/terminus/internal/database"
	"github.com/dfodeker/terminus/middleware"
	"github.com/google/uuid"
)

// type ProductOption struct {
// 	Id        uint64   `json:"id,omitempty"`
// 	ProductId uint64   `json:"product_id,omitempty"`
// 	Name      string   `json:"name,omitempty"`
// 	Position  int      `json:"position,omitempty"`
// 	Values    []string `json:"values,omitempty"`
// }

type Product struct {
	Id          uuid.UUID `json:"id,omitempty"`
	Title       string    `json:"title,omitempty"`
	Description string    `json:"body_html,omitempty"`
	//Vendor         string     `json:"vendor,omitempty"`
	//ProductType    string     `json:"product_type,omitempty"`
	Handle    string    `json:"handle,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	//PublishedAt    *time.Time `json:"published_at,omitempty"`
	//PublishedScope string     `json:"published_scope,omitempty"`
	Tags   string `json:"tags,omitempty"`
	Status string `json:"status,omitempty"`
	//Options                        []ProductOption `json:"options,omitempty"`
	//Variants                       []Variant   `json:"variants,omitempty"`
	//Image                          Image       `json:"image,omitempty"`
	//Images                         []Image     `json:"images,omitempty"`
	//TemplateSuffix                 string      `json:"template_suffix,omitempty"`
	//MetafieldsGlobalTitleTag       string      `json:"metafields_global_title_tag,omitempty"`
	//MetafieldsGlobalDescriptionTag string      `json:"metafields_global_description_tag,omitempty"`
	//Metafields                     []Metafield `json:"metafields,omitempty"`
	//AdminGraphqlApiId              string      `json:"admin_graphql_api_id,omitempty"`
}

// type Variant struct {
// 	Id                   uint64                 `json:"id,omitempty"`
// 	ProductId            uint64                 `json:"product_id,omitempty"`
// 	Title                string                 `json:"title,omitempty"`
// 	Sku                  string                 `json:"sku,omitempty"`
// 	Position             int                    `json:"position,omitempty"`
// 	Grams                int                    `json:"grams,omitempty"`
// 	InventoryPolicy      VariantInventoryPolicy `json:"inventory_policy,omitempty"`
// 	Price                *decimal.Decimal       `json:"price,omitempty"`
// 	CompareAtPrice       *decimal.Decimal       `json:"compare_at_price,omitempty"`
// 	FulfillmentService   string                 `json:"fulfillment_service,omitempty"`
// 	InventoryManagement  string                 `json:"inventory_management,omitempty"`
// 	InventoryItemId      uint64                 `json:"inventory_item_id,omitempty"`
// 	Option1              string                 `json:"option1,omitempty"`
// 	Option2              string                 `json:"option2,omitempty"`
// 	Option3              string                 `json:"option3,omitempty"`
// 	CreatedAt            *time.Time             `json:"created_at,omitempty"`
// 	UpdatedAt            *time.Time             `json:"updated_at,omitempty"`
// 	Taxable              bool                   `json:"taxable,omitempty"`
// 	TaxCode              string                 `json:"tax_code,omitempty"`
// 	Barcode              string                 `json:"barcode,omitempty"`
// 	ImageId              uint64                 `json:"image_id,omitempty"`
// 	InventoryQuantity    int                    `json:"inventory_quantity,omitempty"`
// 	Weight               *decimal.Decimal       `json:"weight,omitempty"`
// 	WeightUnit           string                 `json:"weight_unit,omitempty"`
// 	OldInventoryQuantity int                    `json:"old_inventory_quantity,omitempty"`
// 	RequireShipping      bool                   `json:"requires_shipping"`
// 	AdminGraphqlApiId    string                 `json:"admin_graphql_api_id,omitempty"`
// 	Metafields           []Metafield            `json:"metafields,omitempty"`
// 	PresentmentPrices    []presentmentPrices    `json:"presentment_prices,omitempty"`
// }

func (cfg *apiConfig) handlerCreateProducts(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	slog.InfoContext(r.Context(), "creating resource : stores", "request_id", reqID)
	user, ok := userFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}
	_ = user
	type parameters struct {
		Name           string `json:"name"` //rename to title later
		Description    string `json:"description"`
		Handle         string `json:"handle"`
		Tags           string `json:"tags"`
		TrackInventory string `json:"track_inventory"`
		Status         string `json:"status"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error Decoding Params: %s", err)
		respondWithError(w, 400, "Please provide a valid request body", err)
		return
	}
	if params.Description == "" {
		params.Description = " "
	}
	product, err := cfg.db.CreateProduct(r.Context(), database.CreateProductParams{
		Name:        params.Name,
		Handle:      params.Handle,
		Description: sql.NullString{String: params.Description, Valid: true},
		Status:      params.Status,
	})
	if err != nil {
		log.Printf("Unable to create product: %s", err)
		respondWithError(w, http.StatusBadRequest, "Unable to create product", err)
		return
	}
	respondWithJSON(w, http.StatusCreated, Product{
		Id:          product.ID,
		Title:       product.Name,
		Description: product.Description.String,
		Handle:      product.Handle,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	})

}
