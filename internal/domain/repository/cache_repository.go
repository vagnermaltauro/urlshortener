package repository

import (
	"context"
	"time"

	"urlshortner/internal/domain/entity"
)

// CacheRepository defines the interface for caching URLs in memory.
// This interface lives in the domain layer and is implemented by adapters (e.g., Redis).
type CacheRepository interface {
	// Set stores a URL in the cache with the given TTL
	Set(ctx context.Context, shortCode string, url entity.URL, ttl time.Duration) error

	// Get retrieves a URL from the cache
	// Returns ErrNotFound if the URL is not in the cache
	Get(ctx context.Context, shortCode string) (*entity.URL, error)

	// Delete removes a URL from the cache
	Delete(ctx context.Context, shortCode string) error

	// IncrementClicks buffers a click increment in the cache
	// These buffered increments will be batch-flushed to persistent storage
	IncrementClicks(ctx context.Context, shortCode string) error

	// GetPendingClicks retrieves all buffered click increments and clears them atomically
	// The map key is the short code, the value is the total increment amount
	// This is used by background jobs to flush buffered clicks to the database
	GetPendingClicks(ctx context.Context) (map[string]int64, error)
}
