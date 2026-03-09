package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"urlshortner/internal/domain/entity"
	"urlshortner/internal/domain/repository"
)

type RedisCacheRepository struct {
	client redis.UniversalClient
}

func NewRedisCacheRepository(singleAddr string, clusterAddrs []string) repository.CacheRepository {
	var client redis.UniversalClient

	if singleAddr != "" {

		client = redis.NewClient(&redis.Options{
			Addr: singleAddr,

			PoolSize:     100,
			MinIdleConns: 20,

			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,

			MaxRetries:      3,
			MinRetryBackoff: 8 * time.Millisecond,
			MaxRetryBackoff: 512 * time.Millisecond,
		})
	} else if len(clusterAddrs) > 0 {

		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: clusterAddrs,

			PoolSize:     100,
			MinIdleConns: 20,

			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,

			MaxRetries:      3,
			MinRetryBackoff: 8 * time.Millisecond,
			MaxRetryBackoff: 512 * time.Millisecond,

			RouteByLatency: true,
			RouteRandomly:  false,
		})
	} else {

		client = redis.NewClient(&redis.Options{
			Addr:         "localhost:6379",
			PoolSize:     100,
			MinIdleConns: 20,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		})
	}

	return &RedisCacheRepository{
		client: client,
	}
}

func (r *RedisCacheRepository) Set(ctx context.Context, shortCode string, url entity.URL, ttl time.Duration) error {
	key := fmt.Sprintf("url:%s", shortCode)

	data, err := json.Marshal(url)
	if err != nil {
		return fmt.Errorf("failed to marshal URL: %w", err)
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisCacheRepository) Get(ctx context.Context, shortCode string) (*entity.URL, error) {
	key := fmt.Sprintf("url:%s", shortCode)

	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get URL from cache: %w", err)
	}

	var url entity.URL
	if err := json.Unmarshal(data, &url); err != nil {
		return nil, fmt.Errorf("failed to unmarshal URL: %w", err)
	}

	return &url, nil
}

func (r *RedisCacheRepository) Delete(ctx context.Context, shortCode string) error {
	key := fmt.Sprintf("url:%s", shortCode)
	return r.client.Del(ctx, key).Err()
}

func (r *RedisCacheRepository) IncrementClicks(ctx context.Context, shortCode string) error {

	bufferKey := fmt.Sprintf("clicks:buffer:%s", shortCode)

	return r.client.Incr(ctx, bufferKey).Err()
}

func (r *RedisCacheRepository) GetPendingClicks(ctx context.Context) (map[string]int64, error) {
	pattern := "clicks:buffer:*"
	clicks := make(map[string]int64)

	iter := r.client.Scan(ctx, 0, pattern, 1000).Iterator()

	for iter.Next(ctx) {
		key := iter.Val()

		shortCode := key[len("clicks:buffer:"):]

		count, err := r.client.GetDel(ctx, key).Int64()
		if err != nil {

			continue
		}

		clicks[shortCode] = count
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan buffered clicks: %w", err)
	}

	return clicks, nil
}

func (r *RedisCacheRepository) Close() error {
	return r.client.Close()
}
