package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
)

// Ajusta limites razoáveis para API. Tudo tunável via env.
func tunePool(sqlDB *sql.DB) {
	maxOpen := getenvInt("DB_MAX_OPEN_CONNS", 20)
	maxIdle := getenvInt("DB_MAX_IDLE_CONNS", 10)
	maxLifetime := getenvDuration("DB_CONN_MAX_LIFETIME", 55*time.Minute)
	maxIdleTime := getenvDuration("DB_CONN_MAX_IDLE_TIME", 10*time.Minute)

	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(maxLifetime)
	sqlDB.SetConnMaxIdleTime(maxIdleTime)

	zap.L().Info("db pool tuned",
		zap.Int("max_open_conns", maxOpen),
		zap.Int("max_idle_conns", maxIdle),
		zap.Duration("max_lifetime", maxLifetime),
		zap.Duration("max_idle_time", maxIdleTime),
	)
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		var out int
		_, _ = fmt.Sscanf(v, "%d", &out)
		if out > 0 {
			return out
		}
	}
	return def
}

func getenvDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
