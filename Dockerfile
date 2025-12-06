FROM golang:1.24-alpine AS builder
RUN apk add --no-cache build-base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /urlshortner ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite
WORKDIR /root/
COPY --from=builder /urlshortner .
COPY web/static ./web/static
EXPOSE 8080
CMD ["./urlshortner"]
