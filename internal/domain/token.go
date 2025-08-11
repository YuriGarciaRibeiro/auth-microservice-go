package domain

import "time"

// TokenPair wraps both access and refresh tokens.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	AccessExp    time.Time
	RefreshExp   time.Time
}

// TokenClaims are the normalized claims used across the app.
type TokenClaims struct {
	SubjectType PrincipalType
	SubjectID   string
	Email       string
	Roles       []string
	Scopes      []string
	ClientID    string
	Audience    []string

	// Standard JWT claims we often need to access explicitly.
	ID        string    // jti
	IssuedAt  time.Time // iat
	ExpiresAt time.Time // exp
	Issuer    string    // iss (optional)
}

// TokenService defines the auth core behaviors.
type TokenService interface {
	// IssuePair generates a new access+refresh pair for a given principal.
	IssuePair(p Principal) (TokenPair, error)

	// VerifyAccess validates an access token (signature/exp/blacklist) and returns claims.
	VerifyAccess(accessToken string) (*TokenClaims, error)

	// Rotate validates a refresh token against Redis and returns a new pair.
	Rotate(refreshToken string) (TokenPair, error)

	// RevokePair blacklists the access token and invalidates the refresh token (Redis).
	RevokePair(accessToken, refreshToken string) error

	// Introspect tells whether the token is active and returns claims if active.
	Introspect(token string) (active bool, claims *TokenClaims, err error)
}
