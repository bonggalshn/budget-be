package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration settings.
type Config struct {
	DB        DBConfig        `yaml:"database"`
	JWT       JWTConfig       `yaml:"jwt"`
	Server    ServerConfig    `yaml:"server"`
	RateLimit RateLimitConfig `yaml:"rateLimit"`
}

// DBConfig holds database connection settings.
// Supported environment variables: DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD
type DBConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

// JWTConfig holds JWT authentication settings.
// Supported environment variables: JWT_SECRET, JWT_EXPIRY
type JWTConfig struct {
	Secret string        `yaml:"secret"` // Secret key for signing JWT tokens
	Expiry time.Duration `yaml:"expiry"` // Token expiration duration
}

// ServerConfig holds HTTP server settings.
// Supported environment variables: SERVER_HOST, SERVER_PORT
type ServerConfig struct {
	Host string `yaml:"host"` // Server listen address
	Port string `yaml:"port"` // Server listen port
}

// RateLimitConfig holds rate limiting settings.
type RateLimitConfig struct {
	IPRequestsPerMinute    int `yaml:"ipRequestsPerMinute"` // Maximum requests per IP per minute
	UsernameRequestsPerMin int `yaml:"usernameRequestsPerMinute"` // Maximum requests per username per window
	WindowMinutes          int `yaml:"windowMinutes"` // Rate limiting window in minutes
}

// Load returns a Config populated from config file and environment variables.
// Config file is loaded first, then environment variables override config values.
func Load() *Config {
	cfg := LoadFromFile("config.yaml")
	overrideFromEnv(cfg)
	return cfg
}

// LoadFromFile loads configuration from a YAML file.
// Returns default config if file doesn't exist.
func LoadFromFile(path string) *Config {
	cfg := defaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return cfg
	}

	return cfg
}

func defaultConfig() *Config {
	return &Config{
		DB: DBConfig{
			Host:     "localhost",
			Port:     "5432",
			Name:     "budget",
			User:     "budget_user",
			Password: "",
		},
		JWT: JWTConfig{
			Secret: "change-me-in-production",
			Expiry: 24 * time.Hour,
		},
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: "8080",
		},
		RateLimit: RateLimitConfig{
			IPRequestsPerMinute:    10,
			UsernameRequestsPerMin: 5,
			WindowMinutes:          15,
		},
	}
}

func overrideFromEnv(cfg *Config) {
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.DB.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		cfg.DB.Port = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.DB.Name = v
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.DB.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.DB.Password = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.JWT.Secret = v
	}
	if v := os.Getenv("JWT_EXPIRY"); v != "" {
		cfg.JWT.Expiry = parseDuration(v, 24*time.Hour)
	}
	if v := os.Getenv("SERVER_HOST"); v != "" {
		cfg.Server.Host = v
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		cfg.Server.Port = v
	}
}

func parseDuration(value string, defaultDuration time.Duration) time.Duration {
	d, err := time.ParseDuration(value)
	if err != nil {
		return defaultDuration
	}
	return d
}
