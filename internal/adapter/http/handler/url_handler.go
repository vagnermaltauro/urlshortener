package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"urlshortner/internal/adapter/http/dto"
	"urlshortner/internal/domain/repository"
	"urlshortner/internal/infrastructure/logger"
	"urlshortner/internal/infrastructure/metrics"
	"urlshortner/internal/usecase"
)

// URLHandler handles HTTP requests for URL shortening operations
type URLHandler struct {
	createUseCase    *usecase.CreateShortURLUseCase
	getUseCase       *usecase.GetOriginalURLUseCase
	incrementUseCase *usecase.IncrementClicksUseCase
	log              logger.Logger
}

// NewURLHandler creates a new URL handler
func NewURLHandler(
	createUseCase *usecase.CreateShortURLUseCase,
	getUseCase *usecase.GetOriginalURLUseCase,
	incrementUseCase *usecase.IncrementClicksUseCase,
	log logger.Logger,
) *URLHandler {
	return &URLHandler{
		createUseCase:    createUseCase,
		getUseCase:       getUseCase,
		incrementUseCase: incrementUseCase,
		log:              log,
	}
}

// ServeHome serves the static HTML homepage
func (h *URLHandler) ServeHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/static/index.html")
}

// CreateShortURL handles POST /api/shorten
func (h *URLHandler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	traceID := middleware.GetReqID(ctx)
	start := time.Now()

	// Parse request
	var req dto.CreateShortURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		metrics.IncrementErrors()
		h.respondError(w, "invalid_request", "Invalid JSON payload", http.StatusBadRequest, traceID)
		return
	}

	// Validate URL
	if req.URL == "" {
		metrics.IncrementErrors()
		h.respondError(w, "validation_error", "URL is required", http.StatusBadRequest, traceID)
		return
	}

	// Execute use case
	url, err := h.createUseCase.Execute(ctx, req.URL)
	if err != nil {
		metrics.IncrementErrors()
		metrics.RecordCreateDuration(time.Since(start))

		h.log.Error("Failed to create short URL",
			"error", err,
			"trace_id", traceID,
			"duration_ms", time.Since(start).Milliseconds())

		if err == usecase.ErrInvalidURL {
			h.respondError(w, "validation_error", "Invalid URL format", http.StatusBadRequest, traceID)
		} else {
			h.respondError(w, "internal_error", "Failed to create short URL", http.StatusInternalServerError, traceID)
		}
		return
	}

	// Track metrics
	metrics.IncrementURLsCreated()
	metrics.RecordCreateDuration(time.Since(start))

	h.log.Info("Short URL created",
		"short_code", url.ShortCode,
		"trace_id", traceID,
		"duration_ms", time.Since(start).Milliseconds())

	// Build full short URL
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	host := r.Host
	if host == "" {
		host = "localhost:8080"
	}
	shortURL := scheme + "://" + host + "/" + url.ShortCode

	// Return response
	resp := dto.CreateShortURLResponse{
		ShortURL:  shortURL,
		ShortCode: url.ShortCode,
		ExpiresAt: url.ExpiresAt.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// Redirect handles GET /{shortCode}
func (h *URLHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	traceID := middleware.GetReqID(ctx)
	shortCode := chi.URLParam(r, "shortCode")
	start := time.Now()

	if shortCode == "" {
		http.NotFound(w, r)
		return
	}

	// Get original URL
	url, err := h.getUseCase.Execute(ctx, shortCode)
	duration := time.Since(start)

	if err != nil || url == nil {
		metrics.RecordRedirectDuration(duration)

		if err == repository.ErrNotFound {
			h.log.Warn("Short URL not found",
				"short_code", shortCode,
				"trace_id", traceID,
				"duration_ms", duration.Milliseconds())
		} else {
			metrics.IncrementErrors()
			h.log.Error("Error retrieving URL",
				"short_code", shortCode,
				"error", err,
				"trace_id", traceID,
				"duration_ms", duration.Milliseconds())
		}

		http.NotFound(w, r)
		return
	}

	// Increment clicks asynchronously (non-blocking for performance)
	go func() {
		if err := h.incrementUseCase.Execute(ctx, shortCode); err != nil {
			h.log.Error("Failed to increment clicks", "short_code", shortCode, "error", err)
		}
	}()

	// Track metrics
	metrics.IncrementRedirects()
	metrics.RecordRedirectDuration(duration)

	h.log.Info("Redirect performed",
		"short_code", shortCode,
		"original_url", url.OriginalURL,
		"trace_id", traceID,
		"duration_ms", duration.Milliseconds())

	// Redirect to original URL
	http.Redirect(w, r, url.OriginalURL, http.StatusMovedPermanently)
}

// respondError sends an error response
func (h *URLHandler) respondError(w http.ResponseWriter, errorType, message string, status int, traceID string) {
	resp := dto.ErrorResponse{
		Error:   errorType,
		Message: message,
		TraceID: traceID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}
