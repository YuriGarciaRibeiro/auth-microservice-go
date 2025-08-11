package db

import (
	"context"

	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/infra/db/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedClients(ctx context.Context, gdb *gorm.DB) error {
	const (
		clientID    = "service-a"
		secretPlain = "service-a-secret"
		name        = "Service A"
		scopes      = "read:users write:users" // pode ser "read:users,write:users" também
		audience    = "service-b"
	)

	var exists int64
	if err := gdb.WithContext(ctx).
		Model(&model.Client{}).
		Where("client_id = ?", clientID).
		Count(&exists).Error; err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(secretPlain), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return gdb.WithContext(ctx).Create(&model.Client{
		ClientID:        clientID,
		SecretHash:      string(hash),
		Name:            name,
		AllowedScopes:   scopes,
		AllowedAudience: audience,
		Active:          true,
	}).Error
}
