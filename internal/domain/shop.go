package domain

import (
	"fmt"
	"strings"
	"time"
)

// Shop represents a store/shop entity with all its configuration and metadata.
type Shop struct {
	Id                              uint64     `json:"id"`
	Name                            string     `json:"name"`
	ShopOwner                       string     `json:"shop_owner"`
	Email                           string     `json:"email"`
	CustomerEmail                   string     `json:"customer_email"`
	CreatedAt                       *time.Time `json:"created_at"`
	UpdatedAt                       *time.Time `json:"updated_at"`
	Address1                        string     `json:"address1"`
	Address2                        string     `json:"address2"`
	City                            string     `json:"city"`
	Country                         string     `json:"country"`
	CountryCode                     string     `json:"country_code"`
	CountryName                     string     `json:"country_name"`
	Currency                        string     `json:"currency"`
	Domain                          string     `json:"domain"`
	Latitude                        float64    `json:"latitude"`
	Longitude                       float64    `json:"longitude"`
	Phone                           string     `json:"phone"`
	Province                        string     `json:"province"`
	ProvinceCode                    string     `json:"province_code"`
	Zip                             string     `json:"zip"`
	MoneyFormat                     string     `json:"money_format"`
	MoneyWithCurrencyFormat         string     `json:"money_with_currency_format"`
	WeightUnit                      string     `json:"weight_unit"`
	MyShopifyDomain                 string     `json:"myshopify_domain"`
	PlanName                        string     `json:"plan_name"`
	PlanDisplayName                 string     `json:"plan_display_name"`
	PasswordEnabled                 bool       `json:"password_enabled"`
	PrimaryLocale                   string     `json:"primary_locale"`
	PrimaryLocationId               uint64     `json:"primary_location_id"`
	Timezone                        string     `json:"timezone"`
	IanaTimezone                    string     `json:"iana_timezone"`
	ForceSSL                        bool       `json:"force_ssl"`
	TaxShipping                     bool       `json:"tax_shipping"`
	TaxesIncluded                   bool       `json:"taxes_included"`
	HasStorefront                   bool       `json:"has_storefront"`
	HasDiscounts                    bool       `json:"has_discounts"`
	HasGiftcards                    bool       `json:"has_gift_cards"`
	CountyTaxes                     bool       `json:"county_taxes"`
	CheckoutAPISupported            bool       `json:"checkout_api_supported"`
	Source                          string     `json:"source"`
	MoneyInEmailsFormat             string     `json:"money_in_emails_format"`
	MoneyWithCurrencyInEmailsFormat string     `json:"money_with_currency_in_emails_format"`
	EligibleForPayments             bool       `json:"eligible_for_payments"`
	RequiresExtraPaymentsAgreement  bool       `json:"requires_extra_payments_agreement"`
	PreLaunchEnabled                bool       `json:"pre_launch_enabled"`
}

// NewShop creates a new Shop with the required fields and sets timestamps.
func NewShop(id uint64, name, email string) *Shop {
	now := time.Now()
	return &Shop{
		Id:        id,
		Name:      name,
		Email:     email,
		CreatedAt: &now,
		UpdatedAt: &now,
		ForceSSL:  true, // Default to secure
	}
}

// Validate checks if the shop has all required fields populated.
func (s *Shop) Validate() error {
	var errs []string

	if s.Id == 0 {
		errs = append(errs, "id is required")
	}
	if strings.TrimSpace(s.Name) == "" {
		errs = append(errs, "name is required")
	}
	if strings.TrimSpace(s.Email) == "" {
		errs = append(errs, "email is required")
	}
	if strings.TrimSpace(s.Domain) == "" && strings.TrimSpace(s.MyShopifyDomain) == "" {
		errs = append(errs, "domain or myshopify_domain is required")
	}

	if len(errs) > 0 {
		return fmt.Errorf("shop validation failed: %s", strings.Join(errs, "; "))
	}
	return nil
}

