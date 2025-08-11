package token

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/domain"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/golang-jwt/jwt/v5"
)

// Config holds secrets and TTLs for tokens.
// add (ou ajuste) sua struct Config para incluir estes campos:
type Config struct {
	AccessSecret    []byte
	RefreshSecret   []byte
	AccessTTL       time.Duration
	RefreshTTL      time.Duration
	Issuer          string   // NEW
	DefaultAudience []string // NEW
}


// Service is the concrete TokenService implementation.
type Service struct {
	cfg   Config
	redis *redis.Client
	now   func() time.Time // injectable for tests
}

// NewService builds a new token service.
func NewService(cfg Config, redisClient *redis.Client) *Service {
	return &Service{
		cfg:   cfg,
		redis: redisClient,
		now:   time.Now,
	}
}


func refreshKey(jti string) string  { return "auth:refresh:" + jti }
func blacklistKey(jti string) string { return "auth:blacklist:" + jti }


// IssuePair creates a fresh access+refresh token pair for the given principal.
func (s *Service) IssuePair(p domain.Principal) (domain.TokenPair, error) {
	now := s.now()

	accessJTI := uuid.NewString()
	refreshJTI := uuid.NewString()

	accessExp := now.Add(s.cfg.AccessTTL)
	refreshExp := now.Add(s.cfg.RefreshTTL)

	aud := p.Audience
	if len(aud) == 0 && len(s.cfg.DefaultAudience) > 0 {
		aud = s.cfg.DefaultAudience
	}

	baseClaims := jwt.MapClaims{
		"iss":           s.cfg.Issuer,
		"sub":           p.ID,
		"subject_type":  string(p.Type),
		"email":         p.Email,
		"roles":         p.Roles,
		"scope":         p.Scopes,
		"aud":           aud,                   
		"client_id":     p.ClientID,
	}

	// ----- Access token -----
	accessClaims := jwt.MapClaims{}
	for k, v := range baseClaims {
		accessClaims[k] = v
	}
	accessClaims["jti"] = accessJTI
	accessClaims["iat"] = now.Unix()
	accessClaims["exp"] = accessExp.Unix()

	accessToken, err := s.sign(accessClaims, s.cfg.AccessSecret)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf("sign access: %w", err)
	}

	// ----- Refresh token -----
	refreshClaims := jwt.MapClaims{}
	for k, v := range baseClaims {
		refreshClaims[k] = v
	}
	refreshClaims["jti"] = refreshJTI
	refreshClaims["iat"] = now.Unix()
	refreshClaims["exp"] = refreshExp.Unix()

	refreshToken, err := s.sign(refreshClaims, s.cfg.RefreshSecret)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf("sign refresh: %w", err)
	}

	ctx := context.Background()
	if err := s.saveRefresh(ctx, refreshJTI, p.ID, s.cfg.RefreshTTL); err != nil {
		return domain.TokenPair{}, fmt.Errorf("save refresh: %w", err)
	}

	return domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		AccessExp:    accessExp,
		RefreshExp:   refreshExp,
	}, nil
}

// VerifyAccess validates the access token signature/exp and blacklist.
func (s *Service) VerifyAccess(token string) (*domain.TokenClaims, error) {
	claims, jti, err := s.parseAndValidate(token, s.cfg.AccessSecret)
	if err != nil {
		return nil, err
	}
	// Check blacklist for access token.
	ctx := context.Background()
	revoked, err := s.isAccessRevoked(ctx, jti)
	if err != nil {
		return nil, fmt.Errorf("blacklist check: %w", err)
	}
	if revoked {
		return nil, errors.New("token revoked")
	}
	return claims, nil
}

// Rotate validates the refresh token, checks Redis, then returns a brand-new pair.
func (s *Service) Rotate(refreshToken string) (domain.TokenPair, error) {
	claims, refreshJTI, err := s.parseAndValidate(refreshToken, s.cfg.RefreshSecret)
	if err != nil {
		return domain.TokenPair{}, err
	}

	// Ensure refresh JTI is active in Redis.
	ctx := context.Background()
	ok, err := s.refreshExists(ctx, refreshJTI)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf("check refresh in redis: %w", err)
	}
	if !ok {
		return domain.TokenPair{}, errors.New("refresh not found (revoked or expired)")
	}

	// Build Principal from claims to re-issue a fresh pair.
	p := principalFromClaims(*claims)

	// Issue new pair.
	pair, err := s.IssuePair(p)
	if err != nil {
		return domain.TokenPair{}, err
	}

	// Invalidate the old refresh JTI.
	if err := s.deleteRefresh(ctx, refreshJTI); err != nil {
		return domain.TokenPair{}, fmt.Errorf("delete old refresh: %w", err)
	}

	return pair, nil
}

// RevokePair blacklists the access token and deletes the refresh token entry from Redis.
func (s *Service) RevokePair(accessToken, refreshToken string) error {
	ctx := context.Background()

	// Parse both tokens (do not strictly require success on both to avoid leaks).
	_, accessJTI, _ := s.parseAndValidate(accessToken, s.cfg.AccessSecret)
	_, refreshJTI, _ := s.parseAndValidate(refreshToken, s.cfg.RefreshSecret)

	// Blacklist access if we extracted a JTI.
	if accessJTI != "" {
		// Set a TTL equal to the remaining exp to avoid indefinite growth.
		ttl, _ := s.remainingTTL(accessToken, s.cfg.AccessSecret)
		if ttl <= 0 {
			ttl = s.cfg.AccessTTL // fallback
		}
		if err := s.blacklistAccess(ctx, accessJTI, ttl); err != nil {
			return fmt.Errorf("blacklist access: %w", err)
		}
	}

	// Delete refresh JTI entry (idempotent).
	if refreshJTI != "" {
		if err := s.deleteRefresh(ctx, refreshJTI); err != nil {
			return fmt.Errorf("delete refresh: %w", err)
		}
	}

	return nil
}

