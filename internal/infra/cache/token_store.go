package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenStore handles token-related keys on Redis (refresh and blacklist).
type TokenStore struct {
	rdb *redis.Client
}

func NewTokenStore(rdb *redis.Client) *TokenStore {
	return &TokenStore{rdb: rdb}
}

// Key patterns. Keep them centralized to avoid typos.
func refreshKey(jti string) string     { return fmt.Sprintf("auth:refresh:%s", jti) }
func blacklistKey(jti string) string   { return fmt.Sprintf("auth:blacklist:%s", jti) }
func userRefreshSet(userID string) string { return fmt.Sprintf("auth:user:%s:refresh", userID) } // optional

// SetRefresh stores a refresh token JTI with TTL, value can be userID or any metadata you need.
func (s *TokenStore) SetRefresh(ctx context.Context, jti string, userID string, ttl time.Duration) error {
	key := refreshKey(jti)
	if err := s.rdb.Set(ctx, key, userID, ttl).Err(); err != nil {
		return err
	}
	// Optional: track all refresh JTIs per user to revoke them all later
	_ = s.rdb.SAdd(ctx, userRefreshSet(userID), jti).Err()
	_ = s.rdb.Expire(ctx, userRefreshSet(userID), ttl).Err()
	return nil
}

// DeleteRefresh removes a refresh token JTI (invalidate the token).
func (s *TokenStore) DeleteRefresh(ctx context.Context, jti string, userID string) error {
	if err := s.rdb.Del(ctx, refreshKey(jti)).Err(); err != nil {
		return err
	}
	// Optional: keep the set tidy
	if userID != "" {
		_ = s.rdb.SRem(ctx, userRefreshSet(userID), jti).Err()
	}
	return nil
}

// IsAccessRevoked checks if access token JTI is blacklisted.
func (s *TokenStore) IsAccessRevoked(ctx context.Context, jti string) (bool, error) {
	key := blacklistKey(jti)
	exists, err := s.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

// BlacklistAccess marks an access token JTI as revoked until its natural expiration.
func (s *TokenStore) BlacklistAccess(ctx context.Context, jti string, ttl time.Duration) error {
	return s.rdb.Set(ctx, blacklistKey(jti), "1", ttl).Err()
}

// RevokeAllUserRefresh invalidates all refresh tokens for a user (optional bulk).
func (s *TokenStore) RevokeAllUserRefresh(ctx context.Context, userID string) error {
	setKey := userRefreshSet(userID)
	jtis, err := s.rdb.SMembers(ctx, setKey).Result()
	if err != nil {
		return err
	}
	if len(jtis) == 0 {
		return nil
	}
	// Delete all refresh keys
	keys := make([]string, 0, len(jtis))
	for _, j := range jtis {
		keys = append(keys, refreshKey(j))
	}
	if len(keys) > 0 {
		if err := s.rdb.Del(ctx, keys...).Err(); err != nil {
			return err
		}
	}
	// Clear the set
	if err := s.rdb.Del(ctx, setKey).Err(); err != nil {
		return err
	}
	return nil
}
