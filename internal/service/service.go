package service

import (
    "context"
    "urlshortner/internal/model"
    "urlshortner/internal/shortener"
    "time"
)

type URLRepository interface {
    Save(ctx context.Context, url model.URL) error
    FindByID(ctx context.Context, id string) (*model.URL, error)
}

type URLService struct {
    repo URLRepository
}

func NewURLService(repo URLRepository) *URLService {
    return &URLService{repo: repo}
}

func (s *URLService) CreateShortURL(ctx context.Context, longURL string) (*model.URL, error) {
    id := shortener.Generate()
    url := &model.URL{
        ID:        id,
        LongURL:   longURL,
        CreatedAt: time.Now().Format(time.RFC3339),
    }
    if err := s.repo.Save(ctx, *url); err != nil {
        return nil, err
    }
    return url, nil
}

func (s *URLService) GetURL(ctx context.Context, id string) (*model.URL, error) {
    return s.repo.FindByID(ctx, id)
}
