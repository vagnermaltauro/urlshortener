package repository

import (
    "context"
    "urlshortner/internal/model"
    "github.com/redis/go-redis/v9"
    "os"
    "strconv"
)

type RedisRepository struct {
    client *redis.Client
}

func NewRedisRepository() *RedisRepository {
    addr := os.Getenv("REDIS_ADDR")
    if addr == "" {
        addr = "localhost:6379"
    }
    return &RedisRepository{
        client: redis.NewClient(&redis.Options{Addr: addr}),
    }
}

func (r *RedisRepository) Save(ctx context.Context, url model.URL) error {
    key := "url:" + url.ID
    return r.client.HSet(ctx, key, map[string]interface{}{
        "long":    url.LongURL,
        "clicks":  url.Clicks,
        "created": url.CreatedAt,
    }).Err()
}

func (r *RedisRepository) FindByID(ctx context.Context, id string) (*model.URL, error) {
    key := "url:" + id
    data, err := r.client.HGetAll(ctx, key).Result()
    if err != nil || len(data) == 0 {
        return nil, err
    }
    clicks, _ := strconv.ParseInt(data["clicks"], 10, 64)
    return &model.URL{
        ID:        id,
        LongURL:   data["long"],
        Clicks:    clicks,
        CreatedAt: data["created"],
    }, nil
}

func (r *RedisRepository) IncrementClicks(ctx context.Context, id string) error {
    key := "url:" + id
    return r.client.HIncrBy(ctx, key, "clicks", 1).Err()
}


