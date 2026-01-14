// internal/graphql/resolver/mutation.go
package resolver

import (
	"context"

	"github.com/dfodeker/storeos/internal/application/product"
	"github.com/dfodeker/storeos/internal/domain"
	"github.com/dfodeker/storeos/internal/graphql/generated"
	"github.com/dfodeker/storeos/internal/graphql/model"
	"github.com/dfodeker/storeos/internal/http/middleware"
)

type mutationResolver struct{ *Resolver }

func (r *Resolver) Mutation() generated.MutationResolver {
	return &mutationResolver{r}
}

func (r *mutationResolver) ProductCreate(ctx context.Context, input model.ProductInput) (*model.ProductCreatePayload, error) {
	shopID := middleware.ShopIDFromContext(ctx)

	cmd := product.CreateProductCmd{
		ShopID:      shopID,
		Title:       ptrValue(input.Title),
		Description: ptrValue(input.Description),
		Slug:        ptrValue(input.Slug),
	}

	if input.Status != nil {
		cmd.Status = inputStatusToDomain(*input.Status)
	}

	p, err := r.ProductService.Create(ctx, cmd)
	if err != nil {
		return &model.ProductCreatePayload{
			UserErrors: toUserErrors(err),
		}, nil
	}

	return &model.ProductCreatePayload{
		Product: p,
	}, nil
}

func (r *mutationResolver) ProductUpdate(ctx context.Context, id string, input model.ProductInput) (*model.ProductUpdatePayload, error) {
	productID, err := domain.ParseProductID(id)
	if err != nil {
		return &model.ProductUpdatePayload{
			UserErrors: []*model.UserError{{
				Field:   []string{"id"},
				Message: "Invalid product ID",
			}},
		}, nil
	}

	shopID := middleware.ShopIDFromContext(ctx)

	p, err := r.ProductService.Update(ctx, shopID, productID, product.UpdateProductCmd{
		Title:       input.Title,
		Description: input.Description,
		Slug:        input.Slug,
	})
	if err != nil {
		return &model.ProductUpdatePayload{
			UserErrors: toUserErrors(err),
		}, nil
	}

	return &model.ProductUpdatePayload{
		Product: p,
	}, nil
}

func (r *mutationResolver) ProductDelete(ctx context.Context, id string) (*model.ProductDeletePayload, error) {
	productID, err := domain.ParseProductID(id)
	if err != nil {
		return &model.ProductDeletePayload{
			UserErrors: []*model.UserError{{
				Field:   []string{"id"},
				Message: "Invalid product ID",
			}},
		}, nil
	}

	shopID := middleware.ShopIDFromContext(ctx)

	if err := r.ProductService.Delete(ctx, shopID, productID); err != nil {
		return &model.ProductDeletePayload{
			UserErrors: toUserErrors(err),
		}, nil
	}

	return &model.ProductDeletePayload{
		DeletedProductID: &id,
	}, nil
}

// Helper to convert errors to user errors
func toUserErrors(err error) []*model.UserError {
	// Could do more sophisticated error mapping here
	return []*model.UserError{{
		Message: err.Error(),
	}}
}
