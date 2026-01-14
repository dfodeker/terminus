// internal/graphql/resolver/node.go
package resolver

import (
	"context"
	"fmt"

	"yourapp/internal/domain"
	"yourapp/internal/graphql/model"
	"yourapp/internal/http/middleware"
)

// Node resolves any object by its GID
func (r *queryResolver) Node(ctx context.Context, id string) (model.Node, error) {
	gid, err := domain.ParseGID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID: %w", err)
	}

	shopID := middleware.ShopIDFromContext(ctx)

	switch gid.Type {
	case domain.GIDTypeProduct:
		return r.ProductService.GetByID(ctx, shopID, gid.ID)

	case domain.GIDTypeProductVariant:
		return r.ProductService.GetVariantByID(ctx, shopID, gid.ID)

	case domain.GIDTypeOrder:
		return r.OrderService.GetByID(ctx, shopID, gid.ID)

	case domain.GIDTypeCustomer:
		return r.CustomerService.GetByID(ctx, shopID, gid.ID)

	default:
		return nil, fmt.Errorf("unknown type: %s", gid.Type)
	}
}

// Nodes resolves multiple objects by their GIDs
func (r *queryResolver) Nodes(ctx context.Context, ids []string) ([]model.Node, error) {
	nodes := make([]model.Node, len(ids))

	for i, id := range ids {
		node, err := r.Node(ctx, id)
		if err != nil {
			// Return nil for not found, don't fail entire query
			nodes[i] = nil
			continue
		}
		nodes[i] = node
	}

	return nodes, nil
}
