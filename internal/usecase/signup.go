package usecase

import (
	"errors"

	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/domain"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type SignupUseCase struct {
	UserRepo domain.UserRepository
}

func NewSignupUseCase(userRepo domain.UserRepository) *SignupUseCase {
	return &SignupUseCase{
		UserRepo: userRepo,
	}
}
	
func (uc *SignupUseCase) Execute(email, password string) (*domain.User, error) {
	existingUser, _ := uc.UserRepo.FindByEmail(email)
	if existingUser != nil {
		return nil, errors.New("Email already in use")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("error generating password hash")
	}

	newUser := &domain.User{
		ID:       generateID(),
		Email:    email,
		Password: string(hashedPassword),
		Verified: false,
	}

	return newUser, uc.UserRepo.Create(newUser)
}

func generateID() string {
	return uuid.NewString()
}

