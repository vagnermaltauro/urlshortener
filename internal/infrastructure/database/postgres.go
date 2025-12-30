package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// PostgresConfig holds the configuration for PostgreSQL connection
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string // disable, require, verify-ca, verify-full

	// Connection pool settings for optimal performance
	MaxOpenConns    int           // Maximum number of open connections
	MaxIdleConns    int           // Maximum number of idle connections
	ConnMaxLifetime time.Duration // Maximum lifetime of a connection
	ConnMaxIdleTime time.Duration // Maximum idle time of a connection
}

// NewPostgresConnection creates a new PostgreSQL database connection with connection pooling
func NewPostgresConnection(cfg PostgresConfig) (*sql.DB, error) {
	// Build Data Source Name (DSN)
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode,
	)

	// Open database connection
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool for high performance
	// These settings are optimized for ~1,500 writes/sec and ~15,000 reads/sec
	db.SetMaxOpenConns(cfg.MaxOpenConns)       // Limit concurrent connections
	db.SetMaxIdleConns(cfg.MaxIdleConns)       // Keep connections ready
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime) // Prevent stale connections
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime) // Close unused connections

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// DefaultWriteConfig returns recommended configuration for write (primary) database
func DefaultWriteConfig() PostgresConfig {
	return PostgresConfig{
		MaxOpenConns:    50,              // Allow 50 concurrent writes
		MaxIdleConns:    25,              // Keep 25 connections ready
		ConnMaxLifetime: 5 * time.Minute, // Rotate connections every 5 minutes
		ConnMaxIdleTime: 1 * time.Minute, // Close idle connections after 1 minute
		SSLMode:         "disable",       // For local dev; use "require" in prod
	}
}

// DefaultReadConfig returns recommended configuration for read (replica) database
func DefaultReadConfig() PostgresConfig {
	return PostgresConfig{
		MaxOpenConns:    100,             // Allow more concurrent reads
		MaxIdleConns:    50,              // Keep more connections ready for burst reads
		ConnMaxLifetime: 5 * time.Minute, // Rotate connections every 5 minutes
		ConnMaxIdleTime: 2 * time.Minute, // Keep idle connections longer for read bursts
		SSLMode:         "disable",       // For local dev; use "require" in prod
	}
}
