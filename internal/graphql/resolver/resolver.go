package resolver

// internal/graphql/resolver/resolver.go

import (
	"github.com/dfodeker/storeos/internal/application/customer"
	"github.com/dfodeker/storeos/internal/application/order"
	"github.com/dfodeker/storeos/internal/application/product"
	"github.com/dfodeker/storeos/internal/graphql/dataloader"
)

// Resolver is the root resolver with all dependencies
type Resolver struct {
	ProductService  *product.Service
	OrderService    *order.Service
	CustomerService *customer.Service
	Loaders         *dataloader.Loaders
}

func NewResolver(
	productSvc *product.Service,
	orderSvc *order.Service,
	customerSvc *customer.Service,
	loaders *dataloader.Loaders,
) *Resolver {
	return &Resolver{
		ProductService:  productSvc,
		OrderService:    orderSvc,
		CustomerService: customerSvc,
		Loaders:         loaders,
	}
}
