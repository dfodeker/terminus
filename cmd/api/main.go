// cmd/api/main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dfodeker/storeos/internal/adapters/postgres"
	"github.com/dfodeker/storeos/internal/config"
	"github.com/dfodeker/storeos/internal/graphql"
	"github.com/dfodeker/storeos/internal/graphql/dataloader"
	"github.com/dfodeker/storeos/internal/graphql/resolver"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"

	redisadapter "yourapp/internal/adapters/redis"

	httpserver "yourapp/internal/http"
	httphandlers "yourapp/internal/http/handlers"
	"yourapp/internal/http/middleware"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("yolo")
	}

	// =========================================================================
	// Infrastructure
	// =========================================================================

	// Postgres
	pool, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	tenantDB := postgres.NewTenantDB(pool)

	// Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Addr,
	})
	defer redisClient.Close()

	cache := redisadapter.NewCache(redisClient)

	// =========================================================================
	// Repositories
	// =========================================================================

	productRepo := postgres.NewProductRepository(tenantDB)
	// orderRepo := postgres.NewOrderRepository(tenantDB)
	// customerRepo := postgres.NewCustomerRepository(tenantDB)
	shopRepo := postgres.NewShopRepository(pool) // No tenant context needed
	// credentialRepo := postgres.NewAPICredentialRepository(pool)

	// =========================================================================
	// Application Services
	// =========================================================================

	productService := productRepo.NewService(productRepo, cache, nil)
	// orderService := order.NewService(orderRepo, productRepo, customerRepo, cache, nil)
	// customerService := custome.NewService(customerRepo, cache)

	// =========================================================================
	// HTTP Middleware
	// =========================================================================

	authMiddleware := middleware.NewAPIAuthMiddleware(credentialRepo, nil, shopRepo, cache)
	rateLimiter := middleware.NewRateLimiter(redisClient)

	// =========================================================================
	// REST Handlers
	// =========================================================================

	productHandler := httphandlers.NewProductHandler(productService)
	orderHandler := httphandlers.NewOrderHandler(orderService)
	customerHandler := httphandlers.NewCustomerHandler(customerService)

	// =========================================================================
	// GraphQL
	// =========================================================================

	loaders := dataloader.NewLoaders(productRepo, customerRepo)
	gqlResolver := resolver.NewResolver(productService, orderService, customerService, loaders)
	gqlServer := graphql.NewServer(gqlResolver, loaders)

	// =========================================================================
	// Router
	// =========================================================================

	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))

	// Health check (no auth)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// REST API
	r.Mount("/api/2025-01", httpserver.NewRouter(
		authMiddleware,
		rateLimiter,
		productHandler,
		orderHandler,
		customerHandler,
		nil, // webhook handler
	))

	// GraphQL API
	r.Mount("/", gqlServer.Routes(authMiddleware, rateLimiter))

	// =========================================================================
	// Server
	// =========================================================================

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Starting server on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
