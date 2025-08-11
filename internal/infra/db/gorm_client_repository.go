package db

import (
	"strings"

	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/domain"
	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/infra/db/model"
	"gorm.io/gorm"
)

type GormClientRepository struct{ db *gorm.DB }

func NewGormClientRepository(db *gorm.DB) *GormClientRepository { return &GormClientRepository{db: db} }

func (r *GormClientRepository) FindByClientID(clientID string) (*domain.Client, error) {
	var m model.Client
	if err := r.db.Where("client_id = ?", clientID).First(&m).Error; err != nil {
		return nil, err
	}
	return &domain.Client{
		ID:              m.ID,
		ClientID:        m.ClientID,
		SecretHash:      m.SecretHash,
		Name:            m.Name,
		AllowedScopes:   splitCSV(m.AllowedScopes),
		AllowedAudience: splitCSV(m.AllowedAudience),
		Active:          m.Active,
	}, nil
}

func splitCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" { return nil }
	parts := strings.Split(s, ",")
	for i := range parts { parts[i] = strings.TrimSpace(parts[i]) }
	return parts
}
