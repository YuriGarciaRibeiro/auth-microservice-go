package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Scope struct {
	ID          string    `gorm:"type:uuid"`
	Key         string    `gorm:"type:varchar(100);uniqueIndex"`
	Description string    `gorm:"type:varchar(255)"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (s *Scope) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" { s.ID = uuid.NewString() }
	return nil
}

type Role struct {
	ID          string    `gorm:"type:uuid"`
	Key         string    `gorm:"type:varchar(100);uniqueIndex"`
	Description string    `gorm:"type:varchar(255)"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" { r.ID = uuid.NewString() }
	return nil
}

type RoleScope struct {
	RoleID  string `gorm:"type:uuid;primaryKey"`
	ScopeID string `gorm:"type:uuid;primaryKey"`
}

type UserRole struct {
	UserID string `gorm:"type:uuid;primaryKey"`
	RoleID string `gorm:"type:uuid;primaryKey"`
}

type ClientScope struct {
	ClientID string `gorm:"type:uuid;primaryKey"`
	ScopeID  string `gorm:"type:uuid;primaryKey"`
}

type UserScope struct {
	UserID    string     `gorm:"type:uuid;primaryKey;index"`
	ScopeID   string     `gorm:"type:uuid;primaryKey;index"`
	ExpiresAt *time.Time
	GrantedBy string
	CreatedAt time.Time
	UpdatedAt time.Time
}
