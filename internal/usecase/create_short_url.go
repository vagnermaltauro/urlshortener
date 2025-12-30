package usecase

import (
	"context"
	"errors"
	"net/url"
	"time"

	"urlshortner/internal/domain/entity"
	"urlshortner/internal/domain/repository"
)

var (
	// ErrInvalidURL is returned when the provided URL is invalid
	ErrInvalidURL = errors.New("invalid url format")
)

// CreateShortURLUseCase handles the business logic for creating a new short URL
type CreateShortURLUseCase struct {
	urlRepo   repository.URLRepository
	cacheRepo repository.CacheRepository
	idGen     repository.IDGenerator
}

// NewCreateShortURLUseCase creates a new instance of CreateShortURLUseCase
func NewCreateShortURLUseCase(
	urlRepo repository.URLRepository,
	cacheRepo repository.CacheRepository,
	idGen repository.IDGenerator,
) *CreateShortURLUseCase {
	return &CreateShortURLUseCase{
		urlRepo:   urlRepo,
		cacheRepo: cacheRepo,
		idGen:     idGen,
	}
}

// Execute creates a new short URL from the given original URL
func (uc *CreateShortURLUseCase) Execute(ctx context.Context, originalURL string) (*entity.URL, error) {
	// Validate URL format
	if err := uc.validateURL(originalURL); err != nil {
		return nil, err
	}

	// Generate unique ID
	id, err := uc.idGen.Generate()
	if err != nil {
		return nil, err
	}

	// Encode ID to short code
	shortCode := uc.idGen.Encode(id)

	// Create entity with 5-year expiration
	now := time.Now()
	expiresAt := now.AddDate(5, 0, 0) // 5 years from now

	urlEntity := entity.URL{
		ID:          id,
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		Clicks:      0,
		CreatedAt:   now,
		ExpiresAt:   expiresAt,
	}

	// Validate entity
	if !urlEntity.IsValid() {
		return nil, errors.New("invalid url entity created")
	}

	// Save to persistent storage (critical - must succeed)
	if err := uc.urlRepo.Save(ctx, urlEntity); err != nil {
		return nil, err
	}

	// Cache the URL (best-effort - non-fatal if fails)
	// Use 30-day TTL for hot URLs
	cacheTTL := 30 * 24 * time.Hour
	_ = uc.cacheRepo.Set(ctx, shortCode, urlEntity, cacheTTL)

	return &urlEntity, nil
}

// validateURL validates the URL format
func (uc *CreateShortURLUseCase) validateURL(urlStr string) error {
	if urlStr == "" {
		return ErrInvalidURL
	}

	parsedURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return ErrInvalidURL
	}

	// Must have scheme (http or https)
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ErrInvalidURL
	}

	// Must have host
	if parsedURL.Host == "" {
		return ErrInvalidURL
	}

	return nil
}
