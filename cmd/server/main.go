package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/mattn/go-sqlite3"
	"urlshortner/internal/handler"
	"urlshortner/internal/repository"
	"urlshortner/internal/service"
)

func main() {
	// SQLite path (persistent storage)
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "/data/urls.db"
	}

	// Initialize repositories
	redisRepo := repository.NewRedisRepository()
	sqliteRepo, err := repository.NewSQLiteRepository(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize SQLite: %v", err)
	}

	// Composite: Redis (cache) + SQLite (persistent)
	compositeRepo := repository.NewCompositeRepository(redisRepo, sqliteRepo)

	urlService := service.NewURLService(compositeRepo)
	h := handler.NewURLHandler(urlService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)

	r.Get("/", h.ServeHome)
	r.Post("/api/shorten", h.CreateShortURL)
	r.Get("/{id}", h.Redirect)

	log.Println("Server starting on :8080 - Redis cache + SQLite persistence")
	log.Fatal(http.ListenAndServe(":8080", r))
}

