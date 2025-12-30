package background

import (
	"context"
	"log"
	"time"

	"urlshortner/internal/domain/repository"
)

// ClickFlusher is a background job that periodically flushes buffered click counts to PostgreSQL
// This provides optimal performance for the 10:1 read/write ratio while maintaining consistency
type ClickFlusher struct {
	urlRepo   repository.URLRepository
	cacheRepo repository.CacheRepository
	interval  time.Duration
}

// NewClickFlusher creates a new click flusher background job
func NewClickFlusher(
	urlRepo repository.URLRepository,
	cacheRepo repository.CacheRepository,
	interval time.Duration,
) *ClickFlusher {
	return &ClickFlusher{
		urlRepo:   urlRepo,
		cacheRepo: cacheRepo,
		interval:  interval,
	}
}

// Start begins the background flush loop
// This should be called as a goroutine: go clickFlusher.Start(ctx)
func (f *ClickFlusher) Start(ctx context.Context) {
	ticker := time.NewTicker(f.interval)
	defer ticker.Stop()

	log.Printf("[ClickFlusher] Started with interval %v", f.interval)

	for {
		select {
		case <-ctx.Done():
			// Graceful shutdown: flush one last time before exiting
			log.Println("[ClickFlusher] Shutting down, flushing final batch...")
			f.flush(context.Background())
			log.Println("[ClickFlusher] Stopped")
			return

		case <-ticker.C:
			f.flush(ctx)
		}
	}
}

// flush retrieves pending clicks from cache and batch-updates PostgreSQL
func (f *ClickFlusher) flush(ctx context.Context) {
	// Get all buffered clicks from Redis
	clicks, err := f.cacheRepo.GetPendingClicks(ctx)
	if err != nil {
		log.Printf("[ClickFlusher] Error getting pending clicks: %v", err)
		return
	}

	if len(clicks) == 0 {
		return // Nothing to flush
	}

	// Batch update PostgreSQL
	if err := f.urlRepo.BatchIncrementClicks(ctx, clicks); err != nil {
		log.Printf("[ClickFlusher] Error flushing %d click counts: %v", len(clicks), err)
		return
	}

	log.Printf("[ClickFlusher] Successfully flushed %d URL click counts to PostgreSQL", len(clicks))
}
