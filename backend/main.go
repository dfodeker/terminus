package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/dfodeker/terminus/internal/database"
	"github.com/dfodeker/terminus/internal/metrics"
	mw "github.com/dfodeker/terminus/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	platform       string
	db             *database.Queries
	port           string
	signingKey     string
	sqlDB          *sql.DB
}

func main() {
	godotenv.Load()

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM MUST BE SET")
	}

	signingKey := os.Getenv("SIGNING_KEY")
	if signingKey == "" {
		log.Fatal("SIGNING_KEY must be set")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_Url Must BE SET")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error Loading DB, %s", err)
	}
	defer db.Close()
	dbQueries := database.New(db)
	sqlDB, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error Loading DBConn, %s", err)
	}

	apiCfg := apiConfig{
		db:         dbQueries,
		platform:   platform,
		port:       port,
		sqlDB:      sqlDB,
		signingKey: signingKey,
	}
	metrics.Register(prometheus.DefaultRegisterer)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
	r := chi.NewRouter()
	// r.Use(middleware.Logger)
	// r.Use(middleware.Recoverer)
	r.Use(mw.RequestID)
	r.Use(mw.Metrics)
	if apiCfg.platform == "dev" {
		r.Use(middleware.Logger) // colored, pretty
	} else {
		r.Use(mw.RequestLogger(logger)) // structured for prod
	}
	r.Use(httprate.Limit(
		5,             // requests
		1*time.Second, // per duration
		httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
	))

	r.Mount("/debug", middleware.Profiler())
	r.Get("/", homeHandler)
	r.Get("/health", apiCfg.healthHandler)
	r.Get("/metrics", promhttp.Handler().ServeHTTP)

	r.Route("/api/v1", func(r chi.Router) {

		r.Post("/users", apiCfg.CreateUserHandler)
		r.Get("/users", apiCfg.handlerGetUsers)

		r.Group(func(r chi.Router) {
			r.Use(apiCfg.requireAuth)

			r.Route("/stores", func(r chi.Router) {
				r.Post("/", apiCfg.handlerCreateStore)
				r.Get("/", apiCfg.handlerGetStores)
				r.Route("/{store}", func(r chi.Router) {
					r.Post("/products", apiCfg.handlerCreateProducts)
					r.Get("/products", apiCfg.handlerListProducts)
				})

			})

			// Add more protected routes here and they all get auth automatically.
		})
	})

	r.Post("/login", apiCfg.handlerLoginUsers)
	r.Post("/refresh", apiCfg.handlerRefresh)
	r.Post("/revoke", apiCfg.handlerRevoke)

	r.Post("/admin/reset", apiCfg.handlerReset)

	srv := &http.Server{
		Addr:              ":" + apiCfg.port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("api listening on :%s", apiCfg.port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(fmt.Errorf("listen: %w", err))
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("You've hit our application"))
}
func (cfg *apiConfig) healthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	if err := cfg.sqlDB.PingContext(ctx); err != nil {
		// 503 signals “unhealthy”
		http.Error(w, "db: unavailable", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
