// internal/graphql/resolver/product.go
package resolver

import (
	"context"

	"github.com/dfodeker/storeos/internal/domain"
	"github.com/dfodeker/storeos/internal/graphql/generated"
	"github.com/dfodeker/storeos/internal/graphql/model"
	"github.com/dfodeker/storeos/internal/graphql/scalar"
)

type productResolver struct{ *Resolver }

func (r *Resolver) Product() generated.ProductResolver {
	return &productResolver{r}
}

// ID returns the GID for a product
func (r *productResolver) ID(ctx context.Context, obj *domain.Product) (string, error) {
	return domain.ProductGID(obj.ID).Encode(), nil
}

// Variants resolves product variants with dataloader
func (r *productResolver) Variants(
	ctx context.Context,
	obj *domain.Product,
	first *int,
	after *scalar.Cursor,
) (*model.ProductVariantConnection, error) {
	// Use dataloader to batch variant loading
	variants, err := r.Loaders.VariantsByProductID.Load(ctx, obj.ID)
	if err != nil {
		return nil, err
	}

	return buildVariantConnection(variants), nil
}

// PriceRange calculates min/max price from variants
func (r *productResolver) PriceRange(ctx context.Context, obj *domain.Product) (*model.PriceRange, error) {
	variants, err := r.Loaders.VariantsByProductID.Load(ctx, obj.ID)
	if err != nil {
		return nil, err
	}

	if len(variants) == 0 {
		return &model.PriceRange{
			MinPrice: scalar.Money(obj.PriceCents),
			MaxPrice: scalar.Money(obj.PriceCents),
		}, nil
	}

	minPrice := variants[0].PriceCents
	maxPrice := variants[0].PriceCents

	for _, v := range variants {
		if v.PriceCents < minPrice {
			minPrice = v.PriceCents
		}
		if v.PriceCents > maxPrice {
			maxPrice = v.PriceCents
		}
	}

	return &model.PriceRange{
		MinPrice: scalar.Money(minPrice),
		MaxPrice: scalar.Money(maxPrice),
	}, nil
}

// Status converts domain status to GraphQL enum
func (r *productResolver) Status(ctx context.Context, obj *domain.Product) (model.ProductStatus, error) {
	switch obj.Status {
	case domain.ProductStatusDraft:
		return model.ProductStatusDraft, nil
	case domain.ProductStatusActive:
		return model.ProductStatusActive, nil
	case domain.ProductStatusArchived:
		return model.ProductStatusArchived, nil
	default:
		return model.ProductStatusDraft, nil
	}
}
