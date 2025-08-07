package token

import (
    "time"
    "github.com/golang-jwt/jwt/v5"
)



type TokenService struct {
	secretKey               string
	issuer                  string
	accessTokenExpiration    time.Duration
	refreshTokenExpiration    time.Duration
}

func NewTokenService(secretKey, issuer string, accessTokenExpiration, refreshTokenExpiration time.Duration) *TokenService {
	return &TokenService{
		secretKey:              secretKey,
		issuer:                  issuer,
		accessTokenExpiration:   accessTokenExpiration,
		refreshTokenExpiration:  refreshTokenExpiration,
	}
}

func (s *TokenService) GenerateToken(userID string) (string, error) {
	claims := &jwt.RegisteredClaims{
		Issuer:    s.issuer,
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTokenExpiration)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secretKey))
}

func (s *TokenService) ValidateToken(token string) (string, error) {
	claims := &jwt.RegisteredClaims{}
	tkn, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.secretKey), nil
	})
	if err != nil || !tkn.Valid {
		return "", err
	}
	return claims.Subject, nil
}

func (s *TokenService) AccessTokenExpiration(token string) time.Duration {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.secretKey), nil
	})
	if err != nil {
		return 0
	}
	return s.accessTokenExpiration
}