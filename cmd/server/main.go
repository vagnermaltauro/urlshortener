package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"urlshortner/internal/adapter/http/handler"
	"urlshortner/internal/adapter/idgen"
	"urlshortner/internal/adapter/repository/postgres"
	"urlshortner/internal/adapter/repository/redis"
	"urlshortner/internal/infrastructure/config"
	"urlshortner/internal/infrastructure/database"
	"urlshortner/internal/infrastructure/logger"
	"urlshortner/internal/usecase"
	"urlshortner/internal/usecase/background"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize structured logger
	appLogger := logger.New(cfg.LogLevel, cfg.Environment)

	appLogger.Info("Starting URL Shortener",
		"version", cfg.Version,
		"environment", cfg.Environment,
		"machine_id", cfg.MachineID)

	// Initialize PostgreSQL connections
	writeDB, readDB := initializePostgreSQL(cfg, appLogger)
	defer writeDB.Close()
	defer readDB.Close()

	// Initialize Redis (supports both single instance and cluster modes)
	cacheRepo := redis.NewRedisCacheRepository(cfg.RedisAddr, cfg.RedisClusterAddrs)

	// Initialize repositories
	urlRepo := postgres.NewPostgresURLRepository(writeDB, readDB)

	// Initialize ID generator
	idGen, err := idgen.NewSnowflakeGenerator(cfg.MachineID)
	if err != nil {
		appLogger.Fatal("Failed to initialize ID generator", "error", err)
	}

	// Initialize use cases
	createURLUseCase := usecase.NewCreateShortURLUseCase(urlRepo, cacheRepo, idGen)
	getURLUseCase := usecase.NewGetOriginalURLUseCase(urlRepo, cacheRepo)
	incrementClicksUseCase := usecase.NewIncrementClicksUseCase(cacheRepo)
	flushClicksUseCase := usecase.NewFlushPendingClicksUseCase(urlRepo, cacheRepo)

	// Initialize HTTP handlers
	urlHandler := handler.NewURLHandler(createURLUseCase, getURLUseCase, incrementClicksUseCase, appLogger)
	healthHandler := handler.NewHealthHandler(writeDB, cacheRepo.(*redis.RedisCacheRepository))

	// Initialize background jobs
	clickFlusher := background.NewClickFlusher(urlRepo, cacheRepo, 10*time.Second)
	partitionMgr := background.NewPartitionManager(writeDB, 24*time.Hour)

	// Start background jobs
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go clickFlusher.Start(ctx)
	go partitionMgr.Start(ctx)

	// Setup HTTP router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(logger.Middleware(appLogger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware.Compress(5)) // Gzip compression

	// Routes
	r.Get("/", urlHandler.ServeHome)
	r.Post("/api/shorten", urlHandler.CreateShortURL)
	r.Get("/{shortCode}", urlHandler.Redirect)

	// Health check endpoints (support both GET and HEAD for Docker healthcheck)
	r.Get("/health/live", healthHandler.LivenessProbe)
	r.Head("/health/live", healthHandler.LivenessProbe)
	r.Get("/health/ready", healthHandler.ReadinessProbe)
	r.Head("/health/ready", healthHandler.ReadinessProbe)
	r.Get("/health/startup", healthHandler.StartupProbe)
	r.Head("/health/startup", healthHandler.StartupProbe)

	// Metrics endpoint
	r.Get("/metrics", healthHandler.Metrics)

	// HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		appLogger.Info("Server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Server failed", "error", err)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	appLogger.Info("Shutdown signal received, draining connections...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("Shutdown failed", "error", err)
	}

	// Stop background jobs
	cancel()

	// Final flush of pending clicks
	if count, err := flushClicksUseCase.Execute(context.Background()); err == nil {
		appLogger.Info("Final click flush completed", "count", count)
	}

	appLogger.Info("Server stopped gracefully")
}

// initializePostgreSQL initializes PostgreSQL write and read connections
func initializePostgreSQL(cfg *config.Config, appLogger logger.Logger) (*sql.DB, *sql.DB) {
	// Write database (primary)
	writeCfg := database.DefaultWriteConfig()
	writeCfg.Host = cfg.PostgresPrimaryHost
	writeCfg.Port = cfg.PostgresPrimaryPort
	writeCfg.User = cfg.PostgresUser
	writeCfg.Password = cfg.PostgresPassword
	writeCfg.DBName = cfg.PostgresDBName
	writeCfg.SSLMode = cfg.PostgresSSLMode

	writeDB, err := database.NewPostgresConnection(writeCfg)
	if err != nil {
		appLogger.Fatal("Failed to connect to PostgreSQL primary", "error", err)
	}

	appLogger.Info("Connected to PostgreSQL primary",
		"host", writeCfg.Host,
		"port", writeCfg.Port,
		"db", writeCfg.DBName)

	// Read database (replica or fallback to primary)
	readCfg := database.DefaultReadConfig()
	if len(cfg.PostgresReplicaHosts) > 0 && cfg.PostgresReplicaHosts[0] != "" {
		readCfg.Host = cfg.PostgresReplicaHosts[0] // Simple round-robin (TODO: load balancing)
	} else {
		readCfg.Host = cfg.PostgresPrimaryHost // Fallback to primary
	}
	readCfg.Port = cfg.PostgresPrimaryPort
	readCfg.User = cfg.PostgresUser
	readCfg.Password = cfg.PostgresPassword
	readCfg.DBName = cfg.PostgresDBName
	readCfg.SSLMode = cfg.PostgresSSLMode

	readDB, err := database.NewPostgresConnection(readCfg)
	if err != nil {
		appLogger.Warn("Failed to connect to PostgreSQL replica, using primary for reads", "error", err)
		readDB = writeDB // Fallback to primary
	} else {
		appLogger.Info("Connected to PostgreSQL replica",
			"host", readCfg.Host,
			"port", readCfg.Port,
			"db", readCfg.DBName)
	}

	return writeDB, readDB
}
