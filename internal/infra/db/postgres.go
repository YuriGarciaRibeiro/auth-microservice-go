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

	// Auto migrate user model
	if err := db.AutoMigrate(&model.User{}); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	return db
}
