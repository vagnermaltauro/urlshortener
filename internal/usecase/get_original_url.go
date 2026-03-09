package usecase

import (
	"context"
	"time"

	"urlshortner/internal/domain/entity"
	"urlshortner/internal/domain/repository"
	"urlshortner/internal/infrastructure/metrics"
)

type GetOriginalURLUseCase struct {
	urlRepo   repository.URLRepository
	cacheRepo repository.CacheRepository
}

func NewGetOriginalURLUseCase(
	urlRepo repository.URLRepository,
	cacheRepo repository.CacheRepository,
) *GetOriginalURLUseCase {
	return &GetOriginalURLUseCase{
		urlRepo:   urlRepo,
		cacheRepo: cacheRepo,
	}
}

func (uc *GetOriginalURLUseCase) Execute(ctx context.Context, shortCode string) (*entity.URL, error) {

	url, err := uc.cacheRepo.Get(ctx, shortCode)
	if err == nil && url != nil {

		if !url.IsExpired() {
			metrics.IncrementCacheHits()
			return url, nil
		}

		_ = uc.cacheRepo.Delete(ctx, shortCode)
	}

	metrics.IncrementCacheMisses()

	url, err = uc.urlRepo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	if url == nil {
		return nil, repository.ErrNotFound
	}

	if url.IsExpired() {
		return nil, repository.ErrNotFound
	}

	go func() {
		cacheTTL := 30 * 24 * time.Hour
		_ = uc.cacheRepo.Set(context.Background(), shortCode, *url, cacheTTL)
	}()

	return url, nil
}
