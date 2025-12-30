package dto

// CreateShortURLResponse represents the HTTP response for a successfully created short URL
type CreateShortURLResponse struct {
	ShortURL  string `json:"short_url"`
	ShortCode string `json:"short_code"`
	ExpiresAt string `json:"expires_at"`
}

// ErrorResponse represents an HTTP error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	TraceID string `json:"trace_id"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Checks    map[string]string `json:"checks,omitempty"`
}

// MetricsResponse represents the metrics endpoint response
type MetricsResponse struct {
	URLsCreated   int64   `json:"urls_created"`
	Redirects     int64   `json:"redirects"`
	CacheHits     int64   `json:"cache_hits"`
	CacheMisses   int64   `json:"cache_misses"`
	Errors        int64   `json:"errors"`
	CacheHitRate  float64 `json:"cache_hit_rate"`
	UptimeSeconds int64   `json:"uptime_seconds"`
}
