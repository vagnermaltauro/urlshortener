package usecase

import (
	"context"
	"time"

	"urlshortner/internal/domain/entity"
	"urlshortner/internal/domain/repository"
)

// GetOriginalURLUseCase handles the business logic for retrieving an original URL
type GetOriginalURLUseCase struct {
	urlRepo   repository.URLRepository
	cacheRepo repository.CacheRepository
}

// NewGetOriginalURLUseCase creates a new instance of GetOriginalURLUseCase
func NewGetOriginalURLUseCase(
	urlRepo repository.URLRepository,
	cacheRepo repository.CacheRepository,
) *GetOriginalURLUseCase {
	return &GetOriginalURLUseCase{
		urlRepo:   urlRepo,
		cacheRepo: cacheRepo,
	}
}

// Execute retrieves the original URL for a given short code
// Returns repository.ErrNotFound if the URL doesn't exist or has expired
func (uc *GetOriginalURLUseCase) Execute(ctx context.Context, shortCode string) (*entity.URL, error) {
	// Try cache first (fast path)
	url, err := uc.cacheRepo.Get(ctx, shortCode)
	if err == nil && url != nil {
		// Cache hit - verify not expired
		if !url.IsExpired() {
			return url, nil
		}
		// Expired in cache - delete and fall through to database
		_ = uc.cacheRepo.Delete(ctx, shortCode)
	}

	// Cache miss - query persistent storage (slow path)
	url, err = uc.urlRepo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	if url == nil {
		return nil, repository.ErrNotFound
	}

	// Verify not expired (database query should handle this, but double-check)
	if url.IsExpired() {
		return nil, repository.ErrNotFound
	}

	// Backfill cache asynchronously (non-blocking)
	// Use 30-day TTL for hot URLs
	go func() {
		cacheTTL := 30 * 24 * time.Hour
		_ = uc.cacheRepo.Set(context.Background(), shortCode, *url, cacheTTL)
	}()

	return url, nil
}
