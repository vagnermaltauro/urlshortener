package repository

import (
	"context"
	"errors"
	"time"

	"urlshortner/internal/domain/entity"
)

var (
	ErrNotFound = errors.New("url not found")

	ErrDuplicateKey = errors.New("duplicate url id or short code")
)

type URLRepository interface {
	Save(ctx context.Context, url entity.URL) error

	FindByShortCode(ctx context.Context, shortCode string) (*entity.URL, error)

	IncrementClicks(ctx context.Context, shortCode string) error

	BatchIncrementClicks(ctx context.Context, clicks map[string]int64) error

	DeleteExpired(ctx context.Context, before time.Time) (int64, error)
}
