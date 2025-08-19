package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ConnectPostgres abre conexão GORM com Postgres, ajusta pool,
// habilita tracing do GORM, roda migrations + seed e faz ping.
func ConnectPostgres() *gorm.DB {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = defaultDSN()
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig())
	if err != nil {
		zap.L().Fatal("failed to connect to database", zap.Error(err))
	}

	// sql.DB cru para pool e ping 
	sqlDB, err := db.DB()
	if err != nil {
		zap.L().Fatal("failed to get sql.DB from gorm", zap.Error(err))
	}

	// Pool + tracing do GORM
	tunePool(sqlDB)
	enableTracing(db)

	// Ping de sanidade com timeout curto
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		zap.L().Fatal("failed to ping database", zap.Error(err))
	}

	// Migrações e seed
	runMigrations(db)

	zap.L().Info("postgres connected and migrated")
	return db
}

func gormConfig() *gorm.Config {
	return &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}
}

func defaultDSN() string {
	host := getenv("PGHOST", "localhost")
	port := getenv("PGPORT", "5432")
	user := getenv("PGUSER", "user")
	pass := getenv("PGPASSWORD", "pass")
	name := getenv("PGDATABASE", "auth_db")
	ssl := getenv("PGSSLMODE", "disable")
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		host, port, user, pass, name, ssl,
	)
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
