package usecase

import (
	"errors"

	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type LoginUseCase struct {
	UserRepo domain.UserRepository
}

func NewLoginUseCase(userRepo domain.UserRepository) *LoginUseCase {
	return &LoginUseCase{
		UserRepo: userRepo,
	}
}

func (uc *LoginUseCase) Execute(email string, password string) (string, error) {
	user, err := uc.UserRepo.FindByEmail(email)
	if err != nil {
		return "", errors.New("error finding user")
	}
	if user == nil {
		return "", errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid password")
	}

	return user.ID, nil
}