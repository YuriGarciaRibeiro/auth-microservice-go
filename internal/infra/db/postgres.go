package db

import (
	"log"
	"os"

	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/infra/db/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectPostgres() *gorm.DB {
	dsn := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.Client{},
		&model.Scope{},
		&model.Role{},
		&model.RoleScope{},
		&model.UserRole{},
		&model.ClientScope{},
		&model.UserScope{},
	); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	if err := SeedInitialData(db); err != nil {
		log.Fatalf("failed to seed initial data: %v", err)
	}

	return db
}
