package http

// internal/http/routes.go

import (
	"github.com/dfodeker/storeos/internal/domain"
	"github.com/dfodeker/storeos/internal/http/handlers"
	"github.com/dfodeker/storeos/internal/http/middleware"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(
	authMiddleware *middleware.APIAuthMiddleware,
	rateLimiter *middleware.RateLimiter,
	productHandler *handlers.ProductHandler,
	orderHandler *handlers.OrderHandler,
	customerHandler *handlers.CustomerHandler,
	webhookHandler *handlers.WebhookHandler,
	shopHandler *handlers.ShopHandler,
	organizationHandler *handlers.OrganizationHandler,
) chi.Router {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(middleware.RequestLogger)

	// API routes
	r.Route("/api/2025-01", func(r chi.Router) {
		// Auth required for all API routes
		r.Use(authMiddleware.Handler)
		r.Use(rateLimiter.Handler)

		// Products
		r.Route("/products", func(r chi.Router) {
			r.With(middleware.RequireScopes(domain.ScopeProductsRead)).
				Get("/", productHandler.List)
			r.With(middleware.RequireScopes(domain.ScopeProductsRead)).
				Get("/{productID}", productHandler.GetByID)
			r.With(middleware.RequireScopes(domain.ScopeProductsWrite)).
				Post("/", productHandler.Create)
			r.With(middleware.RequireScopes(domain.ScopeProductsWrite)).
				Put("/{productID}", productHandler.Update)
			r.With(middleware.RequireScopes(domain.ScopeProductsWrite)).
				Delete("/{productID}", productHandler.Delete)

			// Nested variants
			r.Route("/{productID}/variants", func(r chi.Router) {
				r.With(middleware.RequireScopes(domain.ScopeProductsRead)).
					Get("/", productHandler.ListVariants)
				r.With(middleware.RequireScopes(domain.ScopeProductsWrite)).
					Post("/", productHandler.CreateVariant)
			})
		})

		// Orders
		r.Route("/orders", func(r chi.Router) {
			r.With(middleware.RequireScopes(domain.ScopeOrdersRead)).
				Get("/", orderHandler.List)
			r.With(middleware.RequireScopes(domain.ScopeOrdersRead)).
				Get("/{orderID}", orderHandler.GetByID)
			r.With(middleware.RequireScopes(domain.ScopeOrdersWrite)).
				Post("/", orderHandler.Create)
			r.With(middleware.RequireScopes(domain.ScopeOrdersWrite)).
				Put("/{orderID}", orderHandler.Update)
		})

		// Customers
		r.Route("/customers", func(r chi.Router) {
			r.With(middleware.RequireScopes(domain.ScopeCustomersRead)).
				Get("/", customerHandler.List)
			r.With(middleware.RequireScopes(domain.ScopeCustomersRead)).
				Get("/{customerID}", customerHandler.GetByID)
			r.With(middleware.RequireScopes(domain.ScopeCustomersWrite)).
				Post("/", customerHandler.Create)
		})

		// Webhooks
		r.Route("/webhooks", func(r chi.Router) {
			r.With(middleware.RequireScopes(domain.ScopeWebhooksManage)).
				Get("/", webhookHandler.List)
			r.With(middleware.RequireScopes(domain.ScopeWebhooksManage)).
				Post("/", webhookHandler.Create)
			r.With(middleware.RequireScopes(domain.ScopeWebhooksManage)).
				Delete("/{webhookID}", webhookHandler.Delete)
		})

		// Shops
		r.Route("/shops", func(r chi.Router) {
			r.With(middleware.RequireScopes(domain.ScopeShopsRead)).
				Get("/", shopHandler.List)
			r.With(middleware.RequireScopes(domain.ScopeShopsRead)).
				Get("/{shopID}", shopHandler.GetByID)
			r.With(middleware.RequireScopes(domain.ScopeShopsWrite)).
				Post("/", shopHandler.Create)
			r.With(middleware.RequireScopes(domain.ScopeShopsWrite)).
				Put("/{shopID}", shopHandler.Update)
			r.With(middleware.RequireScopes(domain.ScopeShopsWrite)).
				Delete("/{shopID}", shopHandler.Delete)
		})

		// Organizations
		r.Route("/organizations", func(r chi.Router) {
			r.With(middleware.RequireScopes(domain.ScopeOrganizationsRead)).
				Get("/", organizationHandler.List)
			r.With(middleware.RequireScopes(domain.ScopeOrganizationsRead)).
				Get("/{orgID}", organizationHandler.GetByID)
			r.With(middleware.RequireScopes(domain.ScopeOrganizationsWrite)).
				Post("/", organizationHandler.Create)
			r.With(middleware.RequireScopes(domain.ScopeOrganizationsWrite)).
				Put("/{orgID}", organizationHandler.Update)
			r.With(middleware.RequireScopes(domain.ScopeOrganizationsWrite)).
				Delete("/{orgID}", organizationHandler.Delete)

			// Nested shops
			r.With(middleware.RequireScopes(domain.ScopeOrganizationsRead, domain.ScopeShopsRead)).
				Get("/{orgID}/shops", organizationHandler.ListShops)
			r.With(middleware.RequireScopes(domain.ScopeOrganizationsRead)).
				Get("/{orgID}/can-create-shop", organizationHandler.CanCreateShop)
		})

		// Shop info (always readable with any valid auth)
		r.Get("/shop", handlers.GetShop)
	})

	return r
}
