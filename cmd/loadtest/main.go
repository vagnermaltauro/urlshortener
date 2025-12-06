package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Stats struct {
	totalRequests   int64
	successRequests int64
	failedRequests  int64
	latencies       []time.Duration
	mu              sync.Mutex
}

func (s *Stats) RecordLatency(d time.Duration) {
	s.mu.Lock()
	s.latencies = append(s.latencies, d)
	s.mu.Unlock()
}

func (s *Stats) Percentile(p float64) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.latencies) == 0 {
		return 0
	}

	sorted := make([]time.Duration, len(s.latencies))
	copy(sorted, s.latencies)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	idx := int(float64(len(sorted)) * p)
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func main() {
	baseURL := flag.String("url", "http://localhost", "Base URL of the service")
	duration := flag.Duration("duration", 10*time.Second, "Test duration")
	concurrency := flag.Int("concurrency", 10, "Number of concurrent workers")
	createRatio := flag.Float64("create-ratio", 0.3, "Ratio of create vs read requests (0.0-1.0)")
	flag.Parse()

	fmt.Printf("🚀 Load Test Starting\n")
	fmt.Printf("   URL: %s\n", *baseURL)
	fmt.Printf("   Duration: %v\n", *duration)
	fmt.Printf("   Concurrency: %d\n", *concurrency)
	fmt.Printf("   Create Ratio: %.0f%%\n\n", *createRatio*100)

	stats := &Stats{}
	shortURLs := &sync.Map{}
	client := &http.Client{Timeout: 5 * time.Second}

	var wg sync.WaitGroup
	stop := make(chan struct{})

	// Start workers
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					atomic.AddInt64(&stats.totalRequests, 1)

					start := time.Now()
					var err error

					if rand.Float64() < *createRatio {
						// Create new short URL
						err = createShortURL(client, *baseURL, shortURLs)
					} else {
						// Access existing short URL
						err = accessShortURL(client, *baseURL, shortURLs)
					}

					if err != nil {
						atomic.AddInt64(&stats.failedRequests, 1)
					} else {
						atomic.AddInt64(&stats.successRequests, 1)
						stats.RecordLatency(time.Since(start))
					}
				}
			}
		}(i)
	}

	// Progress ticker
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				total := atomic.LoadInt64(&stats.totalRequests)
				success := atomic.LoadInt64(&stats.successRequests)
				failed := atomic.LoadInt64(&stats.failedRequests)
				fmt.Printf("   📊 Requests: %d (✓ %d / ✗ %d)\n", total, success, failed)
			}
		}
	}()

	// Run for duration
	time.Sleep(*duration)
	close(stop)
	ticker.Stop()
	wg.Wait()

	// Print results
	fmt.Printf("\n" + "═"*50 + "\n")
	fmt.Printf("📈 RESULTS\n")
	fmt.Printf("═"*50 + "\n")
	fmt.Printf("   Total Requests:  %d\n", stats.totalRequests)
	fmt.Printf("   Successful:      %d (%.1f%%)\n", stats.successRequests,
		float64(stats.successRequests)/float64(stats.totalRequests)*100)
	fmt.Printf("   Failed:          %d (%.1f%%)\n", stats.failedRequests,
		float64(stats.failedRequests)/float64(stats.totalRequests)*100)
	fmt.Printf("   Requests/sec:    %.1f\n", float64(stats.totalRequests)/duration.Seconds())
	fmt.Printf("\n   Latency:\n")
	fmt.Printf("      p50: %v\n", stats.Percentile(0.50))
	fmt.Printf("      p95: %v\n", stats.Percentile(0.95))
	fmt.Printf("      p99: %v\n", stats.Percentile(0.99))
	fmt.Printf("═"*50 + "\n")
}

func createShortURL(client *http.Client, baseURL string, store *sync.Map) error {
	url := fmt.Sprintf("https://example.com/page/%d", rand.Intn(100000))
	body, _ := json.Marshal(map[string]string{"url": url})

	resp, err := client.Post(baseURL+"/api/shorten", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if shortURL, ok := result["short_url"]; ok {
		store.Store(rand.Intn(10000), shortURL)
	}
	return nil
}

func accessShortURL(client *http.Client, baseURL string, store *sync.Map) error {
	// Try to get a random stored URL
	var shortURL string
	store.Range(func(key, value interface{}) bool {
		if rand.Float32() < 0.1 { // Random sampling
			shortURL = value.(string)
			return false
		}
		return true
	})

	if shortURL == "" {
		// No URLs stored yet, create one instead
		return createShortURL(client, baseURL, store)
	}

	// Disable redirect to measure just the response time
	noRedirectClient := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := noRedirectClient.Get(shortURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 301 && resp.StatusCode != 302 && resp.StatusCode != 200 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}
