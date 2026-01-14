package dataloader

// internal/graphql/dataloader/loaders.go

import (
	"context"
	"time"

	"github.com/dfodeker/storeos/internal/ports"
	"github.com/google/uuid"
	"github.com/vikstrous/dataloadgen"
)

type Loaders struct {
	VariantsByProductID *dataloadgen.Loader[uuid.UUID, []*domain.ProductVariant]
	CustomerByID        *dataloadgen.Loader[uuid.UUID, *domain.Customer]
	ProductByID         *dataloadgen.Loader[uuid.UUID, *domain.Product]
}

type loaderConfig struct {
	productRepo  ports.ProductRepository
	customerRepo ports.CustomerRepository
}

func NewLoaders(productRepo ports.ProductRepository, customerRepo ports.CustomerRepository) *Loaders {
	cfg := &loaderConfig{
		productRepo:  productRepo,
		customerRepo: customerRepo,
	}

	return &Loaders{
		VariantsByProductID: dataloadgen.NewLoader(
			cfg.loadVariantsByProductID,
			dataloadgen.WithWait(2*time.Millisecond),
		),
		CustomerByID: dataloadgen.NewLoader(
			cfg.loadCustomersByID,
			dataloadgen.WithWait(2*time.Millisecond),
		),
		ProductByID: dataloadgen.NewLoader(
			cfg.loadProductsByID,
			dataloadgen.WithWait(2*time.Millisecond),
		),
	}
}

func (c *loaderConfig) loadVariantsByProductID(ctx context.Context, productIDs []uuid.UUID) ([]*domain.ProductVariant, []error) {
	// Batch load all variants for all product IDs
	variants, err := c.productRepo.ListVariantsByProductIDs(ctx, productIDs)
	if err != nil {
		errors := make([]error, len(productIDs))
		for i := range errors {
			errors[i] = err
		}
		return nil, errors
	}

	// Group by product ID
	variantMap := make(map[uuid.UUID][]*domain.ProductVariant)
	for i := range variants {
		v := &variants[i]
		variantMap[v.ProductID] = append(variantMap[v.ProductID], v)
	}

	// Return in order of input
	results := make([]*domain.ProductVariant, len(productIDs))
	for i, id := range productIDs {
		results[i] = variantMap[id]
	}

	return results, nil
}

func (c *loaderConfig) loadCustomersByID(ctx context.Context, ids []uuid.UUID) ([]*domain.Customer, []error) {
	customers, err := c.customerRepo.ListByIDs(ctx, ids)
	if err != nil {
		errors := make([]error, len(ids))
		for i := range errors {
			errors[i] = err
		}
		return nil, errors
	}

	// Map by ID for ordering
	customerMap := make(map[uuid.UUID]*domain.Customer)
	for i := range customers {
		c := &customers[i]
		customerMap[c.ID] = c
	}

	results := make([]*domain.Customer, len(ids))
	for i, id := range ids {
		results[i] = customerMap[id] // May be nil if not found
	}

	return results, nil
}

func (c *loaderConfig) loadProductsByID(ctx context.Context, ids []uuid.UUID) ([]*domain.Product, []error) {
	products, err := c.productRepo.ListByIDs(ctx, ids)
	if err != nil {
		errors := make([]error, len(ids))
		for i := range errors {
			errors[i] = err
		}
		return nil, errors
	}

	productMap := make(map[uuid.UUID]*domain.Product)
	for i := range products {
		p := &products[i]
		productMap[p.ID] = p
	}

	results := make([]*domain.Product, len(ids))
	for i, id := range ids {
		results[i] = productMap[id]
	}

	return results, nil
}
