package usecase

import (
	"context"

	"urlshortner/internal/domain/repository"
)

type IncrementClicksUseCase struct {
	cacheRepo repository.CacheRepository
}

func NewIncrementClicksUseCase(cacheRepo repository.CacheRepository) *IncrementClicksUseCase {
	return &IncrementClicksUseCase{
		cacheRepo: cacheRepo,
	}
}

func (uc *IncrementClicksUseCase) Execute(ctx context.Context, shortCode string) error {

	return uc.cacheRepo.IncrementClicks(ctx, shortCode)
}
