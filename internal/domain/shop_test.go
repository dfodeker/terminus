package domain

import (
	"strings"
	"testing"
	"time"
)

func TestNewShop(t *testing.T) {
	shop := NewShop(123, "Test Shop", "test@example.com")

	if shop.Id != 123 {
		t.Errorf("expected Id to be 123, got %d", shop.Id)
	}
	if shop.Name != "Test Shop" {
		t.Errorf("expected Name to be 'Test Shop', got %s", shop.Name)
	}
	if shop.Email != "test@example.com" {
		t.Errorf("expected Email to be 'test@example.com', got %s", shop.Email)
	}
	if shop.CreatedAt == nil {
		t.Error("expected CreatedAt to be set")
	}
	if shop.UpdatedAt == nil {
		t.Error("expected UpdatedAt to be set")
	}
	if !shop.ForceSSL {
		t.Error("expected ForceSSL to be true by default")
	}
}

func TestShop_Validate(t *testing.T) {
	tests := []struct {
		name    string
		shop    Shop
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid shop with domain",
			shop: Shop{
				Id:     1,
				Name:   "Test Shop",
				Email:  "test@example.com",
				Domain: "testshop.com",
			},
			wantErr: false,
		},
		{
			name: "valid shop with mystoreos domain",
			shop: Shop{
				Id:              1,
				Name:            "Test Shop",
				Email:           "test@example.com",
				MyShopifyDomain: "testshop.mystoreos.com",
			},
			wantErr: false,
		},
		{
			name: "missing id",
			shop: Shop{
				Name:   "Test Shop",
				Email:  "test@example.com",
				Domain: "testshop.com",
			},
			wantErr: true,
			errMsg:  "id is required",
		},
		{
			name: "missing name",
			shop: Shop{
				Id:     1,
				Email:  "test@example.com",
				Domain: "testshop.com",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "missing email",
			shop: Shop{
				Id:     1,
				Name:   "Test Shop",
				Domain: "testshop.com",
			},
			wantErr: true,
			errMsg:  "email is required",
		},
		{
			name: "missing both domains",
			shop: Shop{
				Id:    1,
				Name:  "Test Shop",
				Email: "test@example.com",
			},
			wantErr: true,
			errMsg:  "domain or mystoreos_domain is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.shop.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Validate() error = %v, should contain %s", err, tt.errMsg)
			}
		})
	}
}

