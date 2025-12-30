package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration
type Config struct {
	// Application settings
	Environment string
	LogLevel    string
	Version     string
	MachineID   uint16

	// PostgreSQL settings
	PostgresPrimaryHost  string
	PostgresPrimaryPort  int
	PostgresReplicaHosts []string
	PostgresUser         string
	PostgresPassword     string
	PostgresDBName       string
	PostgresSSLMode      string

	// Redis settings (supports both single instance and cluster)
	RedisAddr         string   // Single Redis instance address (e.g., "localhost:6379")
	RedisClusterAddrs []string // Redis Cluster addresses (for cluster mode)

	// Server settings
	ServerPort string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		// Application
		Environment: getEnv("ENVIRONMENT", "production"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		Version:     getEnv("VERSION", "2.0.0"),
		MachineID:   uint16(getEnvAsInt("MACHINE_ID", 1)),

		// PostgreSQL
		PostgresPrimaryHost:  getEnv("POSTGRES_PRIMARY_HOST", "localhost"),
		PostgresPrimaryPort:  getEnvAsInt("POSTGRES_PRIMARY_PORT", 5432),
		PostgresReplicaHosts: getEnvAsSlice("POSTGRES_REPLICA_HOSTS", []string{"localhost"}, ","),
		PostgresUser:         getEnv("POSTGRES_USER", "urlshortener"),
		PostgresPassword:     getEnv("POSTGRES_PASSWORD", "change-me-in-prod"),
		PostgresDBName:       getEnv("POSTGRES_DB", "urlshortener"),
		PostgresSSLMode:      getEnv("POSTGRES_SSLMODE", "disable"),

		// Redis (single instance or cluster)
		RedisAddr:         getEnv("REDIS_ADDR", ""),
		RedisClusterAddrs: getEnvAsSlice("REDIS_CLUSTER_ADDRS", []string{}, ","),

		// Server
		ServerPort: getEnv("SERVER_PORT", "8080"),
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as int or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// getEnvAsSlice gets an environment variable as a slice of strings
func getEnvAsSlice(key string, defaultValue []string, separator string) []string {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	return strings.Split(valueStr, separator)
}
