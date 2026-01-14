package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/dfodeker/terminus/internal/database"
	"github.com/dfodeker/terminus/internal/gid"
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
	gidGen         *gid.Generator
	baseDomain     string
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

	// Initialize GID generator with machine ID from environment
	machineIDStr := os.Getenv("MACHINE_ID")
	machineID := uint16(0)
	if machineIDStr != "" {
		id, err := strconv.ParseUint(machineIDStr, 10, 16)
		if err != nil {
			log.Fatalf("Invalid MACHINE_ID: %s", err)
		}
		machineID = uint16(id)
	}
	gidGen, err := gid.NewGenerator(machineID)
	if err != nil {
		log.Fatalf("Failed to create GID generator: %s", err)
	}

	baseDomain := os.Getenv("BASE_DOMAIN")
	if baseDomain == "" {
		baseDomain = "storeos.org"
	}

	apiCfg := apiConfig{
		db:         dbQueries,
		platform:   platform,
		port:       port,
		sqlDB:      sqlDB,
		signingKey: signingKey,
		gidGen:     gidGen,
		baseDomain: baseDomain,
	}
	metrics.Register(prometheus.DefaultRegisterer)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
	r := chi.NewRouter()
	// r.Use(middleware.Logger)
	r.Use(middleware.Recoverer) // Recover from panics and log them
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

	// Subdomain and store resolution middleware
	r.Use(mw.Subdomain(mw.SubdomainConfig{
		BaseDomain:     apiCfg.baseDomain,
		APISubdomain:   "api",
		AdminSubdomain: "admin",
	}))
	r.Use(mw.StoreResolver(mw.StoreResolverConfig{
		DB: dbQueries,
	}))

	r.Mount("/debug", middleware.Profiler())
	r.Get("/", homeHandler)
	r.Get("/health", apiCfg.healthHandler)
	r.Get("/metrics", promhttp.Handler().ServeHTTP)

	r.Route("/api/v1", func(r chi.Router) {

		// ============================================================
		// PUBLIC ROUTES (No authentication required)
		// ============================================================

		// POST /api/v1/users - Register a new user
		r.Post("/users", apiCfg.CreateUserHandler)

		// GET /api/v1/users - List all users (public for now)
		r.Get("/users", apiCfg.handlerGetUsers)

		// ============================================================
		// AUTHENTICATED ROUTES
		// ============================================================
		r.Group(func(r chi.Router) {
			r.Use(apiCfg.requireAuth)

			// ------------------------------------------------------------
			// STORE-BASED ROUTES (Primary user workflow)
			// These are the main routes for merchants working with stores
			// ------------------------------------------------------------
			r.Route("/stores", func(r chi.Router) {
				// POST /api/v1/stores - Create a new store
				r.Post("/", apiCfg.handlerCreateStore)

				// GET /api/v1/stores - List all stores for current user
				r.Get("/", apiCfg.handlerGetStores)

				r.Route("/{storeID}", func(r chi.Router) {
					// Products within a store
					r.Route("/products", func(r chi.Router) {
						// POST /api/v1/stores/{storeID}/products - Create product
						r.Post("/", apiCfg.handlerTenantProductCreate)

						// GET /api/v1/stores/{storeID}/products - List products
						r.Get("/", apiCfg.handlerTenantProductsList)

						r.Route("/{productID}", func(r chi.Router) {
							// GET /api/v1/stores/{storeID}/products/{productID} - Get product
							r.Get("/", apiCfg.handlerTenantProductGet)

							// PUT /api/v1/stores/{storeID}/products/{productID} - Update product
							r.Put("/", apiCfg.handlerTenantProductUpdate)

							// DELETE /api/v1/stores/{storeID}/products/{productID} - Delete product
							r.Delete("/", apiCfg.handlerTenantProductDelete)

							// Variants within a product
							r.Route("/variants", func(r chi.Router) {
								// POST /api/v1/stores/{storeID}/products/{productID}/variants - Create variant
								r.Post("/", apiCfg.handlerTenantVariantCreate)

								// GET /api/v1/stores/{storeID}/products/{productID}/variants - List variants
								r.Get("/", apiCfg.handlerTenantVariantsList)

								r.Route("/{variantID}", func(r chi.Router) {
									// PUT /api/v1/stores/{storeID}/products/{productID}/variants/{variantID} - Update variant
									r.Put("/", apiCfg.handlerTenantVariantUpdate)

									// DELETE /api/v1/stores/{storeID}/products/{productID}/variants/{variantID} - Delete variant
									r.Delete("/", apiCfg.handlerTenantVariantDelete)
								})
							})
						})
					})
				})
			})

			// ------------------------------------------------------------
			// VARIANT DIRECT ACCESS (Convenience routes)
			// ------------------------------------------------------------
			r.Route("/variants", func(r chi.Router) {
				// GET /api/v1/variants - List all variants (across stores)
				r.Get("/", apiCfg.handlerTenantVariantsList)

				r.Route("/{variantID}", func(r chi.Router) {
					// GET /api/v1/variants/{variantID} - Get variant by ID
					r.Get("/", apiCfg.handlerVariantGet)
				})
			})

			// ------------------------------------------------------------
			// TENANT ADMIN ROUTES (Organization management)
			// These are for tenant owners/admins to manage their organization
			// ------------------------------------------------------------
			r.Route("/tenants", func(r chi.Router) {
				// POST /api/v1/tenants - Create a new tenant/organization
				r.Post("/", apiCfg.handlerTenantsCreate)

				// GET /api/v1/tenants - List tenants for current user
				r.Get("/", apiCfg.handlerTenantsList)

				r.Route("/{tenantID}", func(r chi.Router) {
					// Store management under tenant
					r.Route("/stores", func(r chi.Router) {
						// POST /api/v1/tenants/{tenantID}/stores - Create store for tenant
						r.Post("/", apiCfg.handlerTenantStoresCreate)

						// GET /api/v1/tenants/{tenantID}/stores - List stores for tenant
						r.Get("/", apiCfg.handlerTenantStoresList)
					})

					// Member management
					r.Route("/members", func(r chi.Router) {
						// GET /api/v1/tenants/{tenantID}/members - List tenant members
						r.Get("/", apiCfg.handlerTenantMembersList)

						// POST /api/v1/tenants/{tenantID}/members/invite - Invite new member
						r.Post("/invite", apiCfg.handlerTenantMembersInvite)

						r.Route("/{memberID}/roles", func(r chi.Router) {
							// POST /api/v1/tenants/{tenantID}/members/{memberID}/roles - Assign role to member
							r.Post("/", apiCfg.handlerTenantMemberAssignRole)

							// DELETE /api/v1/tenants/{tenantID}/members/{memberID}/roles/{roleID} - Remove role from member
							r.Delete("/{roleID}", apiCfg.handlerTenantMemberRemoveRole)
						})
					})

					// Role management
					r.Route("/roles", func(r chi.Router) {
						// POST /api/v1/tenants/{tenantID}/roles - Create role
						r.Post("/", apiCfg.handlerTenantRolesCreate)

						// GET /api/v1/tenants/{tenantID}/roles - List roles
						r.Get("/", apiCfg.handlerTenantRolesList)

						r.Route("/{roleID}/permissions", func(r chi.Router) {
							// POST /api/v1/tenants/{tenantID}/roles/{roleID}/permissions - Add permission to role
							r.Post("/", apiCfg.handlerTenantRoleAddPermission)

							// DELETE /api/v1/tenants/{tenantID}/roles/{roleID}/permissions/{permissionKey} - Remove permission
							r.Delete("/{permissionKey}", apiCfg.handlerTenantRoleRemovePermission)
						})
					})
				})
			})

			// ------------------------------------------------------------
			// GLOBAL RESOURCES
			// ------------------------------------------------------------

			// GET /api/v1/permissions - List all available permissions
			r.Get("/permissions", apiCfg.handlerPermissionsList)
		})
	})

	// ============================================================
	// AUTH ROUTES (No /api/v1 prefix)
	// ============================================================

	// POST /login - Authenticate user and get tokens
	r.Post("/login", apiCfg.handlerLoginUsers)

	// POST /refresh - Refresh access token
	r.Post("/refresh", apiCfg.handlerRefresh)

	// POST /revoke - Revoke refresh token
	r.Post("/revoke", apiCfg.handlerRevoke)

	// ============================================================
	// ADMIN ROUTES (Development/testing only)
	// ============================================================

	// POST /admin/reset - Reset database (dev only)
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
