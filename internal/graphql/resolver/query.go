// internal/graphql/resolver/query.go
package resolver

import (
	"context"

	"yourapp/internal/domain"
	"yourapp/internal/graphql/generated"
	"yourapp/internal/graphql/model"
	"yourapp/internal/graphql/scalar"
	"yourapp/internal/http/middleware"
)

type queryResolver struct{ *Resolver }

func (r *Resolver) Query() generated.QueryResolver {
	return &queryResolver{r}
}

func (r *queryResolver) Shop(ctx context.Context) (*domain.Shop, error) {
	return middleware.ShopFromContext(ctx), nil
}

func (r *queryResolver) Product(ctx context.Context, id string) (*domain.Product, error) {
	gid, err := domain.ParseGID(id)
	if err != nil {
		return nil, err
	}

	if gid.Type != domain.GIDTypeProduct {
		return nil, fmt.Errorf("invalid product ID")
	}

	shopID := middleware.ShopIDFromContext(ctx)
	return r.ProductService.GetByID(ctx, shopID, gid.ID)
}

func (r *queryResolver) Products(
	ctx context.Context,
	first *int,
	after *scalar.Cursor,
	last *int,
	before *scalar.Cursor,
	query *string,
	sortKey *model.ProductSortKey,
	reverse *bool,
) (*model.ProductConnection, error) {
	shopID := middleware.ShopIDFromContext(ctx)

	limit := 20
	if first != nil {
		limit = *first
	}

	var cursor string
	if after != nil {
		cursor, _ = after.Decode()
	}

	products, nextCursor, totalCount, err := r.ProductService.List(ctx, shopID, product.ListParams{
		Limit:   limit,
		Cursor:  cursor,
		Query:   ptrValue(query),
		SortKey: sortKeyToString(sortKey),
		Reverse: ptrValue(reverse),
	})
	if err != nil {
		return nil, err
	}

	return buildProductConnection(products, nextCursor, totalCount, limit), nil
}

func buildProductConnection(products []domain.Product, nextCursor string, totalCount int, limit int) *model.ProductConnection {
	edges := make([]*model.ProductEdge, len(products))
	nodes := make([]*domain.Product, len(products))

	for i := range products {
		p := &products[i]
		nodes[i] = p
		edges[i] = &model.ProductEdge{
			Cursor: scalar.EncodeCursor(p.CreatedAt.Format(time.RFC3339Nano)),
			Node:   p,
		}
	}

	var endCursor *scalar.Cursor
	if len(edges) > 0 {
		c := edges[len(edges)-1].Cursor
		endCursor = &c
	}

	return &model.ProductConnection{
		Edges:      edges,
		Nodes:      nodes,
		TotalCount: totalCount,
		PageInfo: &model.PageInfo{
			HasNextPage:     nextCursor != "",
			HasPreviousPage: false, // Would need backward pagination support
			EndCursor:       endCursor,
		},
	}
}
