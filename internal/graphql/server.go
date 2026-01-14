// internal/graphql/server.go
package graphql

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"

	"github.com/dfodeker/storeos/internal/graphql/dataloader"
	"github.com/dfodeker/storeos/internal/graphql/directive"
	"github.com/dfodeker/storeos/internal/graphql/generated"
	"github.com/dfodeker/storeos/internal/graphql/resolver"
	"github.com/dfodeker/storeos/internal/http/middleware"
)

type Server struct {
	handler http.Handler
	loaders *dataloader.Loaders
}

func NewServer(
	resolver *resolver.Resolver,
	loaders *dataloader.Loaders,
) *Server {
	// Configure directives
	directives := directive.NewDirectiveConfig()

	// Create executable schema
	cfg := generated.Config{
		Resolvers: resolver,
		Directives: generated.DirectiveRoot{
			RequireScope:    directives.RequireScope,
			RequireAnyScope: directives.RequireAnyScope,
		},
	}

	schema := generated.NewExecutableSchema(cfg)

	// Create handler with options
	srv := handler.New(schema)

	// Transports
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	// Caching
	srv.SetQueryCache(lru.New(1000))
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New(100),
	})

	// Introspection (disable in production if needed)
	srv.Use(extension.Introspection{})

	return &Server{
		handler: srv,
		loaders: loaders,
	}
}

// Routes returns chi router with GraphQL endpoints
func (s *Server) Routes(authMiddleware *middleware.APIAuthMiddleware, rateLimiter *middleware.RateLimiter) chi.Router {
	r := chi.NewRouter()

	// GraphQL endpoint
	r.Route("/graphql", func(r chi.Router) {
		// Auth and rate limiting
		r.Use(authMiddleware.Handler)
		r.Use(rateLimiter.Handler)

		// Inject dataloaders per-request
		r.Use(dataloader.Middleware(s.loaders))

		r.Handle("/", s.handler)
	})

	// Playground (disable in production)
	r.Get("/playground", playground.Handler("GraphQL", "/graphql"))

	return r
}
