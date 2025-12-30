package repository

import (
	"context"
	"errors"
	"time"

	"urlshortner/internal/domain/entity"
)

var (
	// ErrNotFound is returned when a URL is not found in the repository
	ErrNotFound = errors.New("url not found")

	// ErrDuplicateKey is returned when attempting to save a URL with a duplicate ID or short code
	ErrDuplicateKey = errors.New("duplicate url id or short code")
)

// URLRepository defines the interface for persistent URL storage.
// This interface lives in the domain layer and is implemented by adapters.
type URLRepository interface {
	// Save persists a URL to the storage
	Save(ctx context.Context, url entity.URL) error

	// FindByShortCode retrieves a URL by its short code
	// Returns ErrNotFound if the URL doesn't exist or has expired
	FindByShortCode(ctx context.Context, shortCode string) (*entity.URL, error)

	// IncrementClicks atomically increments the click counter for a URL
	IncrementClicks(ctx context.Context, shortCode string) error

	// BatchIncrementClicks atomically increments click counters for multiple URLs
	// The map key is the short code, the value is the increment amount
	BatchIncrementClicks(ctx context.Context, clicks map[string]int64) error

	// DeleteExpired removes all URLs that expired before the given time
	// Returns the number of URLs deleted
	DeleteExpired(ctx context.Context, before time.Time) (int64, error)
}
