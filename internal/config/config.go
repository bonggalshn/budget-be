package config

import (
	"os"
	"time"
)

// Config holds all application configuration settings.
type Config struct {
	DB        DBConfig
	JWT       JWTConfig
	Server    ServerConfig
	RateLimit RateLimitConfig
}

// DBConfig holds database connection settings.
// Supported environment variables: DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD
type DBConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

// JWTConfig holds JWT authentication settings.
// Supported environment variables: JWT_SECRET, JWT_EXPIRY
type JWTConfig struct {
	Secret string        // Secret key for signing JWT tokens
	Expiry time.Duration // Token expiration duration
}

// ServerConfig holds HTTP server settings.
// Supported environment variables: SERVER_HOST, SERVER_PORT
type ServerConfig struct {
	Host string // Server listen address
	Port string // Server listen port
}

// RateLimitConfig holds rate limiting settings.
type RateLimitConfig struct {
	IPRequestsPerMinute    int // Maximum requests per IP per minute
	UsernameRequestsPerMin int // Maximum requests per username per window
	WindowMinutes          int // Rate limiting window in minutes
}

// Load returns a Config populated from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			Name:     getEnv("DB_NAME", "budget"),
			User:     getEnv("DB_USER", "budget_user"),
			Password: getEnv("DB_PASSWORD", ""),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "change-me-in-production"),
			Expiry: parseDuration(getEnv("JWT_EXPIRY", "24h"), 24*time.Hour),
		},
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnv("SERVER_PORT", "8080"),
		},
		RateLimit: RateLimitConfig{
			IPRequestsPerMinute:    10,
			UsernameRequestsPerMin: 5,
			WindowMinutes:          15,
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDuration(value string, defaultDuration time.Duration) time.Duration {
	d, err := time.ParseDuration(value)
	if err != nil {
		return defaultDuration
	}
	return d
}