// FullAddress returns the complete formatted address of the shop.
func (s *Shop) FullAddress() string {
	var parts []string

	if s.Address1 != "" {
		parts = append(parts, s.Address1)
	}
	if s.Address2 != "" {
		parts = append(parts, s.Address2)
	}
	if s.City != "" {
		parts = append(parts, s.City)
	}
	if s.Province != "" {
		parts = append(parts, s.Province)
	}
	if s.Zip != "" {
		parts = append(parts, s.Zip)
	}
	if s.CountryName != "" {
		parts = append(parts, s.CountryName)
	} else if s.Country != "" {
		parts = append(parts, s.Country)
	}

	return strings.Join(parts, ", ")
}

// PrimaryDomain returns the primary domain for the shop.
// It prefers the custom domain over the myshopify domain.
func (s *Shop) PrimaryDomain() string {
	if s.Domain != "" {
		return s.Domain
	}
	return s.MyShopifyDomain
}

// HasCoordinates checks if the shop has valid latitude and longitude set.
func (s *Shop) HasCoordinates() bool {
	return s.Latitude != 0 || s.Longitude != 0
}

// Coordinates returns the latitude and longitude as a tuple.
func (s *Shop) Coordinates() (float64, float64) {
	return s.Latitude, s.Longitude
}

// IsActive checks if the shop is active (not in pre-launch mode and not password protected).
func (s *Shop) IsActive() bool {
	return !s.PreLaunchEnabled && !s.PasswordEnabled
}

// CanAcceptPayments checks if the shop can process payments.
func (s *Shop) CanAcceptPayments() bool {
	return s.EligibleForPayments && !s.RequiresExtraPaymentsAgreement
}

// FormatMoney formats an amount using the shop's money format.
// If no format is set, it returns a basic format with the currency.
func (s *Shop) FormatMoney(amount float64) string {
	if s.MoneyFormat != "" {
		// Replace {{amount}} placeholder with the actual amount
		return strings.ReplaceAll(s.MoneyFormat, "{{amount}}", fmt.Sprintf("%.2f", amount))
	}
	if s.Currency != "" {
		return fmt.Sprintf("%s %.2f", s.Currency, amount)
	}
	return fmt.Sprintf("%.2f", amount)
}

// FormatMoneyWithCurrency formats an amount using the shop's money with currency format.
func (s *Shop) FormatMoneyWithCurrency(amount float64) string {
	if s.MoneyWithCurrencyFormat != "" {
		formatted := strings.ReplaceAll(s.MoneyWithCurrencyFormat, "{{amount}}", fmt.Sprintf("%.2f", amount))
		return strings.ReplaceAll(formatted, "{{currency}}", s.Currency)
	}
	return s.FormatMoney(amount)
}

// Touch updates the UpdatedAt timestamp to the current time.
func (s *Shop) Touch() {
	now := time.Now()
	s.UpdatedAt = &now
}

// SetAddress sets all address-related fields at once.
func (s *Shop) SetAddress(address1, address2, city, province, provinceCode, zip, country, countryCode, countryName string) {
	s.Address1 = address1
	s.Address2 = address2
	s.City = city
	s.Province = province
	s.ProvinceCode = provinceCode
	s.Zip = zip
	s.Country = country
	s.CountryCode = countryCode
	s.CountryName = countryName
	s.Touch()
}

// SetCoordinates sets the latitude and longitude for the shop.
func (s *Shop) SetCoordinates(lat, lng float64) {
	s.Latitude = lat
	s.Longitude = lng
	s.Touch()
}

// HasCustomDomain checks if the shop has a custom domain configured.
func (s *Shop) HasCustomDomain() bool {
	return s.Domain != "" && s.Domain != s.MyShopifyDomain
}

// SupportsFeature checks if the shop supports a particular feature based on its configuration.
func (s *Shop) SupportsFeature(feature string) bool {
	switch strings.ToLower(feature) {
	case "storefront":
		return s.HasStorefront
	case "discounts":
		return s.HasDiscounts
	case "giftcards", "gift_cards":
		return s.HasGiftcards
	case "checkout_api":
		return s.CheckoutAPISupported
	default:
		return false
	}
}
