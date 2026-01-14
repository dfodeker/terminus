// cmd/api/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dfodeker/storeos/internal/adapters/postgres"
	redisadapter "github.com/dfodeker/storeos/internal/adapters/redis"
	"github.com/dfodeker/storeos/internal/config"
	"github.com/dfodeker/storeos/internal/http/middleware"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	goredis "github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// =========================================================================
	// Infrastructure
	// =========================================================================

	// Postgres - convert config types
	pgCfg := postgres.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Database: cfg.Database.Database,
		MaxConns: cfg.Database.MaxConns,
		MinConns: cfg.Database.MinConns,
	}

	pool, err := postgres.NewPool(ctx, pgCfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Redis
	redisClient := goredis.NewClient(&goredis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	// Test Redis connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("warning: redis connection failed: %v", err)
	}

	cache := redisadapter.NewCache(redisClient)

	// =========================================================================
	// Repositories
	// =========================================================================

	shopRepo := postgres.NewShopRepository(pool)
	userRepo := postgres.NewUserRepository(pool)
	credentialRepo := postgres.NewAPICredentialRepository(pool)
	oauthInstallationRepo := postgres.NewOAuthInstallationRepository(pool)

	// =========================================================================
	// HTTP Middleware
	// =========================================================================

	authMiddleware := middleware.NewAPIAuthMiddleware(credentialRepo, oauthInstallationRepo, shopRepo, cache)
	rateLimiter := middleware.NewRateLimiter(redisClient)

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

	// API routes with auth
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(authMiddleware.Handler)
		r.Use(rateLimiter.Handler)

		// User info endpoint
		r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
			auth := middleware.AuthFromContext(r.Context())
			if auth == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"shop_id": "%s", "auth_type": "%s"}`, auth.ShopID, auth.AuthType)
		})
	})

	// Keep track of unused variables to satisfy compiler
	_ = userRepo

	// =========================================================================
	// Server
	// =========================================================================

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Starting server on %s", addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
