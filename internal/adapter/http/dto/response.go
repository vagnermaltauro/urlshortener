package dto

type CreateShortURLResponse struct {
	ShortURL  string `json:"short_url"`
	ShortCode string `json:"short_code"`
	ExpiresAt string `json:"expires_at"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	TraceID string `json:"trace_id"`
}

type MetricsResponse struct {
	URLsCreated   int64   `json:"urls_created"`
	Redirects     int64   `json:"redirects"`
	CacheHits     int64   `json:"cache_hits"`
	CacheMisses   int64   `json:"cache_misses"`
	Errors        int64   `json:"errors"`
	CacheHitRate  float64 `json:"cache_hit_rate"`
	UptimeSeconds int64   `json:"uptime_seconds"`
}
