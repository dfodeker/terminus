// cmd/api/main.go
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dfodeker/storeos/internal/adapters/postgres"
	redisadapter "github.com/dfodeker/storeos/internal/adapters/redis"
	"github.com/dfodeker/storeos/internal/config"
	"github.com/dfodeker/storeos/internal/http/handlers"
	"github.com/dfodeker/storeos/internal/http/middleware"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	goredis "github.com/redis/go-redis/v9"
)

func main() {
	_ = godotenv.Load()
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// =========================================================================
	// Logger Setup
	// =========================================================================
	var logger *slog.Logger
	logLevel := slog.LevelInfo
	if cfg.Logging.Level == "debug" {
		logLevel = slog.LevelDebug
	} else if cfg.Logging.Level == "warn" {
		logLevel = slog.LevelWarn
	} else if cfg.Logging.Level == "error" {
		logLevel = slog.LevelError
	}

	opts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: cfg.Logging.IncludeCaller,
	}

	if cfg.Logging.Format == "json" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, opts))
	}
	slog.SetDefault(logger)

	// =========================================================================
	// Infrastructure
	// =========================================================================

	// Postgres - convert config types
	pgCfg := postgres.Config{
		URL:      cfg.Database.URL,
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Database: cfg.Database.Database,
		SSLMode:  cfg.Database.SSLMode,
		MaxConns: cfg.Database.MaxConns,
		MinConns: cfg.Database.MinConns,
	}
	logger.Info("connecting to database",
		"host", pgCfg.Host,
		"name", pgCfg.User,
		"port", pgCfg.Port,
		"database", pgCfg.Database,
		"sslmode", pgCfg.SSLMode,
	)

	pool, err := postgres.NewPool(ctx, pgCfg)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Debug: verify database schema
	var colExists bool
	err = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'users' AND column_name = 'password_hash'
		)
	`).Scan(&colExists)
	if err != nil {
		logger.Error("schema check failed", "error", err)
	} else {
		logger.Info("schema check", "password_hash_exists", colExists)
	}

	// Redis
	redisClient := goredis.NewClient(&goredis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	// Test Redis connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Warn("redis connection failed", "error", err)
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
	// Handlers
	// =========================================================================

	userAuthHandler := handlers.NewUserAuthHandler(userRepo, cache, handlers.UserAuthConfig{
		JWTSecret:       cfg.Auth.JWTSecret,
		AccessTokenTTL:  cfg.Auth.JWTAccessTokenTTL,
		RefreshTokenTTL: cfg.Auth.JWTRefreshTokenTTL,
		BcryptCost:      cfg.Auth.BcryptCost,
		MinPasswordLen:  cfg.Auth.PasswordMinLength,
	})

	// =========================================================================
	// Router
	// =========================================================================

	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(middleware.RequestLogger(logger))
	r.Use(chimw.Timeout(30 * time.Second))

	// Health check (no auth)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// =========================================================================
	// Auth routes (no auth required)
	// =========================================================================
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", userAuthHandler.Register)
		r.Post("/login", userAuthHandler.Login)
		r.Post("/refresh", userAuthHandler.RefreshToken)
		r.Post("/logout", userAuthHandler.Logout)
	})

	// =========================================================================
	// User routes (JWT auth for dashboard/admin)
	// =========================================================================
	userAuthMiddleware := middleware.NewRequireAuthMiddleware(cfg.Auth.JWTSecret, userRepo)
	r.Route("/api/v1/user", func(r chi.Router) {
		r.Use(userAuthMiddleware.Handler)

		// Current user info endpoint
		r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
			userCtx := middleware.UserFromContext(r.Context())
			if userCtx == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"user_id": "%s", "email": "%s"}`, userCtx.UserID, userCtx.Email)
		})
	})

	// =========================================================================
	// Shop API routes (API keys and OAuth tokens for integrations)
	// =========================================================================
	r.Route("/api/v1/shop", func(r chi.Router) {
		r.Use(authMiddleware.Handler)
		r.Use(rateLimiter.Handler)

		// Shop info endpoint
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
		logger.Info("starting server", "addr", addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped")
}
