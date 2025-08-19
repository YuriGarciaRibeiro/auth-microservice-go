package db

import (
	gormotel "gorm.io/plugin/opentelemetry/tracing"
	"gorm.io/gorm"

	"go.uber.org/zap"
)

func enableTracing(db *gorm.DB) {
	if err := db.Use(gormotel.NewPlugin()); err != nil {
		zap.L().Warn("failed to enable gorm otel plugin", zap.Error(err))
	} else {
		zap.L().Info("gorm otel plugin enabled")
	}
}
