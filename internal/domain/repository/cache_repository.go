package repository

import (
	"context"
	"time"

	"urlshortner/internal/domain/entity"
)

type CacheRepository interface {
	Set(ctx context.Context, shortCode string, url entity.URL, ttl time.Duration) error

	Get(ctx context.Context, shortCode string) (*entity.URL, error)

	Delete(ctx context.Context, shortCode string) error

	IncrementClicks(ctx context.Context, shortCode string) error

	GetPendingClicks(ctx context.Context) (map[string]int64, error)
}
