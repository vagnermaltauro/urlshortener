# CLAUDE.md - URL Shortener Service

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A production-grade URL shortener service built with **Clean Architecture** in Go, designed to handle **100 million URLs per day** (~1,157 writes/sec, ~11,570 reads/sec with 10:1 read/write ratio).

**Key Features:**
- 7-character alphanumeric short codes (Base62 encoding)
- 5-year URL retention with automatic expiration
- High availability 24/7 with horizontal scaling
- PostgreSQL with monthly partitioning for persistent storage
- Redis (single instance) for caching
- Graceful shutdown with connection draining
- Structured logging with zerolog (JSON output)
- Health check endpoints (liveness, readiness, startup)
- Metrics collection and monitoring

**Performance Characteristics:**
- Target: 1,500 writes/sec, 15,000 reads/sec (30% margin)
- p95 latency: <50ms (cache hit), <200ms (cache miss)
- Cache hit ratio: 95%+
- Zero ID collisions (Snowflake algorithm guarantee)

---

## Architecture

### Clean Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│                      HTTP Layer (Adapters)                  │
│  handler/url_handler.go                                     │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                    Use Cases (Business Logic)               │
│  usecase/create_short_url.go, usecase/get_original_url.go  │
│  usecase/increment_clicks.go, usecase/flush_pending_clicks.go │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                  Domain Layer (Core Business)               │
│  domain/entity/url.go                                       │
│  domain/repository/*.go (interfaces only)                   │
└─────────────────────────────────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│              Adapters (Infrastructure Implementations)      │
│  adapter/repository/postgres/url_repository.go              │
│  adapter/repository/redis/cache_repository.go               │
│  adapter/idgen/snowflake.go, adapter/idgen/base62.go        │
└─────────────────────────────────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                Infrastructure (External Services)           │
│  infrastructure/database/postgres.go                        │
│  infrastructure/logger/logger.go                            │
│  infrastructure/config/config.go                            │
│  infrastructure/metrics/metrics.go                          │
└─────────────────────────────────────────────────────────────┘
```

**Dependency Rule:** Dependencies point inward. Domain layer has ZERO external dependencies.

---

## Build and Run

### Prerequisites

- Go 1.24+
- Docker & Docker Compose
- PostgreSQL 17 (via Docker)
- Redis 7 (single instance via Docker)

### Local Development (Recommended: Docker Compose)

```bash
# Start entire infrastructure (PostgreSQL, Redis, App)
docker compose up --build

# The application will be available at:
# - http://localhost:8080
# - Metrics: http://localhost:8080/metrics

# View logs
docker compose logs -f app

# Stop all services
docker compose down

# Clean up volumes (WARNING: deletes all data)
docker compose down -v
```

### Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
# Application
ENVIRONMENT=production
LOG_LEVEL=info
VERSION=2.0.0
MACHINE_ID=1

# PostgreSQL
POSTGRES_DB=urlshortener
POSTGRES_USER=urlshortener
POSTGRES_PASSWORD=change-me-in-production
POSTGRES_PRIMARY_HOST=postgres-primary
POSTGRES_PRIMARY_PORT=5432
POSTGRES_REPLICA_HOSTS=postgres-primary  # Comma-separated for multiple replicas
POSTGRES_SSLMODE=disable

# Redis
REDIS_ADDR=redis:6379

# Server
SERVER_PORT=8080

```

**Note:** `MACHINE_ID` is required for Snowflake IDs (0-1023).

### Manual Build (Native Go)

```bash
# Build binary
CGO_ENABLED=0 go build -ldflags="-s -w" -o urlshortener ./cmd/server

# Run (requires PostgreSQL and Redis already running)
./urlshortener
```

### Docker Build Only

```bash
# Build image
docker build -t urlshortener:latest .

# Run container (requires PostgreSQL and Redis)
docker run -p 8080:8080 \
  -e POSTGRES_PRIMARY_HOST=host.docker.internal \
  -e REDIS_ADDR=host.docker.internal:6379 \
  urlshortener:latest
```

---

## API Reference

### Endpoints

**1. Create Short URL**
```http
POST /api/shorten
Content-Type: application/json

{
  "url": "https://example.com/very-long-url"
}

Response 201:
{
  "short_url": "http://localhost/0Ab3XyZ",
  "short_code": "0Ab3XyZ",
  "expires_at": "2030-12-29T10:30:00Z"
}

Response 400:
{
  "error": "validation_error",
  "message": "URL must start with http:// or https:// and include a host",
  "trace_id": "abc123..."
}
```

**2. Redirect to Original URL**
```http
GET /{shortCode}

Response 301:
Location: https://example.com/very-long-url

Response 404:
{
  "error": "not_found",
  "message": "Short code not found or expired",
  "trace_id": "abc123..."
}
```

**3. Home Page**
```http
GET /

Response 200:
(HTML frontend from web/static/index.html)
```

**4. Metrics**
```http
GET /metrics
Response 200:
{
  "urls_created": 12345,
  "redirects": 123456,
  "cache_hits": 100000,
  "cache_misses": 5000,
  "cache_hit_rate": 0.952,
  "errors": 10,
  "avg_create_duration_ms": 15.3,
  "avg_redirect_duration_ms": 2.1
}
```

---

## ID Generation (Snowflake + Base62)

### Snowflake Algorithm

64-bit distributed ID generation:
```
┌─────────────────┬──────────────┬──────────────┐
│  Timestamp (42) │ MachineID (10)│ Sequence (12)│
└─────────────────┴──────────────┴──────────────┘
  2024-01-01 epoch   1024 machines  4096/ms each
```

**Capacity:**
- 4,096 IDs per millisecond per machine
- 1,024 unique machines
- 139 years of operation (from 2024-01-01)
- **Zero collision guarantee** with proper MACHINE_ID configuration

**Implementation:** `internal/adapter/idgen/snowflake.go`

### Base62 Encoding

Converts 64-bit Snowflake ID to 7-character alphanumeric string:
- Character set: `0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz`
- Fixed length: 7 characters (zero-padded)
- Example: `1234567890` → `0Ab3XyZ`

**Implementation:** `internal/adapter/idgen/base62.go`

---

## PostgreSQL Schema

### Partitioned URLs Table

```sql
CREATE TABLE urls (
    id BIGINT PRIMARY KEY,
    short_code VARCHAR(10) UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    clicks BIGINT DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
) PARTITION BY RANGE (created_at);

-- Monthly partitions (auto-created by partition_manager background job)
CREATE TABLE urls_2025_01 PARTITION OF urls
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
```

**Indexes:**
- `short_code` (UNIQUE, B-tree) - Fast lookups
- `expires_at` (B-tree) - Efficient expiration queries
- `created_at` (B-tree) - Partition key

**Partition Strategy:**
- Monthly partitions for efficient data management
- Background job creates next partition 24 hours before needed
- Old partitions can be dropped after 5 years

**Maintenance Functions:**
```sql
-- Create next month partition
SELECT create_next_partition();

-- Delete expired URLs (run periodically)
SELECT delete_expired_urls();
```

### Connection Pools

**Write DB (Primary):**
- MaxOpenConns: 50
- MaxIdleConns: 25
- ConnMaxLifetime: 5 minutes
- ConnMaxIdleTime: 2 minutes

**Read DB (Replica):**
- MaxOpenConns: 100
- MaxIdleConns: 50
- ConnMaxLifetime: 5 minutes
- ConnMaxIdleTime: 2 minutes

---

## Redis (Single Instance)

### Configuration

**Topology:** Single instance
- redis: `localhost:6379`

**Persistence:** AOF (Append-Only File)
- `appendfsync everysec` - 1-second durability window
- Lazy freeing enabled for better performance

**Eviction:** `allkeys-lru` - Least Recently Used eviction policy

### Data Structures

**1. Cached URLs**
```
Key: "url:{shortCode}"
Type: String (JSON)
Value: {"id": 123, "original_url": "...", "created_at": "...", "expires_at": "..."}
TTL: 30 days
```

**2. Click Buffers**
```
Key: "clicks:buffer:{shortCode}"
Type: Integer (counter)
Value: pending click count
TTL: None (deleted after flush)
```

### Click Buffering Strategy

1. **Redirect Handler:** Atomic INCR on `clicks:buffer:{shortCode}` (no DB write)
2. **Background Job:** Every 10 seconds, flush all buffered clicks to PostgreSQL
3. **Graceful Shutdown:** Final flush before application exits

**Benefits:**
- Reduces PostgreSQL write load by 10-100x
- Sub-millisecond click tracking latency
- Batch updates for better PostgreSQL throughput

---

## Background Jobs

### 1. Click Flusher

**File:** `internal/usecase/background/click_flusher.go`
**Interval:** 10 seconds
**Function:** Flush buffered clicks from Redis to PostgreSQL

```go
// Workflow:
// 1. SCAN all "clicks:buffer:*" keys
// 2. GETDEL each key atomically
// 3. Batch UPDATE PostgreSQL in single transaction
// 4. Log flush count
```

### 2. Partition Manager

**File:** `internal/usecase/background/partition_manager.go`
**Interval:** 24 hours
**Function:** Create next month's PostgreSQL partition

```go
// Workflow:
// 1. Call PostgreSQL function create_next_partition()
// 2. Creates partition for next month if not exists
// 3. Runs on startup and every 24 hours
```

---

## Observability

### Structured Logging

**Library:** zerolog (JSON output)
**Levels:** debug, info, warn, error, fatal
**Configuration:** `LOG_LEVEL` environment variable

**Log Format:**
```json
{
  "level": "info",
  "service": "url-shortener",
  "environment": "production",
  "timestamp": "2025-12-29T10:30:00Z",
  "method": "POST",
  "path": "/api/shorten",
  "status": 201,
  "duration_ms": 15.3,
  "trace_id": "abc123...",
  "message": "Request completed"
}
```

**Implementation:** `internal/infrastructure/logger/logger.go`

### Metrics

**Counters:**
- `urlsCreated` - Total URLs created
- `redirects` - Total redirects performed
- `cacheHits` - Cache hits
- `cacheMisses` - Cache misses
- `errors` - Total errors

**Histograms:** (Last 1000 samples)
- `createDurations` - Create URL latencies
- `redirectDurations` - Redirect latencies

**Calculated:**
- `cache_hit_rate` - cacheHits / (cacheHits + cacheMisses)
- `avg_create_duration_ms` - Average create latency
- `avg_redirect_duration_ms` - Average redirect latency

**Endpoint:** `GET /metrics` (JSON format)

## Deployment

### Docker Compose Architecture

```yaml
Services:
  - postgres-primary: PostgreSQL 17 (write + read)
  - redis: Redis 7 (single instance)
  - app: Application
```

**Volumes:**
- `postgres_data`: PostgreSQL persistent data
- `redis_data`: Redis persistent data

### Resource Limits

**PostgreSQL:**
- Memory: 1GB limit, 512MB reserved
- CPU: 1.0 limit, 0.5 reserved

**Redis:**
- Memory: 256MB limit, 128MB reserved
- CPU: 0.5 limit, 0.25 reserved

**App:**
- Memory: 256MB limit, 128MB reserved
- CPU: 0.5 limit, 0.25 reserved

### Scaling

**Horizontal:**
- For this study project, keep a single app container.

**Vertical (PostgreSQL/Redis):**
- Edit `docker-compose.yml` resource limits
- Tune PostgreSQL settings (shared_buffers, effective_cache_size, etc.)
- Increase Redis maxmemory in `redis.conf`

---

## Performance Tuning

### PostgreSQL Optimization

**Already Applied (docker-compose.yml):**
```sql
shared_buffers = 256MB
effective_cache_size = 1GB
maintenance_work_mem = 64MB
checkpoint_completion_target = 0.9
wal_buffers = 8MB
default_statistics_target = 100
random_page_cost = 1.1  -- SSD optimization
effective_io_concurrency = 200
work_mem = 2MB
min_wal_size = 512MB
max_wal_size = 2GB
```

**Further Tuning:**
1. Increase `shared_buffers` to 25% of system RAM
2. Set `effective_cache_size` to 50-75% of system RAM
3. Adjust `work_mem` based on concurrent query load
4. Monitor with `pg_stat_statements` extension

### Redis Optimization

**Configuration (redis.conf):**
```conf
maxmemory-policy allkeys-lru
appendonly yes
appendfsync everysec
lazyfree-lazy-eviction yes
lazyfree-lazy-expire yes
slowlog-log-slower-than 10000  # 10ms
```

**Monitoring:**
```bash
# Ping Redis
docker exec redis redis-cli ping

# View slowlog
docker exec redis redis-cli slowlog get 10

# Monitor memory
docker exec redis redis-cli info memory
```

### Application Optimization

**Connection Pooling:**
- Write DB: 50 connections
- Read DB: 100 connections (higher for 10:1 read/write ratio)
- Redis: 100 connections

**Cache Strategy:**
- 30-day TTL for URL cache
- No TTL for click buffers (flushed every 10s)
- Async cache backfill (non-blocking)

**Middleware:**
- Gzip compression (level 5)
- Request timeout: 30 seconds
- Read/Write timeout: 15 seconds
- Idle timeout: 60 seconds

---

## Troubleshooting

### Common Issues

**1. App fails to start: "connection refused" (PostgreSQL)**
```bash
# Check PostgreSQL
docker compose ps postgres-primary
docker compose logs postgres-primary

# Verify network connectivity
docker exec -it <app-container> nc -zv postgres-primary 5432

# Solution: Wait for PostgreSQL to be ready
```

**2. App fails to start: "connection refused" (Redis)**
```bash
# Check Redis
docker exec redis redis-cli ping

# Check Redis logs
docker compose logs redis
```

**3. High latency on redirects**
```bash
# Check cache hit rate
curl http://localhost/metrics | jq .cache_hit_rate

# If <90%: Check Redis connectivity
docker exec redis redis-cli ping

# Check Redis memory usage
docker exec redis redis-cli info memory

# Solution: Increase Redis memory or adjust eviction policy
```

**4. Duplicate short codes (collision)**
```bash
# Check MACHINE_ID
docker compose exec app printenv MACHINE_ID

# Ensure MACHINE_ID is set (0-1023)
# Solution: Set MACHINE_ID in .env
```

**5. Slow PostgreSQL queries**
```bash
# Enable slow query logging
docker exec postgres-primary psql -U urlshortener -c "ALTER SYSTEM SET log_min_duration_statement = 1000;"
docker exec postgres-primary psql -U urlshortener -c "SELECT pg_reload_conf();"

# Check logs
docker compose logs postgres-primary | grep "duration:"

# Solution: Add missing indexes or tune configuration
```

### Debugging Commands

**PostgreSQL:**
```bash
# Connect to database
docker exec -it postgres-primary psql -U urlshortener

# Check replication status
SELECT * FROM pg_stat_replication;

# Check partition sizes
SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename))
FROM pg_tables WHERE tablename LIKE 'urls_%' ORDER BY tablename;

# Check active connections
SELECT count(*) FROM pg_stat_activity WHERE datname = 'urlshortener';
```

**Redis:**
```bash
# Connect to Redis
docker exec -it redis redis-cli

# Count keys by pattern
docker exec redis redis-cli --scan --pattern 'url:*' | wc -l
docker exec redis redis-cli --scan --pattern 'clicks:buffer:*' | wc -l
```

**Application:**
```bash
# Follow logs
docker compose logs -f app

# Check metrics
curl http://localhost:8080/metrics | jq .

# Test create URL
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/test"}' | jq .
```

---

## Testing

### Manual Testing

```bash
# 1. Create short URL
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com"}' | jq .

# Response:
# {
#   "short_url": "http://localhost/0Ab3XyZ",
#   "short_code": "0Ab3XyZ",
#   "expires_at": "2030-12-29T10:30:00Z"
# }

# 2. Test redirect
curl -I http://localhost:8080/0Ab3XyZ

# Response:
# HTTP/1.1 301 Moved Permanently
# Location: https://github.com

# 3. Test expiration (URL not found after 5 years)
curl -I http://localhost:8080/ExpiredCode
# Response: HTTP/1.1 404 Not Found
```

### Load Testing

**Script:** `cmd/loadtest/main.go` (to be implemented)

```bash
# Example with Apache Bench
ab -n 10000 -c 100 -p payload.json -T application/json http://localhost:8080/api/shorten

# Example with wrk
wrk -t4 -c100 -d30s http://localhost:8080/api/shorten
```

**Expected Performance:**
- Writes: 1,500+ req/s (create URLs)
- Reads: 15,000+ req/s (redirects with cache hit)
- p95 latency: <50ms (cached), <200ms (uncached)

---

## Key Files Reference

### Domain Layer (Zero External Dependencies)
- `internal/domain/entity/url.go` - URL entity
- `internal/domain/repository/url_repository.go` - URL repository interface
- `internal/domain/repository/cache_repository.go` - Cache repository interface
- `internal/domain/repository/idgen_repository.go` - ID generator interface

### Use Cases (Business Logic)
- `internal/usecase/create_short_url.go` - Create short URL workflow
- `internal/usecase/get_original_url.go` - Get original URL workflow
- `internal/usecase/increment_clicks.go` - Increment click counter
- `internal/usecase/flush_pending_clicks.go` - Flush buffered clicks to PostgreSQL

### Adapters (Infrastructure Implementations)
- `internal/adapter/idgen/snowflake.go` - Snowflake ID generator
- `internal/adapter/idgen/base62.go` - Base62 encoder/decoder
- `internal/adapter/repository/postgres/url_repository.go` - PostgreSQL implementation
- `internal/adapter/repository/redis/cache_repository.go` - Redis cache implementation
- `internal/adapter/http/handler/url_handler.go` - HTTP handlers
- `internal/adapter/http/dto/*.go` - Request/Response DTOs

### Infrastructure
- `internal/infrastructure/database/postgres.go` - PostgreSQL connection pooling
- `internal/infrastructure/logger/logger.go` - Structured logging
- `internal/infrastructure/logger/middleware.go` - HTTP logging middleware
- `internal/infrastructure/config/config.go` - Environment configuration
- `internal/infrastructure/metrics/metrics.go` - Metrics collection

### Background Jobs
- `internal/usecase/background/click_flusher.go` - Click buffer flusher
- `internal/usecase/background/partition_manager.go` - PostgreSQL partition manager

### Entry Point
- `cmd/server/main.go` - Application bootstrap with dependency injection

### Database
- `migrations/001_initial_schema.up.sql` - PostgreSQL schema
- `migrations/001_initial_schema.down.sql` - Rollback migration

### Infrastructure as Code
- `docker-compose.yml` - Local orchestration
- `Dockerfile` - Multi-stage build
- `redis.conf` - Redis single instance configuration
- `.env.example` - Environment variables template

### Frontend
- `web/static/index.html` - User interface

---

## Security Considerations

**1. Input Validation:**
- All URLs validated (must start with http:// or https://, must include host)
- Use case layer enforces validation before persistence

**2. SQL Injection:**
- All queries use prepared statements
- No string concatenation for SQL queries

**3. Rate Limiting:**
- Not included in this repo; add a reverse proxy or gateway in production

**4. Secrets Management:**
- PostgreSQL password via environment variable
- NEVER commit `.env` file to repository
- Use secrets management (Vault, AWS Secrets Manager) in production

**5. TLS/SSL:**
- Production: Terminate HTTPS at your reverse proxy/gateway
- Database: Enable `sslmode=require` for PostgreSQL in production

**6. CORS:**
- Currently allows all origins (development)
- Production: Configure CORS middleware in chi router

---

## Migration from Previous Version

**Breaking Changes:**
- SQLite → PostgreSQL (schema incompatible)
- hashids → Snowflake + Base62 (ID format incompatible)
- Redis Cluster → Redis single instance (configuration change)

**Migration Strategy:**
1. Deploy new system in parallel
2. Redirect new URL creation to new system
3. Maintain read-only old system for existing URLs
4. Gradual cutover after validation period
5. Export old SQLite data and bulk import to PostgreSQL (optional)

**Data Export (Old System):**
```sql
-- SQLite export
sqlite3 urls.db ".mode csv" ".output urls.csv" "SELECT * FROM urls;"

-- PostgreSQL import (adapt schema)
COPY urls FROM '/path/to/urls.csv' CSV HEADER;
```

---

## Production Checklist

- [ ] Change `POSTGRES_PASSWORD` from default
- [ ] Set `MACHINE_ID` (0-1023)
- [ ] Configure PostgreSQL `sslmode=require`
- [ ] Enable HTTPS at your reverse proxy/gateway
- [ ] Configure CORS policy
- [ ] Set up monitoring (Prometheus + Grafana)
- [ ] Configure log aggregation (ELK, Loki, CloudWatch)
- [ ] Set up alerting for errors and latency
- [ ] Configure backup strategy (PostgreSQL, Redis snapshots)
- [ ] Test disaster recovery procedures
- [ ] Implement PostgreSQL read replicas (optional)
- [ ] Implement Redis replicas (optional)
- [ ] Load test with production-like traffic
- [ ] Configure firewall rules (only expose ports 80/443)
- [ ] Review resource limits based on actual usage

---

## License

MIT License (update as needed)

## Support

For issues or questions, open a GitHub issue or contact the development team.
