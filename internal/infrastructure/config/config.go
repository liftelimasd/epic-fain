package config

import (
	"os"
	"strconv"
)

// Config holds all application configuration.
type Config struct {
	HTTPAddr string
	TCPAddr  string

	DB DBConfig

	// API keys: hash → owner (loaded from env or DB)
	APIKeys map[string]string
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() Config {
	return Config{
		HTTPAddr: envOrDefault("HTTP_ADDR", ":8080"),
		TCPAddr:  envOrDefault("TCP_ADDR", ":9090"),
		DB: DBConfig{
			Host:     envOrDefault("DB_HOST", "localhost"),
			Port:     envOrDefaultInt("DB_PORT", 5432),
			User:     envOrDefault("DB_USER", "epicfain"),
			Password: envOrDefault("DB_PASSWORD", "epicfain"),
			DBName:   envOrDefault("DB_NAME", "epicfain"),
			SSLMode:  envOrDefault("DB_SSLMODE", "disable"),
		},
		APIKeys: map[string]string{},
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envOrDefaultInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
