FROM golang:1.24-alpine AS builder
RUN apk add --no-cache build-base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /urlshortner ./cmd/server

FROM alpine:3.20
RUN apk --no-cache add ca-certificates sqlite
RUN mkdir -p /data && chown 1000:1000 /data
WORKDIR /app
COPY --from=builder /urlshortner .
COPY web/static ./web/static
USER 1000
EXPOSE 8080
CMD ["./urlshortner"]


