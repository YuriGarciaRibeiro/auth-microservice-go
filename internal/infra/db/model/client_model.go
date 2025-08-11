package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Client struct {
	ID              string `gorm:"primaryKey"`
	ClientID        string `gorm:"uniqueIndex;not null"`
	SecretHash      string `gorm:"not null"`
	Name            string
	AllowedScopes   string `gorm:"not null;default:''"`
	AllowedAudience string `gorm:"not null;default:''"`
	Active          bool   `gorm:"not null;default:true"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Ensure ID is set when creating (AutoMigrate não cria default DB aqui)
func (c *Client) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	return nil
}
