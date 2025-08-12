package db

import (
	"log"
	"os"

	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/infra/db/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SeedInitialData(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		scopes := []model.Scope{
			{Key: "read:users", Description: "Read user information"},
			{Key: "write:users", Description: "Modify user information"},
			{Key: "admin:all", Description: "Full admin access"},
		}
		for _, s := range scopes {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "key"}},
				DoNothing: true,
			}).Create(&s).Error; err != nil {
				return err
			}
		}

		roles := []model.Role{
			{Key: "admin", Description: "Administrator with full permissions"},
			{Key: "user", Description: "Regular user"},
		}
		for _, r := range roles {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "key"}},
				DoNothing: true,
			}).Create(&r).Error; err != nil {
				return err
			}
		}

		var adminRole model.Role
		if err := tx.Where("key = ?", "admin").First(&adminRole).Error; err != nil {
			return err
		}
		var allScopes []model.Scope
		if err := tx.Find(&allScopes).Error; err != nil {
			return err
		}
		for _, s := range allScopes {
			link := model.RoleScope{RoleID: adminRole.ID, ScopeID: s.ID}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "role_id"}, {Name: "scope_id"}},
				DoNothing: true,
			}).Create(&link).Error; err != nil {
				return err
			}
		}

		adminEmail := getenv("ADMIN_EMAIL", "admin@example.com")
		adminPlain := getenv("ADMIN_PASSWORD", "admin123")
		hash, err := bcrypt.GenerateFromPassword([]byte(adminPlain), 12)
		if err != nil {
			return err
		}

		var adminUser model.User
		if err := tx.Where("email = ?", adminEmail).First(&adminUser).Error; err != nil {
			adminUser = model.User{
				Email:    adminEmail,
				Password: string(hash),
			}
			if err := tx.Create(&adminUser).Error; err != nil {
				return err
			}
		} else {
			if os.Getenv("RESET_ADMIN_PASSWORD_ON_SEED") == "1" || adminUser.Password == "" {
				if err := tx.Model(&adminUser).Update("password", string(hash)).Error; err != nil {
					return err
				}
			}
		}

		ur := model.UserRole{UserID: adminUser.ID, RoleID: adminRole.ID}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "role_id"}},
			DoNothing: true,
		}).Create(&ur).Error; err != nil {
			return err
		}
		return nil
	})
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
