package config

import (
	"os"
	"time"
)

type Config struct {
	DB        DBConfig
	JWT       JWTConfig
	Server    ServerConfig
	RateLimit RateLimitConfig
}

type DBConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

type ServerConfig struct {
	Host string
	Port string
}

type RateLimitConfig struct {
	IPRequestsPerMinute    int
	UsernameRequestsPerMin int
	WindowMinutes         int
}

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
			WindowMinutes:         15,
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