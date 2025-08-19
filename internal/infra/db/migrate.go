package db

import (
	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/infra/db/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func runMigrations(db *gorm.DB) {
	if err := db.AutoMigrate(allModels()...); err != nil {
		zap.L().Fatal("failed to run migrations", zap.Error(err))
	}
	if err := SeedInitialData(db); err != nil {
		zap.L().Fatal("failed to seed initial data", zap.Error(err))
	}
}

// (Opcional) Centralize os models aqui se preferir.
func allModels() []any {
	return []any{
		&model.User{},
		&model.Client{},
		&model.Scope{},
		&model.Role{},
		&model.RoleScope{},
		&model.UserRole{},
		&model.ClientScope{},
		&model.UserScope{},
	}
}
