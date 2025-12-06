package repository

import (
	"context"
	"log"
	"urlshortner/internal/model"
)


type CompositeRepository struct {
	redis  *RedisRepository
	sqlite *SQLiteRepository
}

func NewCompositeRepository(redis *RedisRepository, sqlite *SQLiteRepository) *CompositeRepository {
	return &CompositeRepository{
		redis:  redis,
		sqlite: sqlite,
	}
}

func (c *CompositeRepository) Save(ctx context.Context, url model.URL) error {
	if err := c.sqlite.Save(ctx, url); err != nil {
		log.Printf("SQLite save error: %v", err)
		return err
	}

	if err := c.redis.Save(ctx, url); err != nil {
		log.Printf("Redis cache error (non-fatal): %v", err)
	}

	return nil
}

func (c *CompositeRepository) FindByID(ctx context.Context, id string) (*model.URL, error) {
	url, err := c.redis.FindByID(ctx, id)
	if err == nil && url != nil {
		return url, nil
	}
	url, err = c.sqlite.FindByID(ctx, id)
	if err != nil || url == nil {
		return nil, err
	}
	_ = c.redis.Save(ctx, *url)

	return url, nil
}