func TestShop_FullAddress(t *testing.T) {
	tests := []struct {
		name string
		shop Shop
		want string
	}{
		{
			name: "complete address",
			shop: Shop{
				Address1:    "123 Main St",
				Address2:    "Suite 100",
				City:        "New York",
				Province:    "NY",
				Zip:         "10001",
				CountryName: "United States",
			},
			want: "123 Main St, Suite 100, New York, NY, 10001, United States",
		},
		{
			name: "address without address2",
			shop: Shop{
				Address1:    "123 Main St",
				City:        "New York",
				Province:    "NY",
				Zip:         "10001",
				CountryName: "United States",
			},
			want: "123 Main St, New York, NY, 10001, United States",
		},
		{
			name: "address with country fallback",
			shop: Shop{
				Address1: "123 Main St",
				City:     "New York",
				Country:  "US",
			},
			want: "123 Main St, New York, US",
		},
		{
			name: "empty address",
			shop: Shop{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.shop.FullAddress()
			if got != tt.want {
				t.Errorf("FullAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShop_PrimaryDomain(t *testing.T) {
	tests := []struct {
		name string
		shop Shop
		want string
	}{
		{
			name: "prefers custom domain",
			shop: Shop{
				Domain:          "mystore.com",
				MyShopifyDomain: "mystore.myshopify.com",
			},
			want: "mystore.com",
		},
		{
			name: "falls back to myshopify domain",
			shop: Shop{
				MyShopifyDomain: "mystore.myshopify.com",
			},
			want: "mystore.myshopify.com",
		},
		{
			name: "empty domains",
			shop: Shop{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.shop.PrimaryDomain()
			if got != tt.want {
				t.Errorf("PrimaryDomain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShop_HasCoordinates(t *testing.T) {
	tests := []struct {
		name string
		shop Shop
		want bool
	}{
		{
			name: "has both coordinates",
			shop: Shop{Latitude: 40.7128, Longitude: -74.0060},
			want: true,
		},
		{
			name: "has only latitude",
			shop: Shop{Latitude: 40.7128},
			want: true,
		},
		{
			name: "has only longitude",
			shop: Shop{Longitude: -74.0060},
			want: true,
		},
		{
			name: "no coordinates",
			shop: Shop{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.shop.HasCoordinates()
			if got != tt.want {
				t.Errorf("HasCoordinates() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShop_Coordinates(t *testing.T) {
	shop := Shop{Latitude: 40.7128, Longitude: -74.0060}
	lat, lng := shop.Coordinates()

	if lat != 40.7128 {
		t.Errorf("expected latitude 40.7128, got %f", lat)
	}
	if lng != -74.0060 {
		t.Errorf("expected longitude -74.0060, got %f", lng)
	}
}

func TestShop_IsActive(t *testing.T) {
	tests := []struct {
		name string
		shop Shop
		want bool
	}{
		{
			name: "active shop",
			shop: Shop{PreLaunchEnabled: false, PasswordEnabled: false},
			want: true,
		},
		{
			name: "pre-launch enabled",
			shop: Shop{PreLaunchEnabled: true, PasswordEnabled: false},
			want: false,
		},
		{
			name: "password enabled",
			shop: Shop{PreLaunchEnabled: false, PasswordEnabled: true},
			want: false,
		},
		{
			name: "both enabled",
			shop: Shop{PreLaunchEnabled: true, PasswordEnabled: true},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.shop.IsActive()
			if got != tt.want {
				t.Errorf("IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShop_CanAcceptPayments(t *testing.T) {
	tests := []struct {
		name string
		shop Shop
		want bool
	}{
		{
			name: "can accept payments",
			shop: Shop{EligibleForPayments: true, RequiresExtraPaymentsAgreement: false},
			want: true,
		},
		{
			name: "not eligible",
			shop: Shop{EligibleForPayments: false, RequiresExtraPaymentsAgreement: false},
			want: false,
		},
		{
			name: "requires extra agreement",
			shop: Shop{EligibleForPayments: true, RequiresExtraPaymentsAgreement: true},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.shop.CanAcceptPayments()
			if got != tt.want {
				t.Errorf("CanAcceptPayments() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShop_FormatMoney(t *testing.T) {
	tests := []struct {
		name   string
		shop   Shop
		amount float64
		want   string
	}{
		{
			name:   "with money format",
			shop:   Shop{MoneyFormat: "${{amount}}"},
			amount: 99.99,
			want:   "$99.99",
		},
		{
			name:   "with currency fallback",
			shop:   Shop{Currency: "USD"},
			amount: 99.99,
			want:   "USD 99.99",
		},
		{
			name:   "no format or currency",
			shop:   Shop{},
			amount: 99.99,
			want:   "99.99",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.shop.FormatMoney(tt.amount)
			if got != tt.want {
				t.Errorf("FormatMoney() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShop_FormatMoneyWithCurrency(t *testing.T) {
	tests := []struct {
		name   string
		shop   Shop
		amount float64
		want   string
	}{
		{
			name:   "with full format",
			shop:   Shop{MoneyWithCurrencyFormat: "{{amount}} {{currency}}", Currency: "USD"},
			amount: 99.99,
			want:   "99.99 USD",
		},
		{
			name:   "fallback to FormatMoney",
			shop:   Shop{MoneyFormat: "${{amount}}"},
			amount: 99.99,
			want:   "$99.99",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.shop.FormatMoneyWithCurrency(tt.amount)
			if got != tt.want {
				t.Errorf("FormatMoneyWithCurrency() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShop_Touch(t *testing.T) {
	shop := NewShop(1, "Test", "test@example.com")
	originalUpdatedAt := shop.UpdatedAt

	// Wait a tiny bit to ensure time difference
	time.Sleep(time.Millisecond)
	shop.Touch()

	if shop.UpdatedAt == nil {
		t.Error("UpdatedAt should not be nil after Touch()")
	}
	if !shop.UpdatedAt.After(*originalUpdatedAt) {
		t.Error("UpdatedAt should be after the original time")
	}
}

func TestShop_SetAddress(t *testing.T) {
	shop := NewShop(1, "Test", "test@example.com")
	shop.SetAddress("123 Main St", "Suite 100", "New York", "New York", "NY", "10001", "US", "USA", "United States")

	if shop.Address1 != "123 Main St" {
		t.Errorf("expected Address1 '123 Main St', got %s", shop.Address1)
	}
	if shop.Address2 != "Suite 100" {
		t.Errorf("expected Address2 'Suite 100', got %s", shop.Address2)
	}
	if shop.City != "New York" {
		t.Errorf("expected City 'New York', got %s", shop.City)
	}
	if shop.Province != "New York" {
		t.Errorf("expected Province 'New York', got %s", shop.Province)
	}
	if shop.ProvinceCode != "NY" {
		t.Errorf("expected ProvinceCode 'NY', got %s", shop.ProvinceCode)
	}
	if shop.Zip != "10001" {
		t.Errorf("expected Zip '10001', got %s", shop.Zip)
	}
	if shop.Country != "US" {
		t.Errorf("expected Country 'US', got %s", shop.Country)
	}
	if shop.CountryCode != "USA" {
		t.Errorf("expected CountryCode 'USA', got %s", shop.CountryCode)
	}
	if shop.CountryName != "United States" {
		t.Errorf("expected CountryName 'United States', got %s", shop.CountryName)
	}
}

func TestShop_SetCoordinates(t *testing.T) {
	shop := NewShop(1, "Test", "test@example.com")
	shop.SetCoordinates(40.7128, -74.0060)

	if shop.Latitude != 40.7128 {
		t.Errorf("expected Latitude 40.7128, got %f", shop.Latitude)
	}
	if shop.Longitude != -74.0060 {
		t.Errorf("expected Longitude -74.0060, got %f", shop.Longitude)
	}
}

func TestShop_HasCustomDomain(t *testing.T) {
	tests := []struct {
		name string
		shop Shop
		want bool
	}{
		{
			name: "has custom domain",
			shop: Shop{Domain: "mystore.com", MyShopifyDomain: "mystore.myshopify.com"},
			want: true,
		},
		{
			name: "domain same as myshopify",
			shop: Shop{Domain: "mystore.myshopify.com", MyShopifyDomain: "mystore.myshopify.com"},
			want: false,
		},
		{
			name: "no custom domain",
			shop: Shop{MyShopifyDomain: "mystore.myshopify.com"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.shop.HasCustomDomain()
			if got != tt.want {
				t.Errorf("HasCustomDomain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShop_SupportsFeature(t *testing.T) {
	shop := Shop{
		HasStorefront:        true,
		HasDiscounts:         true,
		HasGiftcards:         false,
		CheckoutAPISupported: true,
	}

	tests := []struct {
		feature string
		want    bool
	}{
		{"storefront", true},
		{"Storefront", true},
		{"discounts", true},
		{"giftcards", false},
		{"gift_cards", false},
		{"checkout_api", true},
		{"unknown_feature", false},
	}

	for _, tt := range tests {
		t.Run(tt.feature, func(t *testing.T) {
			got := shop.SupportsFeature(tt.feature)
			if got != tt.want {
				t.Errorf("SupportsFeature(%s) = %v, want %v", tt.feature, got, tt.want)
			}
		})
	}
}
