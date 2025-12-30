package usecase

import (
	"context"

	"urlshortner/internal/domain/repository"
)

// FlushPendingClicksUseCase handles the business logic for flushing buffered click counts
// This is executed by a background job every 10 seconds
type FlushPendingClicksUseCase struct {
	urlRepo   repository.URLRepository
	cacheRepo repository.CacheRepository
}

// NewFlushPendingClicksUseCase creates a new instance of FlushPendingClicksUseCase
func NewFlushPendingClicksUseCase(
	urlRepo repository.URLRepository,
	cacheRepo repository.CacheRepository,
) *FlushPendingClicksUseCase {
	return &FlushPendingClicksUseCase{
		urlRepo:   urlRepo,
		cacheRepo: cacheRepo,
	}
}

// Execute retrieves all buffered click increments and flushes them to the database
// Returns the number of URLs updated
func (uc *FlushPendingClicksUseCase) Execute(ctx context.Context) (int, error) {
	// Get all pending clicks from cache (atomic read-and-clear)
	clicks, err := uc.cacheRepo.GetPendingClicks(ctx)
	if err != nil {
		return 0, err
	}

	if len(clicks) == 0 {
		return 0, nil // Nothing to flush
	}

	// Batch update PostgreSQL
	if err := uc.urlRepo.BatchIncrementClicks(ctx, clicks); err != nil {
		return 0, err
	}

	return len(clicks), nil
}
