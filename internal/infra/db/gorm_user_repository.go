package db

import (
	"errors"

	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/domain"
	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/infra/db/model"
	"gorm.io/gorm"
)

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) Create(user *domain.User) error {
	return r.db.Create(fromDomainUser(user)).Error
}

func (r *GormUserRepository) FindByEmail(email string) (*domain.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return toDomainUser(&user), err
}

func toDomainUser(m *model.User) *domain.User {
	if m == nil {
		return nil
	}
	return &domain.User{
		ID:       m.ID,
		Email:    m.Email,
		Password: m.Password,
		Verified: m.Verified,
	}
}

func fromDomainUser(u *domain.User) *model.User {
	if u == nil {
		return nil
	}
	return &model.User{
		ID:       u.ID,
		Email:    u.Email,
		Password: u.Password,
		Verified: u.Verified,
	}
}

func (r *GormUserRepository) GetAll() ([]*domain.User, error) {
	var users []model.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}

	domainUsers := make([]*domain.User, len(users))
	for i, user := range users {
		domainUsers[i] = toDomainUser(&user)
	}
	return domainUsers, nil
}

