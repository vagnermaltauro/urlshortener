FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary (CGO_ENABLED=0 for pure Go binary - PostgreSQL driver doesn't need CGO)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o /urlshortener ./cmd/server

FROM alpine:3.20

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata wget && \
    adduser -D -u 1000 appuser

WORKDIR /app

# Copy binary and static files
COPY --from=builder /urlshortener .
COPY web/static ./web/static

# Use non-root user
USER appuser

EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health/live || exit 1

CMD ["./urlshortener"]


