package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Environment string
	LogLevel    string
	Version     string
	MachineID   uint16

	PostgresPrimaryHost  string
	PostgresPrimaryPort  int
	PostgresReplicaHosts []string
	PostgresUser         string
	PostgresPassword     string
	PostgresDBName       string
	PostgresSSLMode      string

	RedisAddr         string
	RedisClusterAddrs []string

	ServerPort string
}

func Load() *Config {
	return &Config{

		Environment: getEnv("ENVIRONMENT", "production"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		Version:     getEnv("VERSION", "2.0.0"),
		MachineID:   uint16(getEnvAsInt("MACHINE_ID", 1)),

		PostgresPrimaryHost:  getEnv("POSTGRES_PRIMARY_HOST", "localhost"),
		PostgresPrimaryPort:  getEnvAsInt("POSTGRES_PRIMARY_PORT", 5432),
		PostgresReplicaHosts: getEnvAsSlice("POSTGRES_REPLICA_HOSTS", []string{"localhost"}, ","),
		PostgresUser:         getEnv("POSTGRES_USER", "urlshortener"),
		PostgresPassword:     getEnv("POSTGRES_PASSWORD", "change-me-in-prod"),
		PostgresDBName:       getEnv("POSTGRES_DB", "urlshortener"),
		PostgresSSLMode:      getEnv("POSTGRES_SSLMODE", "disable"),

		RedisAddr:         getEnv("REDIS_ADDR", ""),
		RedisClusterAddrs: getEnvAsSlice("REDIS_CLUSTER_ADDRS", []string{}, ","),

		ServerPort: getEnv("SERVER_PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string, separator string) []string {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	return strings.Split(valueStr, separator)
}
