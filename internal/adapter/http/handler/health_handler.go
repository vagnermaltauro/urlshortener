package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"urlshortner/internal/adapter/http/dto"
	"urlshortner/internal/adapter/repository/redis"
	"urlshortner/internal/infrastructure/metrics"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	postgresDB *sql.DB
	redisRepo  *redis.RedisCacheRepository
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(postgresDB *sql.DB, redisRepo *redis.RedisCacheRepository) *HealthHandler {
	return &HealthHandler{
		postgresDB: postgresDB,
		redisRepo:  redisRepo,
	}
}

// LivenessProbe handles GET /health/live
// Returns 200 if the server is running (simple check)
func (h *HealthHandler) LivenessProbe(w http.ResponseWriter, r *http.Request) {
	resp := dto.HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ReadinessProbe handles GET /health/ready
// Returns 200 if the server is ready to accept traffic (checks dependencies)
func (h *HealthHandler) ReadinessProbe(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	checks := make(map[string]string)
	healthy := true

	// Check PostgreSQL
	if err := h.checkPostgres(ctx); err != nil {
		checks["postgres"] = "unhealthy: " + err.Error()
		healthy = false
	} else {
		checks["postgres"] = "healthy"
	}

	// Check Redis
	if err := h.checkRedis(ctx); err != nil {
		checks["redis"] = "unhealthy: " + err.Error()
		healthy = false
	} else {
		checks["redis"] = "healthy"
	}

	status := "ok"
	statusCode := http.StatusOK
	if !healthy {
		status = "degraded"
		statusCode = http.StatusServiceUnavailable
	}

	resp := dto.HealthResponse{
		Status:    status,
		Timestamp: time.Now().Format(time.RFC3339),
		Checks:    checks,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}

// StartupProbe handles GET /health/startup
// Returns 200 when the server has finished starting up
func (h *HealthHandler) StartupProbe(w http.ResponseWriter, r *http.Request) {
	// For now, same as readiness
	h.ReadinessProbe(w, r)
}

// Metrics handles GET /metrics
// Returns application metrics
func (h *HealthHandler) Metrics(w http.ResponseWriter, r *http.Request) {
	snapshot := metrics.Snapshot()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshot)
}

// checkPostgres checks if PostgreSQL is reachable
func (h *HealthHandler) checkPostgres(ctx context.Context) error {
	return h.postgresDB.PingContext(ctx)
}

// checkRedis checks if Redis is reachable
func (h *HealthHandler) checkRedis(ctx context.Context) error {
	return h.redisRepo.Ping(ctx)
}
