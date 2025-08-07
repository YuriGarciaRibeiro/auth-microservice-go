package domain

import (
	"time"

	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/service/token"
)

type TokenService interface {
	GenerateToken(userID, email string) (string, error)
	ValidateToken(token string) (*token.CustomClaims, error)
	AccessTokenExpiration(token string) time.Duration
	GetTokenIdentifier(tokenStr string) (string, error)
}


