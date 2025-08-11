package db

import (
	"context"
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
	); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// Seed default client (idempotente)
	if err := SeedClients(context.Background(), db); err != nil {
		log.Fatalf("failed to seed clients: %v", err)
	}

	return db
}
