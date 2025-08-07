package domain

import "time"

type TokenService interface {
	GenerateToken(userID string) (string, error)
	ValidateToken(token string) (string, error)
	AccessTokenExpiration(token string) time.Duration
}

