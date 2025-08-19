package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Cache    CacheConfig
	Log      LogConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	AccessSecret    string
	RefreshSecret   string
	AccessTTL       time.Duration
	RefreshTTL      time.Duration
	Issuer          string
	DefaultAudience []string
}

type CacheConfig struct {
	ProfileTTL    time.Duration
	PermissionTTL time.Duration
}

type LogConfig struct {
	Level    string
	Encoding string
	AppEnv   string
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: getenv("SERVER_PORT", "8080"),
			Host: getenv("SERVER_HOST", ""),
		},
		Database: DatabaseConfig{
			Host:     getenv("DB_HOST", "localhost"),
			Port:     getenv("DB_PORT", "5432"),
			User:     getenv("DB_USER", "user"),
			Password: getenv("DB_PASSWORD", ""),
			Name:     getenv("DB_NAME", "auth_db"),
		},
		Redis: RedisConfig{
			Addr:     getenv("REDIS_ADDR", "localhost:6379"),
			Password: getenv("REDIS_PASS", ""),
			DB:       getenvInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			AccessSecret:    getenv("ACCESS_SECRET", ""),
			RefreshSecret:   getenv("REFRESH_SECRET", ""),
			AccessTTL:       getenvDuration("ACCESS_TOKEN_TTL", "15m"),
			RefreshTTL:      getenvDuration("REFRESH_TOKEN_TTL", "168h"),
			Issuer:          getenv("JWT_ISSUER", "auth-microservice"),
			DefaultAudience: splitCSV(getenv("JWT_AUDIENCE", "")),
		},
		Cache: CacheConfig{
			ProfileTTL:    getenvDuration("CACHE_PROFILE_TTL", "5m"),
			PermissionTTL: getenvDuration("PERM_CACHE_TTL", "15m"),
		},
		Log: LogConfig{
			Level:    getenv("LOG_LEVEL", "info"),
			Encoding: getenv("LOG_ENCODING", "json"),
			AppEnv:   getenv("APP_ENV", "dev"),
		},
	}

	// Validate required fields
	if cfg.JWT.AccessSecret == "" {
		return nil, fmt.Errorf("ACCESS_SECRET is required")
	}
	if cfg.JWT.RefreshSecret == "" {
		return nil, fmt.Errorf("REFRESH_SECRET is required")
	}
	if cfg.Database.Password == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required")
	}

	return cfg, nil
}

func getenv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getenvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func getenvDuration(key, defaultValue string) time.Duration {
	value := getenv(key, defaultValue)
	if d, err := time.ParseDuration(value); err == nil {
		return d
	}
	// Return default if parsing fails
	if d, err := time.ParseDuration(defaultValue); err == nil {
		return d
	}
	return 0
}

func splitCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
