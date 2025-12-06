package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"urlshortner/internal/handler"
	"urlshortner/internal/repository"
	"urlshortner/internal/service"
)

func main() {
	redisRepo := repository.NewRedisRepository()
	urlService := service.NewURLService(redisRepo)
	h := handler.NewURLHandler(urlService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)

	r.Get("/", h.ServeHome)
	r.Post("/api/shorten", h.CreateShortURL)
	r.Get("/{id}", h.Redirect)

	log.Println("Server starting on :8080 - 4 replicas via Traefik")
	log.Fatal(http.ListenAndServe(":8080", r))
}
