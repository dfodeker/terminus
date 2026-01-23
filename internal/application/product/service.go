package product

import (
	"context"
	"fmt"
	"time"

	"github.com/dfodeker/storeos/internal/domain"
	"github.com/dfodeker/storeos/internal/ports"
	"github.com/google/uuid"
)

type Service struct {
	repo   ports.ProductRepository
	cache  ports.Cache
	events ports.EventPublisher
}

func NewService(repo ports.ProductRepository, cache ports.Cache, events ports.EventPublisher) *Service {
	return &Service{repo: repo, cache: cache, events: events}
}

type CreateProductCmd struct {
	ShopID      domain.ShopID
	Title       string
	Description string
	PriceCents  int64
	SKU         string
	Inventory   int
}

func (s *Service) Create(ctx context.Context, cmd CreateProductCmd) (*domain.Product, error) {
	product := &domain.Product{
		ID:          domain.ProductID(uuid.New()),
		ShopID:      cmd.ShopID,
		Title:       cmd.Title,
		Description: cmd.Description,
		PriceCents:  domain.Money(cmd.PriceCents),
		SKU:         cmd.SKU,
		Inventory:   cmd.Inventory,
		Status:      domain.ProductStatusDraft,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}

	// Async event for indexing, notifications, etc.
	s.events.Publish(ctx, domain.ProductCreatedEvent{Product: product})

	return product, nil
}

func (s *Service) GetByID(ctx context.Context, shopID domain.ShopID, id domain.ProductID) (*domain.Product, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("product:%s:%s", shopID, id)
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		return cached.(*domain.Product), nil
	}

	product, err := s.repo.GetByID(ctx, shopID, id)
	if err != nil {
		return nil, err
	}

	// Cache for 5 minutes
	s.cache.Set(ctx, cacheKey, product, 5*time.Minute)

	return product, nil
}
