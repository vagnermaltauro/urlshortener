package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics holds global application metrics
type Metrics struct {
	// Atomic counters (must use atomic operations)
	urlsCreated int64
	redirects   int64
	cacheHits   int64
	cacheMisses int64
	errors      int64

	// Start time for uptime calculation
	startTime time.Time

	// Histograms (protected by mutex)
	mu                sync.RWMutex
	createDurations   []time.Duration
	redirectDurations []time.Duration
}

var global = &Metrics{
	startTime:         time.Now(),
	createDurations:   make([]time.Duration, 0, 1000),
	redirectDurations: make([]time.Duration, 0, 1000),
}

// IncrementURLsCreated increments the URLs created counter
func IncrementURLsCreated() {
	atomic.AddInt64(&global.urlsCreated, 1)
}

// IncrementRedirects increments the redirects counter
func IncrementRedirects() {
	atomic.AddInt64(&global.redirects, 1)
}

// IncrementCacheHits increments the cache hits counter
func IncrementCacheHits() {
	atomic.AddInt64(&global.cacheHits, 1)
}

// IncrementCacheMisses increments the cache misses counter
func IncrementCacheMisses() {
	atomic.AddInt64(&global.cacheMisses, 1)
}

// IncrementErrors increments the errors counter
func IncrementErrors() {
	atomic.AddInt64(&global.errors, 1)
}

// RecordCreateDuration records the duration of a create operation
func RecordCreateDuration(d time.Duration) {
	global.mu.Lock()
	defer global.mu.Unlock()

	// Keep only last 1000 samples
	if len(global.createDurations) >= 1000 {
		global.createDurations = global.createDurations[1:]
	}
	global.createDurations = append(global.createDurations, d)
}

// RecordRedirectDuration records the duration of a redirect operation
func RecordRedirectDuration(d time.Duration) {
	global.mu.Lock()
	defer global.mu.Unlock()

	// Keep only last 1000 samples
	if len(global.redirectDurations) >= 1000 {
		global.redirectDurations = global.redirectDurations[1:]
	}
	global.redirectDurations = append(global.redirectDurations, d)
}

// Snapshot returns a snapshot of all metrics
func Snapshot() map[string]interface{} {
	global.mu.RLock()
	defer global.mu.RUnlock()

	urlsCreated := atomic.LoadInt64(&global.urlsCreated)
	redirects := atomic.LoadInt64(&global.redirects)
	cacheHits := atomic.LoadInt64(&global.cacheHits)
	cacheMisses := atomic.LoadInt64(&global.cacheMisses)
	errors := atomic.LoadInt64(&global.errors)

	// Calculate cache hit rate
	totalCacheRequests := cacheHits + cacheMisses
	var cacheHitRate float64
	if totalCacheRequests > 0 {
		cacheHitRate = float64(cacheHits) / float64(totalCacheRequests)
	}

	return map[string]interface{}{
		"urls_created":    urlsCreated,
		"redirects":       redirects,
		"cache_hits":      cacheHits,
		"cache_misses":    cacheMisses,
		"errors":          errors,
		"cache_hit_rate":  cacheHitRate,
		"uptime_seconds":  int64(time.Since(global.startTime).Seconds()),
	}
}
