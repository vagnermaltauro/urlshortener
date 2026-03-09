package usecase

import (
	"context"
	"errors"
	"net/url"
	"time"

	"urlshortner/internal/domain/entity"
	"urlshortner/internal/domain/repository"
)

var (
	ErrInvalidURL = errors.New("invalid url format")
)

type CreateShortURLUseCase struct {
	urlRepo   repository.URLRepository
	cacheRepo repository.CacheRepository
	idGen     repository.IDGenerator
}

func NewCreateShortURLUseCase(
	urlRepo repository.URLRepository,
	cacheRepo repository.CacheRepository,
	idGen repository.IDGenerator,
) *CreateShortURLUseCase {
	return &CreateShortURLUseCase{
		urlRepo:   urlRepo,
		cacheRepo: cacheRepo,
		idGen:     idGen,
	}
}

func (uc *CreateShortURLUseCase) Execute(ctx context.Context, originalURL string) (*entity.URL, error) {

	if err := uc.validateURL(originalURL); err != nil {
		return nil, err
	}

	id, err := uc.idGen.Generate()
	if err != nil {
		return nil, err
	}

	shortCode := uc.idGen.Encode(id)

	now := time.Now()
	expiresAt := now.AddDate(5, 0, 0)

	urlEntity := entity.URL{
		ID:          id,
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		Clicks:      0,
		CreatedAt:   now,
		ExpiresAt:   expiresAt,
	}

	if !urlEntity.IsValid() {
		return nil, errors.New("invalid url entity created")
	}

	if err := uc.urlRepo.Save(ctx, urlEntity); err != nil {
		return nil, err
	}

	cacheTTL := 30 * 24 * time.Hour
	_ = uc.cacheRepo.Set(ctx, shortCode, urlEntity, cacheTTL)

	return &urlEntity, nil
}

func (uc *CreateShortURLUseCase) validateURL(urlStr string) error {
	if urlStr == "" {
		return ErrInvalidURL
	}

	parsedURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return ErrInvalidURL
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ErrInvalidURL
	}

	if parsedURL.Host == "" {
		return ErrInvalidURL
	}

	return nil
}
