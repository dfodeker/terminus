package shop

import (
	"context"
	"fmt"
	"time"

	"github.com/dfodeker/storeos/internal/domain"
	"github.com/dfodeker/storeos/internal/ports"
	"github.com/google/uuid"
)

type Service struct {
	repo   ports.ShopRepository
	cache  ports.Cache
	events ports.EventPublisher
}

func NewService(repo ports.ShopRepository, cache ports.Cache, events ports.EventPublisher) *Service {
	return &Service{repo: repo, cache: cache, events: events}
}

type CreateShopCmd struct {
	OrganizationID uuid.UUID
	Name           string
	Handle         string
	Subdomain      string
	CustomDomain   string
	Currency       string
	Locale         string
	Timezone       string
	ShopOwner      string
	Email          string
	Phone          string
	Source         string
}

type UpdateShopCmd struct {
	ID           uuid.UUID
	Name         *string
	CustomDomain *string
	Currency     *string
	Timezone     *string
	ShopOwner    *string
	Email        *string
	Phone        *string
}

// Create creates a new shop
func (s *Service) Create(ctx context.Context, cmd CreateShopCmd) (*domain.Shop, error) {
	now := time.Now()
	shop := &domain.Shop{
		Name:            cmd.Name,
		MyShopifyDomain: cmd.Subdomain,
		Domain:          cmd.CustomDomain,
		Currency:        cmd.Currency,
		Timezone:        cmd.Timezone,
		ShopOwner:       cmd.ShopOwner,
		Email:           cmd.Email,
		CustomerEmail:   cmd.Email,
		Phone:           cmd.Phone,
		Source:          cmd.Source,
		CreatedAt:       &now,
		UpdatedAt:       &now,
		ForceSSL:        true,
	}

	if err := shop.Validate(); err != nil {
		return nil, fmt.Errorf("validate shop: %w", err)
	}

	if err := s.repo.Create(ctx, shop); err != nil {
		return nil, fmt.Errorf("create shop: %w", err)
	}

	// Publish event
	s.events.Publish(ctx, ShopCreatedEvent{Shop: shop})

	return shop, nil
}

// GetByID retrieves a shop by its ID
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Shop, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("shop:%s", id.String())
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		if shop, ok := cached.(*domain.Shop); ok {
			return shop, nil
		}
	}

	shop, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache for 5 minutes
	s.cache.Set(ctx, cacheKey, shop, 5*time.Minute)

	return shop, nil
}

// GetBySubdomain retrieves a shop by its subdomain
func (s *Service) GetBySubdomain(ctx context.Context, subdomain string) (*domain.Shop, error) {
	cacheKey := fmt.Sprintf("shop:subdomain:%s", subdomain)
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		if shop, ok := cached.(*domain.Shop); ok {
			return shop, nil
		}
	}

	shop, err := s.repo.GetBySubdomain(ctx, subdomain)
	if err != nil {
		return nil, err
	}

	s.cache.Set(ctx, cacheKey, shop, 5*time.Minute)

	return shop, nil
}

// GetByCustomDomain retrieves a shop by its custom domain
func (s *Service) GetByCustomDomain(ctx context.Context, customDomain string) (*domain.Shop, error) {
	cacheKey := fmt.Sprintf("shop:domain:%s", customDomain)
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		if shop, ok := cached.(*domain.Shop); ok {
			return shop, nil
		}
	}

	shop, err := s.repo.GetByCustomDomain(ctx, customDomain)
	if err != nil {
		return nil, err
	}

	s.cache.Set(ctx, cacheKey, shop, 5*time.Minute)

	return shop, nil
}

// Update updates an existing shop
func (s *Service) Update(ctx context.Context, cmd UpdateShopCmd) (*domain.Shop, error) {
	shop, err := s.repo.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if cmd.Name != nil {
		shop.Name = *cmd.Name
	}
	if cmd.CustomDomain != nil {
		shop.Domain = *cmd.CustomDomain
	}
	if cmd.Currency != nil {
		shop.Currency = *cmd.Currency
	}
	if cmd.Timezone != nil {
		shop.Timezone = *cmd.Timezone
	}
	if cmd.ShopOwner != nil {
		shop.ShopOwner = *cmd.ShopOwner
	}
	if cmd.Email != nil {
		shop.Email = *cmd.Email
	}
	if cmd.Phone != nil {
		shop.Phone = *cmd.Phone
	}

	now := time.Now()
	shop.UpdatedAt = &now

	if err := s.repo.Update(ctx, shop); err != nil {
		return nil, fmt.Errorf("update shop: %w", err)
	}

	// Invalidate cache
	s.invalidateCache(ctx, shop)

	// Publish event
	s.events.Publish(ctx, ShopUpdatedEvent{Shop: shop})

	return shop, nil
}

// Delete soft-deletes a shop
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	shop, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete shop: %w", err)
	}

	// Invalidate cache
	s.invalidateCache(ctx, shop)

	// Publish event
	s.events.Publish(ctx, ShopDeletedEvent{ShopID: id})

	return nil
}

// List returns a paginated list of shops
func (s *Service) List(ctx context.Context, filter ports.ShopFilter) ([]domain.Shop, string, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) invalidateCache(ctx context.Context, shop *domain.Shop) {
	s.cache.Delete(ctx, fmt.Sprintf("shop:%d", shop.Id))
	if shop.MyShopifyDomain != "" {
		s.cache.Delete(ctx, fmt.Sprintf("shop:subdomain:%s", shop.MyShopifyDomain))
	}
	if shop.Domain != "" {
		s.cache.Delete(ctx, fmt.Sprintf("shop:domain:%s", shop.Domain))
	}
}

// Events

type ShopCreatedEvent struct {
	Shop *domain.Shop
}

type ShopUpdatedEvent struct {
	Shop *domain.Shop
}

type ShopDeletedEvent struct {
	ShopID uuid.UUID
}
