package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestProduct_CanPurchase(t *testing.T) {
	tests := []struct {
		name      string
		product   *Product
		quantity  int
		wantErr   error
		wantNil   bool
	}{
		{
			name: "active product with sufficient inventory",
			product: &Product{
				ID:        ProductID(uuid.New()),
				ShopID:    ShopID(uuid.New()),
				Title:     "Test Product",
				Status:    ProductStatusActive,
				Inventory: 10,
			},
			quantity: 5,
			wantNil:  true,
		},
		{
			name: "active product with exact inventory",
			product: &Product{
				ID:        ProductID(uuid.New()),
				ShopID:    ShopID(uuid.New()),
				Title:     "Test Product",
				Status:    ProductStatusActive,
				Inventory: 5,
			},
			quantity: 5,
			wantNil:  true,
		},
		{
			name: "active product with insufficient inventory",
			product: &Product{
				ID:        ProductID(uuid.New()),
				ShopID:    ShopID(uuid.New()),
				Title:     "Test Product",
				Status:    ProductStatusActive,
				Inventory: 3,
			},
			quantity: 5,
			wantErr:  ErrInsufficientInventory,
		},
		{
			name: "draft product",
			product: &Product{
				ID:        ProductID(uuid.New()),
				ShopID:    ShopID(uuid.New()),
				Title:     "Test Product",
				Status:    ProductStatusDraft,
				Inventory: 10,
			},
			quantity: 1,
			wantErr:  ErrProductNotAvailable,
		},
		{
			name: "archived product",
			product: &Product{
				ID:        ProductID(uuid.New()),
				ShopID:    ShopID(uuid.New()),
				Title:     "Test Product",
				Status:    ProductStatusArchived,
				Inventory: 10,
			},
			quantity: 1,
			wantErr:  ErrProductNotAvailable,
		},
		{
			name: "zero quantity request on active product",
			product: &Product{
				ID:        ProductID(uuid.New()),
				ShopID:    ShopID(uuid.New()),
				Title:     "Test Product",
				Status:    ProductStatusActive,
				Inventory: 10,
			},
			quantity: 0,
			wantNil:  true,
		},
		{
			name: "zero inventory",
			product: &Product{
				ID:        ProductID(uuid.New()),
				ShopID:    ShopID(uuid.New()),
				Title:     "Test Product",
				Status:    ProductStatusActive,
				Inventory: 0,
			},
			quantity: 1,
			wantErr:  ErrInsufficientInventory,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.product.CanPurchase(tt.quantity)

			if tt.wantNil {
				if err != nil {
					t.Errorf("CanPurchase() error = %v, want nil", err)
				}
				return
			}

			if err == nil {
				t.Errorf("CanPurchase() error = nil, want %v", tt.wantErr)
				return
			}

			// Check error type using errors.Is behavior via DomainError.Is
			domainErr, ok := err.(*DomainError)
			if !ok {
				t.Errorf("CanPurchase() error is not a DomainError")
				return
			}

			wantDomainErr, ok := tt.wantErr.(*DomainError)
			if !ok {
				t.Errorf("wantErr is not a DomainError")
				return
			}

			if domainErr.Code != wantDomainErr.Code {
				t.Errorf("CanPurchase() error code = %s, want %s", domainErr.Code, wantDomainErr.Code)
			}
		})
	}
}

func TestProductStatus_Constants(t *testing.T) {
	tests := []struct {
		status ProductStatus
		want   string
	}{
		{ProductStatusDraft, "draft"},
		{ProductStatusActive, "active"},
		{ProductStatusArchived, "archived"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("ProductStatus = %s, want %s", tt.status, tt.want)
			}
		})
	}
}

func TestProduct_Fields(t *testing.T) {
	now := time.Now()
	productID := ProductID(uuid.New())
	shopID := ShopID(uuid.New())

	product := &Product{
		ID:          productID,
		ShopID:      shopID,
		Title:       "Test Product",
		Description: "A test product description",
		Handle:      "test-product",
		BodyHTML:    "<p>Product body HTML</p>",
		PriceCents:  Money(1999),
		SKU:         "TEST-SKU-001",
		Inventory:   100,
		Status:      ProductStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if product.ID != productID {
		t.Errorf("Product.ID = %v, want %v", product.ID, productID)
	}
	if product.ShopID != shopID {
		t.Errorf("Product.ShopID = %v, want %v", product.ShopID, shopID)
	}
	if product.Title != "Test Product" {
		t.Errorf("Product.Title = %s, want %s", product.Title, "Test Product")
	}
	if product.Description != "A test product description" {
		t.Errorf("Product.Description = %s, want %s", product.Description, "A test product description")
	}
	if product.Handle != "test-product" {
		t.Errorf("Product.Handle = %s, want %s", product.Handle, "test-product")
	}
	if product.BodyHTML != "<p>Product body HTML</p>" {
		t.Errorf("Product.BodyHTML = %s, want %s", product.BodyHTML, "<p>Product body HTML</p>")
	}
	if product.PriceCents != Money(1999) {
		t.Errorf("Product.PriceCents = %d, want %d", product.PriceCents, 1999)
	}
	if product.SKU != "TEST-SKU-001" {
		t.Errorf("Product.SKU = %s, want %s", product.SKU, "TEST-SKU-001")
	}
	if product.Inventory != 100 {
		t.Errorf("Product.Inventory = %d, want %d", product.Inventory, 100)
	}
	if product.Status != ProductStatusActive {
		t.Errorf("Product.Status = %s, want %s", product.Status, ProductStatusActive)
	}
}

func TestMoney_Type(t *testing.T) {
	price := Money(1999)
	if int64(price) != 1999 {
		t.Errorf("Money value = %d, want %d", price, 1999)
	}
}
