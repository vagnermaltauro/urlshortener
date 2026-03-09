package background

import (
	"context"
	"log"
	"time"

	"urlshortner/internal/domain/repository"
)

type ClickFlusher struct {
	urlRepo   repository.URLRepository
	cacheRepo repository.CacheRepository
	interval  time.Duration
}

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

func (f *ClickFlusher) Start(ctx context.Context) {
	ticker := time.NewTicker(f.interval)
	defer ticker.Stop()

	log.Printf("[ClickFlusher] Started with interval %v", f.interval)

	for {
		select {
		case <-ctx.Done():

			log.Println("[ClickFlusher] Shutting down, flushing final batch...")
			f.flush(context.Background())
			log.Println("[ClickFlusher] Stopped")
			return

		case <-ticker.C:
			f.flush(ctx)
		}
	}
}

func (f *ClickFlusher) flush(ctx context.Context) {

	clicks, err := f.cacheRepo.GetPendingClicks(ctx)
	if err != nil {
		log.Printf("[ClickFlusher] Error getting pending clicks: %v", err)
		return
	}

	if len(clicks) == 0 {
		return
	}

	if err := f.urlRepo.BatchIncrementClicks(ctx, clicks); err != nil {
		log.Printf("[ClickFlusher] Error flushing %d click counts: %v", len(clicks), err)
		return
	}

	log.Printf("[ClickFlusher] Successfully flushed %d URL click counts to PostgreSQL", len(clicks))
}
