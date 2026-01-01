package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/dfodeker/terminus/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	platform       string
	db             *database.Queries
	port           string
	signingKey     string
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

	apiCfg := apiConfig{
		db:         dbQueries,
		platform:   platform,
		port:       port,
		signingKey: signingKey,
	}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", homeHandler)
	r.Get("/health", healthHandler)

	r.Post("/users", apiCfg.CreateUserHandler)
	r.Get("/users", apiCfg.handlerGetUsers)

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
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
