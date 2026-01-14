package ports

import (
	"context"
	"time"

	"github.com/dfodeker/storeos/internal/domain"
	"github.com/google/uuid"
)

type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	GetByID(ctx context.Context, shopID domain.ShopID, id domain.ProductID) (*domain.Product, error)
	List(ctx context.Context, shopID domain.ShopID, filter ProductFilter) ([]domain.Product, string, error)
	Update(ctx context.Context, product *domain.Product) error
	Delete(ctx context.Context, shopID domain.ShopID, id domain.ProductID) error
}

type ProductFilter struct {
	Status *domain.ProductStatus
	Cursor string
	Limit  int
}

type ShopRepository interface {
	Create(ctx context.Context, shop *domain.Shop) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Shop, error)
	GetBySubdomain(ctx context.Context, subdomain string) (*domain.Shop, error)
	GetByCustomDomain(ctx context.Context, domain string) (*domain.Shop, error)
	Update(ctx context.Context, shop *domain.Shop) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter ShopFilter) ([]domain.Shop, string, error)
}

type ShopFilter struct {
	Status         *string
	OrganizationID *uuid.UUID
	Cursor         string
	Limit          int
}

type OrganizationRepository interface {
	Create(ctx context.Context, org *domain.Organization) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error)
	GetByHandle(ctx context.Context, handle string) (*domain.Organization, error)
	Update(ctx context.Context, org *domain.Organization) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter OrganizationFilter) ([]domain.Organization, string, error)
	CountShops(ctx context.Context, orgID uuid.UUID) (int, error)
	CountMembers(ctx context.Context, orgID uuid.UUID) (int, error)
}

type OrganizationFilter struct {
	Status *string
	Plan   *string
	Cursor string
	Limit  int
}

type OrderRepository interface {
	// Placeholder - to be expanded as needed
}

type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type APICredentialRepository interface {
	GetByPrefix(ctx context.Context, prefix string) (*domain.APICredential, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.APICredential, error)
	Create(ctx context.Context, cred *domain.APICredential) error
	Update(ctx context.Context, cred *domain.APICredential) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error
	ListByShop(ctx context.Context, shopID uuid.UUID) ([]domain.APICredential, error)
}

type OAuthInstallationRepository interface {
	GetByAccessToken(ctx context.Context, tokenHash string) (*domain.OAuthInstallation, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.OAuthInstallation, error)
	Create(ctx context.Context, installation *domain.OAuthInstallation) error
	Update(ctx context.Context, installation *domain.OAuthInstallation) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type Cache interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

type EventPublisher interface {
	Publish(ctx context.Context, event interface{}) error
}

type WebhookRepository interface {
	// Webhook CRUD
	Create(ctx context.Context, webhook *domain.Webhook) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Webhook, error)
	Update(ctx context.Context, webhook *domain.Webhook) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Queries
	ListByOrganization(ctx context.Context, orgID uuid.UUID, filter WebhookFilter) ([]domain.Webhook, error)
	ListByOrganizationAndTopic(ctx context.Context, orgID uuid.UUID, topic string) ([]domain.Webhook, error)
	ListByShop(ctx context.Context, shopID uuid.UUID, filter WebhookFilter) ([]domain.Webhook, error)
	ListByShopAndTopic(ctx context.Context, shopID uuid.UUID, topic string) ([]domain.Webhook, error)

	// Failure tracking
	IncrementFailureCount(ctx context.Context, id uuid.UUID, reason string) error
	ResetFailureCount(ctx context.Context, id uuid.UUID) error
}

type WebhookFilter struct {
	Topic  *string
	Status *domain.WebhookStatus
	Limit  int
	Offset int
}

type WebhookDeliveryRepository interface {
	// Delivery CRUD
	Create(ctx context.Context, delivery *domain.WebhookDelivery) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.WebhookDelivery, error)
	Update(ctx context.Context, delivery *domain.WebhookDelivery) error

	// Queries
	ListByWebhook(ctx context.Context, webhookID uuid.UUID, limit, offset int) ([]domain.WebhookDelivery, error)
	ListPendingRetries(ctx context.Context, limit int) ([]domain.WebhookDelivery, error)

	// Bulk operations
	MarkAsSuccess(ctx context.Context, id uuid.UUID, responseStatus int, responseBody []byte, durationMs int) error
	MarkAsFailed(ctx context.Context, id uuid.UUID, errorMessage string, nextRetryAt *time.Time) error
}

type UnitOfWork interface {
	Begin(ctx context.Context) (Transaction, error)
}

type Transaction interface {
	Products() ProductRepository
	Orders() OrderRepository
	Commit() error
	Rollback() error
}
