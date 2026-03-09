package usecase

import (
	"context"

	"urlshortner/internal/domain/repository"
)

type FlushPendingClicksUseCase struct {
	urlRepo   repository.URLRepository
	cacheRepo repository.CacheRepository
}

func NewFlushPendingClicksUseCase(
	urlRepo repository.URLRepository,
	cacheRepo repository.CacheRepository,
) *FlushPendingClicksUseCase {
	return &FlushPendingClicksUseCase{
		urlRepo:   urlRepo,
		cacheRepo: cacheRepo,
	}
}

func (uc *FlushPendingClicksUseCase) Execute(ctx context.Context) (int, error) {

	clicks, err := uc.cacheRepo.GetPendingClicks(ctx)
	if err != nil {
		return 0, err
	}

	if len(clicks) == 0 {
		return 0, nil
	}

	if err := uc.urlRepo.BatchIncrementClicks(ctx, clicks); err != nil {
		return 0, err
	}

	return len(clicks), nil
}
