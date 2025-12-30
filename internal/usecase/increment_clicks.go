package usecase

import (
	"context"

	"urlshortner/internal/domain/repository"
)

// IncrementClicksUseCase handles the business logic for incrementing URL click counters
type IncrementClicksUseCase struct {
	cacheRepo repository.CacheRepository
}

// NewIncrementClicksUseCase creates a new instance of IncrementClicksUseCase
func NewIncrementClicksUseCase(cacheRepo repository.CacheRepository) *IncrementClicksUseCase {
	return &IncrementClicksUseCase{
		cacheRepo: cacheRepo,
	}
}

// Execute buffers a click increment in the cache
// The increment will be batch-flushed to the database by a background job
// This approach provides high performance for the 10:1 read/write ratio
func (uc *IncrementClicksUseCase) Execute(ctx context.Context, shortCode string) error {
	// Buffer the click in cache (atomic operation in Redis)
	// The background FlushPendingClicksUseCase will batch-flush to PostgreSQL
	return uc.cacheRepo.IncrementClicks(ctx, shortCode)
}