// Introspect returns (active, claims) where active=false means invalid/expired/revoked.
func (s *Service) Introspect(token string) (bool, *domain.TokenClaims, error) {
	claims, jti, err := s.parseAndValidate(token, s.cfg.AccessSecret)
	if err != nil {
		// Could be expired/invalid signature, we treat as inactive.
		return false, nil, nil
	}
	// Check blacklist.
	ctx := context.Background()
	revoked, err := s.isAccessRevoked(ctx, jti)
	if err != nil {
		return false, nil, fmt.Errorf("blacklist check: %w", err)
	}
	if revoked {
		return false, nil, nil
	}
	return true, claims, nil
}

// IssueAccessOnly generates an access token without a refresh token.
func (s *Service) IssueAccessOnly(p domain.Principal) (token string, exp time.Time, err error) {
	now := s.now()
	jti := uuid.NewString()
	exp = now.Add(s.cfg.AccessTTL)

	aud := p.Audience
	if len(aud) == 0 && len(s.cfg.DefaultAudience) > 0 {
		aud = s.cfg.DefaultAudience
	}

	claims := jwt.MapClaims{
		"iss":          s.cfg.Issuer,
		"sub":          p.ID,
		"subject_type": string(p.Type), // "service"
		"email":        p.Email,        // vazio para service
		"roles":        p.Roles,
		"scope":        p.Scopes,
		"aud":          aud,
		"client_id":    p.ClientID,
		"jti":          jti,
		"iat":          now.Unix(),
		"exp":          exp.Unix(),
	}

	tok, err := s.sign(claims, s.cfg.AccessSecret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access: %w", err)
	}
	return tok, exp, nil
}

// ===== Internals =====

// sign signs a MapClaims with the given secret using HS256.
func (s *Service) sign(claims jwt.MapClaims, secret []byte) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(secret)
}

// parseAndValidate parses and validates signature + exp. Returns domain claims and the JTI.
func (s *Service) parseAndValidate(tokenStr string, secret []byte) (*domain.TokenClaims, string, error) {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
		jwt.WithAudience(), // allow aud as array
	)
	var mc jwt.MapClaims
	tok, err := parser.ParseWithClaims(tokenStr, &mc, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil || !tok.Valid {
		return nil, "", errors.New("invalid token")
	}


	jti, _ := mc["jti"].(string)
	claims := claimsFromMap(mc)
	return &claims, jti, nil
}

// remainingTTL estimates remaining validity of a token by its exp claim.
func (s *Service) remainingTTL(tokenStr string, secret []byte) (time.Duration, error) {
	var mc jwt.MapClaims
	_, err := jwt.ParseWithClaims(tokenStr, &mc, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return 0, err
	}
	expFloat, ok := mc["exp"].(float64)
	if !ok {
		return 0, errors.New("exp not found")
	}
	exp := time.Unix(int64(expFloat), 0)
	ttl := time.Until(exp)
	if ttl < 0 {
		return 0, nil
	}
	return ttl, nil
}

// claimsFromMap converts jwt.MapClaims into domain.TokenClaims.
func claimsFromMap(mc jwt.MapClaims) domain.TokenClaims {
	roles := toStringSlice(mc["roles"])
	scopes := toStringSlice(mc["scope"])
	aud := toStringSlice(mc["aud"])
	email, _ := mc["email"].(string)
	clientID, _ := mc["client_id"].(string)
	sub, _ := mc["sub"].(string)

	var st domain.PrincipalType = "user"
	if stStr, ok := mc["subject_type"].(string); ok && stStr != "" {
		st = domain.PrincipalType(stStr)
	}

	return domain.TokenClaims{
		SubjectType: st,
		SubjectID:   sub,
		Email:       email,
		Roles:       roles,
		Scopes:      scopes,
		ClientID:    clientID,
		Audience:    aud,
	}
}

// principalFromClaims reconstructs a Principal from TokenClaims.
func principalFromClaims(c domain.TokenClaims) domain.Principal {
	return domain.Principal{
		Type:     c.SubjectType,
		ID:       c.SubjectID,
		Email:    c.Email,
		Roles:    c.Roles,
		Scopes:   c.Scopes,
		ClientID: c.ClientID,
		Audience: c.Audience,
	}
}

func toStringSlice(v any) []string {
	switch t := v.(type) {
	case []string:
		return t
	case []any:
		out := make([]string, 0, len(t))
		for _, it := range t {
			if s, ok := it.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case string:
		// single string → one-element slice
		return []string{t}
	default:
		return nil
	}
}

// ===== Redis helpers =====

func (s *Service) saveRefresh(ctx context.Context, jti string, userID string, ttl time.Duration) error {
	// Value could be userID or any marker; we only check existence later.
	return s.redis.Set(ctx, refreshKey(jti), userID, ttl).Err()
}

func (s *Service) deleteRefresh(ctx context.Context, jti string) error {
	return s.redis.Del(ctx, refreshKey(jti)).Err()
}

func (s *Service) refreshExists(ctx context.Context, jti string) (bool, error) {
	n, err := s.redis.Exists(ctx, refreshKey(jti)).Result()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}

func (s *Service) blacklistAccess(ctx context.Context, jti string, ttl time.Duration) error {
	return s.redis.Set(ctx, blacklistKey(jti), "1", ttl).Err()
}

func (s *Service) isAccessRevoked(ctx context.Context, jti string) (bool, error) {
	n, err := s.redis.Exists(ctx, blacklistKey(jti)).Result()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}
