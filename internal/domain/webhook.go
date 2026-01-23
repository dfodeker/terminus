package domain

import (
	"time"

	"github.com/google/uuid"
)

type WebhookID uuid.UUID

func (id WebhookID) String() string {
	return uuid.UUID(id).String()
}

type WebhookStatus string

const (
	WebhookStatusActive   WebhookStatus = "active"
	WebhookStatusPaused   WebhookStatus = "paused"
	WebhookStatusFailed   WebhookStatus = "failed"
	WebhookStatusDisabled WebhookStatus = "disabled"
)

// Webhook represents a webhook subscription for a shop or organization
type Webhook struct {
	ID                uuid.UUID
	OrganizationID    uuid.UUID
	ShopID            *uuid.UUID // nil = applies to all shops in org
	Topic             string     // e.g., "orders/create", "products/update"
	EndpointURL       string     // The URL to deliver webhooks to
	SecretHash        string     // Hashed HMAC secret for signing payloads
	Fields            []string   // Optional field filtering
	Status            WebhookStatus
	APIVersion        string // e.g., "2025-01"
	FailureCount      int
	LastFailureAt     *time.Time
	LastFailureReason string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// WebhookDeliveryStatus represents the status of a delivery attempt
type WebhookDeliveryStatus string

const (
	WebhookDeliveryPending  WebhookDeliveryStatus = "pending"
	WebhookDeliverySuccess  WebhookDeliveryStatus = "success"
	WebhookDeliveryFailed   WebhookDeliveryStatus = "failed"
	WebhookDeliveryRetrying WebhookDeliveryStatus = "retrying"
)

// WebhookDelivery represents a single delivery attempt for a webhook
type WebhookDelivery struct {
	ID              uuid.UUID
	WebhookID       uuid.UUID
	OrganizationID  uuid.UUID
	ShopID          *uuid.UUID
	Topic           string
	EndpointURL     string
	RequestHeaders  map[string]string
	RequestBody     []byte
	ResponseStatus  int
	ResponseHeaders map[string]string
	ResponseBody    []byte
	Status          WebhookDeliveryStatus
	Attempts        int
	NextRetryAt     *time.Time
	ErrorMessage    string
	DurationMs      int
	CreatedAt       time.Time
	DeliveredAt     *time.Time
}

// WebhookTopics defines all available webhook topics
var WebhookTopics = []string{
	// Orders
	"orders/create",
	"orders/updated",
	"orders/cancelled",
	"orders/fulfilled",
	"orders/paid",
	"orders/partially_fulfilled",

	// Products
	"products/create",
	"products/update",
	"products/delete",

	// Customers
	"customers/create",
	"customers/update",
	"customers/delete",

	// Inventory
	"inventory_levels/update",
	"inventory_levels/connect",
	"inventory_levels/disconnect",

	// Fulfillments
	"fulfillments/create",
	"fulfillments/update",

	// Refunds
	"refunds/create",

	// Shop
	"shop/update",

	// App
	"app/uninstalled",
}

// IsValidTopic checks if a topic is valid
func IsValidTopic(topic string) bool {
	for _, t := range WebhookTopics {
		if t == topic {
			return true
		}
	}
	return false
}

// CanRetry checks if a delivery can be retried
func (d *WebhookDelivery) CanRetry() bool {
	return d.Status == WebhookDeliveryFailed && d.Attempts < 5
}

// ShouldDisableWebhook checks if the webhook should be disabled after too many failures
func (w *Webhook) ShouldDisableWebhook() bool {
	return w.FailureCount >= 10
}
