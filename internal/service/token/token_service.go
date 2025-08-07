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

type CustomClaims struct {
	jwt.RegisteredClaims
	ID    string `json:"id"`
	Email string `json:"email"`
}

func NewTokenService(secretKey, issuer string, accessTokenExpiration, refreshTokenExpiration time.Duration) *TokenService {
	return &TokenService{
		secretKey:              secretKey,
		issuer:                  issuer,
		accessTokenExpiration:   accessTokenExpiration,
		refreshTokenExpiration:  refreshTokenExpiration,
	}
}

func (s *TokenService) GenerateToken(userID, email string) (string, error) {
	claims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		ID:    userID,
		Email: email,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secretKey))
}

func (s *TokenService) ValidateToken(tokenStr string) (*CustomClaims, error) {
	claims := &CustomClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.secretKey), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return claims, nil
}


func (s *TokenService) AccessTokenExpiration(tokenStr string) time.Duration {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.secretKey), nil
	})
	if err != nil || !token.Valid || claims.ExpiresAt == nil {
		return 0
	}

	return time.Until(claims.ExpiresAt.Time)
}

func (s *TokenService) GetTokenIdentifier(tokenStr string) (string, error) {
	claims, err := s.ValidateToken(tokenStr)
	if err != nil {
		return "", err
	}
	return claims.ID, nil
}
