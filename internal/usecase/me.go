package usecase

import (
	"errors"

	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/domain"
)

type MeUseCase struct {
	UserRepo domain.UserRepository
}

func NewMeUseCase(userRepo domain.UserRepository) *MeUseCase {
	return &MeUseCase{
		UserRepo: userRepo,
	}
}

func (uc *MeUseCase) Execute(email string) (*domain.User, error) {
	existingUser, _ := uc.UserRepo.FindByEmail(email)
	if existingUser != nil {
		return nil, errors.New("Email already in use")
	}

	newUser := &domain.User{
		ID:       generateID(),
		Email:    email,
		Password: "", // Password will be set later
		Verified: false,
	}

	return newUser, uc.UserRepo.Create(newUser)
}