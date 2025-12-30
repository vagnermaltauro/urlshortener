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

// RedisCacheRepository implements the CacheRepository interface using Redis Cluster
type RedisCacheRepository struct {
	client *redis.ClusterClient
}

// NewRedisCacheRepository creates a new Redis cache repository
// addrs should be the addresses of Redis cluster nodes (e.g., ["redis-master1:6379", "redis-master2:6379", "redis-master3:6379"])
func NewRedisCacheRepository(addrs []string) repository.CacheRepository {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: addrs,

		// Connection pool settings for high performance
		PoolSize:     100, // Maximum number of connections
		MinIdleConns: 20,  // Minimum idle connections to keep ready

		// Timeout settings optimized for sub-50ms latency
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,

		// Retry settings for resilience
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,

		// Route commands to replicas for reads
		RouteByLatency: true,
		RouteRandomly:  false,
	})

	return &RedisCacheRepository{
		client: client,
	}
}

// Set stores a URL in the cache with the given TTL
// The URL is serialized to JSON before storage
func (r *RedisCacheRepository) Set(ctx context.Context, shortCode string, url entity.URL, ttl time.Duration) error {
	key := fmt.Sprintf("url:%s", shortCode)

	// Serialize URL to JSON
	data, err := json.Marshal(url)
	if err != nil {
		return fmt.Errorf("failed to marshal URL: %w", err)
	}

	// Store in Redis with TTL
	return r.client.Set(ctx, key, data, ttl).Err()
}

// Get retrieves a URL from the cache
// Returns repository.ErrNotFound if the URL is not in the cache
func (r *RedisCacheRepository) Get(ctx context.Context, shortCode string) (*entity.URL, error) {
	key := fmt.Sprintf("url:%s", shortCode)

	// Get data from Redis
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get URL from cache: %w", err)
	}

	// Deserialize JSON to URL entity
	var url entity.URL
	if err := json.Unmarshal(data, &url); err != nil {
		return nil, fmt.Errorf("failed to unmarshal URL: %w", err)
	}

	return &url, nil
}

// Delete removes a URL from the cache
func (r *RedisCacheRepository) Delete(ctx context.Context, shortCode string) error {
	key := fmt.Sprintf("url:%s", shortCode)
	return r.client.Del(ctx, key).Err()
}

// IncrementClicks buffers a click increment in Redis
// The increment is stored in a separate key and will be batch-flushed to PostgreSQL
// This provides optimal performance for the 10:1 read/write ratio
func (r *RedisCacheRepository) IncrementClicks(ctx context.Context, shortCode string) error {
	// Use a separate key pattern for buffered clicks
	bufferKey := fmt.Sprintf("clicks:buffer:%s", shortCode)

	// Atomically increment the counter
	return r.client.Incr(ctx, bufferKey).Err()
}

// GetPendingClicks retrieves all buffered click increments and clears them atomically
// This is called by the background job every 10 seconds to flush clicks to PostgreSQL
// Returns a map where the key is the short code and the value is the total increment amount
func (r *RedisCacheRepository) GetPendingClicks(ctx context.Context) (map[string]int64, error) {
	pattern := "clicks:buffer:*"
	clicks := make(map[string]int64)

	// Scan for all buffered click keys (cursor-based iteration)
	iter := r.client.Scan(ctx, 0, pattern, 1000).Iterator()

	for iter.Next(ctx) {
		key := iter.Val()

		// Extract short code from key (remove "clicks:buffer:" prefix)
		shortCode := key[len("clicks:buffer:"):]

		// Get and delete the counter atomically (GETDEL)
		count, err := r.client.GetDel(ctx, key).Int64()
		if err != nil {
			// Skip errors and continue processing other keys
			// This ensures partial failures don't block the entire flush
			continue
		}

		clicks[shortCode] = count
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan buffered clicks: %w", err)
	}

	return clicks, nil
}

// Close closes the Redis client connection
// This should be called during application shutdown
func (r *RedisCacheRepository) Close() error {
	return r.client.Close()
}

// Ping checks if the Redis cluster is reachable
// This is used for health checks
func (r *RedisCacheRepository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
