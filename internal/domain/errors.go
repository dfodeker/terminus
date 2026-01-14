// Package domain contains core business entities and errors.
package domain

import (
	"errors"
	"fmt"
)

// =============================================================================
// Base Error Types
// =============================================================================

// DomainError represents a business logic error with a code and message.
type DomainError struct {
	Code    string
	Message string
	Err     error
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// Is implements errors.Is for DomainError.
func (e *DomainError) Is(target error) bool {
	t, ok := target.(*DomainError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// NewDomainError creates a new DomainError.
func NewDomainError(code, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an underlying error with the domain error.
func (e *DomainError) Wrap(err error) *DomainError {
	return &DomainError{
		Code:    e.Code,
		Message: e.Message,
		Err:     err,
	}
}

// WithMessage returns a copy of the error with a custom message.
func (e *DomainError) WithMessage(message string) *DomainError {
	return &DomainError{
		Code:    e.Code,
		Message: message,
		Err:     e.Err,
	}
}

// =============================================================================
// Not Found Errors
// =============================================================================

var (
	// ErrNotFound is the base not found error.
	ErrNotFound = NewDomainError("NOT_FOUND", "resource not found")

	// ErrShopNotFound is returned when a shop cannot be found.
	ErrShopNotFound = NewDomainError("SHOP_NOT_FOUND", "shop not found")

	// ErrProductNotFound is returned when a product cannot be found.
	ErrProductNotFound = NewDomainError("PRODUCT_NOT_FOUND", "product not found")

	// ErrProductVariantNotFound is returned when a product variant cannot be found.
	ErrProductVariantNotFound = NewDomainError("PRODUCT_VARIANT_NOT_FOUND", "product variant not found")

	// ErrOrderNotFound is returned when an order cannot be found.
	ErrOrderNotFound = NewDomainError("ORDER_NOT_FOUND", "order not found")

	// ErrCustomerNotFound is returned when a customer cannot be found.
	ErrCustomerNotFound = NewDomainError("CUSTOMER_NOT_FOUND", "customer not found")

	// ErrCollectionNotFound is returned when a collection cannot be found.
	ErrCollectionNotFound = NewDomainError("COLLECTION_NOT_FOUND", "collection not found")

	// ErrThemeNotFound is returned when a theme cannot be found.
	ErrThemeNotFound = NewDomainError("THEME_NOT_FOUND", "theme not found")

	// ErrPageNotFound is returned when a page cannot be found.
	ErrPageNotFound = NewDomainError("PAGE_NOT_FOUND", "page not found")

	// ErrWebhookNotFound is returned when a webhook cannot be found.
	ErrWebhookNotFound = NewDomainError("WEBHOOK_NOT_FOUND", "webhook not found")

	// ErrMenuNotFound is returned when a menu cannot be found.
	ErrMenuNotFound = NewDomainError("MENU_NOT_FOUND", "menu not found")

	// ErrAddressNotFound is returned when an address cannot be found.
	ErrAddressNotFound = NewDomainError("ADDRESS_NOT_FOUND", "address not found")

	// ErrFulfillmentNotFound is returned when a fulfillment cannot be found.
	ErrFulfillmentNotFound = NewDomainError("FULFILLMENT_NOT_FOUND", "fulfillment not found")

	// ErrTransactionNotFound is returned when a transaction cannot be found.
	ErrTransactionNotFound = NewDomainError("TRANSACTION_NOT_FOUND", "transaction not found")
)

// =============================================================================
// Product Errors
// =============================================================================

var (
	// ErrProductNotAvailable is returned when a product is not available for purchase.
	ErrProductNotAvailable = NewDomainError("PRODUCT_NOT_AVAILABLE", "product is not available for purchase")

	// ErrInsufficientInventory is returned when there is not enough inventory.
	ErrInsufficientInventory = NewDomainError("INSUFFICIENT_INVENTORY", "insufficient inventory to fulfill request")

	// ErrProductAlreadyExists is returned when trying to create a product that already exists.
	ErrProductAlreadyExists = NewDomainError("PRODUCT_ALREADY_EXISTS", "product already exists")

	// ErrDuplicateSKU is returned when a SKU already exists.
	ErrDuplicateSKU = NewDomainError("DUPLICATE_SKU", "SKU already exists")

	// ErrDuplicateSlug is returned when a product slug already exists.
	ErrDuplicateSlug = NewDomainError("DUPLICATE_SLUG", "slug already exists")

	// ErrInvalidProductStatus is returned when an invalid product status is provided.
	ErrInvalidProductStatus = NewDomainError("INVALID_PRODUCT_STATUS", "invalid product status")

	// ErrProductArchived is returned when trying to modify an archived product.
	ErrProductArchived = NewDomainError("PRODUCT_ARCHIVED", "cannot modify archived product")

	// ErrVariantRequired is returned when at least one variant is required.
	ErrVariantRequired = NewDomainError("VARIANT_REQUIRED", "at least one variant is required")

	// ErrMaxVariantsExceeded is returned when the maximum number of variants is exceeded.
	ErrMaxVariantsExceeded = NewDomainError("MAX_VARIANTS_EXCEEDED", "maximum number of variants exceeded")

	// ErrInvalidPrice is returned when an invalid price is provided.
	ErrInvalidPrice = NewDomainError("INVALID_PRICE", "price must be a non-negative value")

	// ErrInvalidInventoryQuantity is returned when an invalid inventory quantity is provided.
	ErrInvalidInventoryQuantity = NewDomainError("INVALID_INVENTORY_QUANTITY", "inventory quantity cannot be negative")
)

// =============================================================================
// Order Errors
// =============================================================================

var (
	// ErrOrderAlreadyPaid is returned when trying to pay for an already paid order.
	ErrOrderAlreadyPaid = NewDomainError("ORDER_ALREADY_PAID", "order has already been paid")

	// ErrOrderAlreadyCancelled is returned when trying to modify a cancelled order.
	ErrOrderAlreadyCancelled = NewDomainError("ORDER_ALREADY_CANCELLED", "order has already been cancelled")

	// ErrOrderAlreadyFulfilled is returned when trying to cancel a fulfilled order.
	ErrOrderAlreadyFulfilled = NewDomainError("ORDER_ALREADY_FULFILLED", "order has already been fulfilled")

	// ErrOrderCannotBeCancelled is returned when an order cannot be cancelled.
	ErrOrderCannotBeCancelled = NewDomainError("ORDER_CANNOT_BE_CANCELLED", "order cannot be cancelled in current state")

	// ErrOrderCannotBeEdited is returned when an order cannot be edited.
	ErrOrderCannotBeEdited = NewDomainError("ORDER_CANNOT_BE_EDITED", "order cannot be edited in current state")

	// ErrOrderCannotBeFulfilled is returned when an order cannot be fulfilled.
	ErrOrderCannotBeFulfilled = NewDomainError("ORDER_CANNOT_BE_FULFILLED", "order cannot be fulfilled in current state")

	// ErrInvalidOrderStatus is returned when an invalid order status is provided.
	ErrInvalidOrderStatus = NewDomainError("INVALID_ORDER_STATUS", "invalid order status")

	// ErrEmptyOrder is returned when trying to create an order with no line items.
	ErrEmptyOrder = NewDomainError("EMPTY_ORDER", "order must have at least one line item")

	// ErrInvalidLineItemQuantity is returned when an invalid line item quantity is provided.
	ErrInvalidLineItemQuantity = NewDomainError("INVALID_LINE_ITEM_QUANTITY", "line item quantity must be positive")

	// ErrRefundExceedsPayment is returned when a refund amount exceeds the payment.
	ErrRefundExceedsPayment = NewDomainError("REFUND_EXCEEDS_PAYMENT", "refund amount exceeds original payment")

	// ErrFulfillmentQuantityExceeded is returned when fulfillment quantity exceeds fulfillable quantity.
	ErrFulfillmentQuantityExceeded = NewDomainError("FULFILLMENT_QUANTITY_EXCEEDED", "fulfillment quantity exceeds fulfillable quantity")
)

// =============================================================================
// Customer Errors
// =============================================================================

var (
	// ErrCustomerAlreadyExists is returned when a customer with the same email already exists.
	ErrCustomerAlreadyExists = NewDomainError("CUSTOMER_ALREADY_EXISTS", "customer with this email already exists")

	// ErrInvalidEmail is returned when an invalid email address is provided.
	ErrInvalidEmail = NewDomainError("INVALID_EMAIL", "invalid email address")

	// ErrInvalidPhone is returned when an invalid phone number is provided.
	ErrInvalidPhone = NewDomainError("INVALID_PHONE", "invalid phone number")

	// ErrCustomerDisabled is returned when trying to perform actions on a disabled customer.
	ErrCustomerDisabled = NewDomainError("CUSTOMER_DISABLED", "customer account is disabled")

	// ErrInvalidPassword is returned when an invalid password is provided.
	ErrInvalidPassword = NewDomainError("INVALID_PASSWORD", "invalid password")

	// ErrPasswordTooWeak is returned when a password doesn't meet requirements.
	ErrPasswordTooWeak = NewDomainError("PASSWORD_TOO_WEAK", "password does not meet minimum requirements")
)

// =============================================================================
// Authentication & Authorization Errors
// =============================================================================

var (
	// ErrUnauthorized is returned when authentication is required but not provided.
	ErrUnauthorized = NewDomainError("UNAUTHORIZED", "authentication required")

	// ErrForbidden is returned when the user doesn't have permission.
	ErrForbidden = NewDomainError("FORBIDDEN", "access denied")

	// ErrInvalidAPIKey is returned when an invalid API key is provided.
	ErrInvalidAPIKey = NewDomainError("INVALID_API_KEY", "invalid API key")

	// ErrCredentialRevoked is returned when the credential has been revoked.
	ErrCredentialRevoked = NewDomainError("CREDENTIAL_REVOKED", "credential has been revoked")

	// ErrCredentialExpired is returned when the credential has expired.
	ErrCredentialExpired = NewDomainError("CREDENTIAL_EXPIRED", "credential has expired")

	// ErrInsufficientScope is returned when the required scope is missing.
	ErrInsufficientScope = NewDomainError("INSUFFICIENT_SCOPE", "insufficient scope for this operation")

	// ErrInvalidToken is returned when an invalid token is provided.
	ErrInvalidToken = NewDomainError("INVALID_TOKEN", "invalid token")

	// ErrTokenExpired is returned when a token has expired.
	ErrTokenExpired = NewDomainError("TOKEN_EXPIRED", "token has expired")

	// ErrInvalidOAuthCode is returned when an invalid OAuth code is provided.
	ErrInvalidOAuthCode = NewDomainError("INVALID_OAUTH_CODE", "invalid authorization code")

	// ErrInvalidRedirectURI is returned when an invalid redirect URI is provided.
	ErrInvalidRedirectURI = NewDomainError("INVALID_REDIRECT_URI", "invalid redirect URI")
)

// =============================================================================
// Shop Errors
// =============================================================================

var (
	// ErrShopSuspended is returned when a shop is suspended.
	ErrShopSuspended = NewDomainError("SHOP_SUSPENDED", "shop is suspended")

	// ErrShopCreateDisable is returned when an organization cannot create more shops.
	ErrShopCreateDisable = NewDomainError("SHOP_CREATE_DISABLED", "organization has reached maximum allowed shops")

	// ErrShopInactive is returned when a shop is not active.
	ErrShopInactive = NewDomainError("SHOP_INACTIVE", "shop is not active")

	// ErrSubdomainTaken is returned when a subdomain is already taken.
	ErrSubdomainTaken = NewDomainError("SUBDOMAIN_TAKEN", "subdomain is already taken")

	// ErrInvalidSubdomain is returned when an invalid subdomain is provided.
	ErrInvalidSubdomain = NewDomainError("INVALID_SUBDOMAIN", "invalid subdomain format")

	// ErrCustomDomainTaken is returned when a custom domain is already taken.
	ErrCustomDomainTaken = NewDomainError("CUSTOM_DOMAIN_TAKEN", "custom domain is already in use")

	// ErrInvalidCustomDomain is returned when an invalid custom domain is provided.
	ErrInvalidCustomDomain = NewDomainError("INVALID_CUSTOM_DOMAIN", "invalid custom domain")
)

// =============================================================================
// Theme Errors
// =============================================================================

var (
	// ErrThemeInvalid is returned when a theme is invalid.
	ErrThemeInvalid = NewDomainError("THEME_INVALID", "theme is invalid")

	// ErrThemeProcessing is returned when a theme is still processing.
	ErrThemeProcessing = NewDomainError("THEME_PROCESSING", "theme is still processing")

	// ErrThemeProcessingFailed is returned when theme processing failed.
	ErrThemeProcessingFailed = NewDomainError("THEME_PROCESSING_FAILED", "theme processing failed")

	// ErrMainThemeCannotBeDeleted is returned when trying to delete the main theme.
	ErrMainThemeCannotBeDeleted = NewDomainError("MAIN_THEME_CANNOT_BE_DELETED", "cannot delete the main theme")

	// ErrInvalidThemeFile is returned when a theme file is invalid.
	ErrInvalidThemeFile = NewDomainError("INVALID_THEME_FILE", "invalid theme file")

	// ErrThemeAssetNotFound is returned when a theme asset cannot be found.
	ErrThemeAssetNotFound = NewDomainError("THEME_ASSET_NOT_FOUND", "theme asset not found")

	// ErrInvalidSectionSchema is returned when a section schema is invalid.
	ErrInvalidSectionSchema = NewDomainError("INVALID_SECTION_SCHEMA", "invalid section schema")

	// ErrSectionNotFound is returned when a section cannot be found.
	ErrSectionNotFound = NewDomainError("SECTION_NOT_FOUND", "section not found")
)

// =============================================================================
// Webhook Errors
// =============================================================================

var (
	// ErrWebhookDeliveryFailed is returned when webhook delivery fails.
	ErrWebhookDeliveryFailed = NewDomainError("WEBHOOK_DELIVERY_FAILED", "webhook delivery failed")

	// ErrInvalidWebhookTopic is returned when an invalid webhook topic is provided.
	ErrInvalidWebhookTopic = NewDomainError("INVALID_WEBHOOK_TOPIC", "invalid webhook topic")

	// ErrInvalidWebhookURL is returned when an invalid webhook URL is provided.
	ErrInvalidWebhookURL = NewDomainError("INVALID_WEBHOOK_URL", "invalid webhook URL")

	// ErrWebhookDisabled is returned when a webhook is disabled.
	ErrWebhookDisabled = NewDomainError("WEBHOOK_DISABLED", "webhook is disabled due to repeated failures")
)

// =============================================================================
// GID Errors
// =============================================================================

var (
	// ErrInvalidGID is returned when a GID is invalid.
	ErrInvalidGID = NewDomainError("INVALID_GID", "invalid global ID format")

	// ErrInvalidGIDType is returned when a GID has an unexpected type.
	ErrInvalidGIDType = NewDomainError("INVALID_GID_TYPE", "invalid global ID type")

	// ErrGIDTypeMismatch is returned when a GID type doesn't match expected type.
	ErrGIDTypeMismatch = NewDomainError("GID_TYPE_MISMATCH", "global ID type mismatch")
)

// =============================================================================
// Payment Errors
// =============================================================================

var (
	// ErrPaymentFailed is returned when a payment fails.
	ErrPaymentFailed = NewDomainError("PAYMENT_FAILED", "payment failed")

	// ErrPaymentDeclined is returned when a payment is declined.
	ErrPaymentDeclined = NewDomainError("PAYMENT_DECLINED", "payment was declined")

	// ErrInvalidPaymentMethod is returned when an invalid payment method is provided.
	ErrInvalidPaymentMethod = NewDomainError("INVALID_PAYMENT_METHOD", "invalid payment method")

	// ErrPaymentAlreadyProcessed is returned when a payment has already been processed.
	ErrPaymentAlreadyProcessed = NewDomainError("PAYMENT_ALREADY_PROCESSED", "payment has already been processed")

	// ErrRefundFailed is returned when a refund fails.
	ErrRefundFailed = NewDomainError("REFUND_FAILED", "refund failed")
)

// =============================================================================
// Validation Errors
// =============================================================================

var (
	// ErrValidation is the base validation error.
	ErrValidation = NewDomainError("VALIDATION_ERROR", "validation failed")

	// ErrRequiredField is returned when a required field is missing.
	ErrRequiredField = NewDomainError("REQUIRED_FIELD", "required field is missing")

	// ErrInvalidInput is returned when input is invalid.
	ErrInvalidInput = NewDomainError("INVALID_INPUT", "invalid input provided")

	// ErrInvalidCursor is returned when an invalid pagination cursor is provided.
	ErrInvalidCursor = NewDomainError("INVALID_CURSOR", "invalid pagination cursor")

	// ErrLimitExceeded is returned when a limit is exceeded.
	ErrLimitExceeded = NewDomainError("LIMIT_EXCEEDED", "limit exceeded")
)

// =============================================================================
// Rate Limiting Errors
// =============================================================================

var (
	// ErrRateLimitExceeded is returned when rate limit is exceeded.
	ErrRateLimitExceeded = NewDomainError("RATE_LIMIT_EXCEEDED", "rate limit exceeded")
)

// =============================================================================
// Concurrency Errors
// =============================================================================

var (
	// ErrConcurrentModification is returned when a concurrent modification is detected.
	ErrConcurrentModification = NewDomainError("CONCURRENT_MODIFICATION", "resource was modified by another request")

	// ErrOptimisticLock is returned when an optimistic lock fails.
	ErrOptimisticLock = NewDomainError("OPTIMISTIC_LOCK_FAILED", "optimistic lock failed")
)

// =============================================================================
// Helper Functions
// =============================================================================

// IsNotFound returns true if the error is a not found error.
func IsNotFound(err error) bool {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		switch domainErr.Code {
		case ErrNotFound.Code,
			ErrShopNotFound.Code,
			ErrProductNotFound.Code,
			ErrProductVariantNotFound.Code,
			ErrOrderNotFound.Code,
			ErrCustomerNotFound.Code,
			ErrCollectionNotFound.Code,
			ErrThemeNotFound.Code,
			ErrPageNotFound.Code,
			ErrWebhookNotFound.Code,
			ErrMenuNotFound.Code,
			ErrAddressNotFound.Code,
			ErrFulfillmentNotFound.Code,
			ErrTransactionNotFound.Code,
			ErrThemeAssetNotFound.Code,
			ErrSectionNotFound.Code:
			return true
		}
	}
	return false
}

// IsUnauthorized returns true if the error is an authorization error.
func IsUnauthorized(err error) bool {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		switch domainErr.Code {
		case ErrUnauthorized.Code,
			ErrInvalidAPIKey.Code,
			ErrCredentialRevoked.Code,
			ErrCredentialExpired.Code,
			ErrInvalidToken.Code,
			ErrTokenExpired.Code:
			return true
		}
	}
	return false
}

// IsForbidden returns true if the error is a forbidden error.
func IsForbidden(err error) bool {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		switch domainErr.Code {
		case ErrForbidden.Code,
			ErrInsufficientScope.Code:
			return true
		}
	}
	return false
}

// IsValidation returns true if the error is a validation error.
func IsValidation(err error) bool {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		switch domainErr.Code {
		case ErrValidation.Code,
			ErrRequiredField.Code,
			ErrInvalidInput.Code,
			ErrInvalidEmail.Code,
			ErrInvalidPhone.Code,
			ErrInvalidPrice.Code,
			ErrInvalidProductStatus.Code,
			ErrInvalidOrderStatus.Code,
			ErrInvalidCursor.Code,
			ErrInvalidGID.Code,
			ErrInvalidGIDType.Code,
			ErrGIDTypeMismatch.Code:
			return true
		}
	}
	return false
}

// IsConflict returns true if the error is a conflict error.
func IsConflict(err error) bool {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		switch domainErr.Code {
		case ErrProductAlreadyExists.Code,
			ErrCustomerAlreadyExists.Code,
			ErrDuplicateSKU.Code,
			ErrDuplicateSlug.Code,
			ErrSubdomainTaken.Code,
			ErrCustomDomainTaken.Code,
			ErrConcurrentModification.Code,
			ErrOptimisticLock.Code:
			return true
		}
	}
	return false
}

// IsRateLimited returns true if the error is a rate limit error.
func IsRateLimited(err error) bool {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Code == ErrRateLimitExceeded.Code
	}
	return false
}

// ErrorCode returns the error code if it's a DomainError, otherwise returns empty string.
func ErrorCode(err error) string {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Code
	}
	return ""
}

// ErrorMessage returns the error message if it's a DomainError, otherwise returns err.Error().
func ErrorMessage(err error) string {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Message
	}
	return err.Error()
}
