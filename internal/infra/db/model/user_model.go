package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID       string `gorm:"type:uuid;primaryKey"`
	Email    string `gorm:"type:varchar(180);uniqueIndex;not null"`
	Password string `gorm:"type:varchar(255);not null"`
	Verified bool   `gorm:"not null;default:false"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.NewString()
	}
	return nil
}
