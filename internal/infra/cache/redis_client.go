// internal/infra/cache/redis_client.go
package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/YuriGarciaRibeiro/auth-microservice-go/internal/infra/loggger"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisClient(addr, password string, db int) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	if err := redisotel.InstrumentTracing(rdb); err != nil {
		loggger.L().Error("failed to instrument Redis client", zap.Error(err))
	}
	return &RedisClient{client: rdb, ctx: context.Background()}
}

func (r *RedisClient) Set(key string, value interface{}, ttl time.Duration) error {
	return r.client.Set(r.ctx, key, value, ttl).Err()
}

func (r *RedisClient) Get(key string) (string, error) {
	return r.client.Get(r.ctx, key).Result()
}

// ✅ Use estes dois abaixo nos handlers
func (r *RedisClient) SetJSON(key string, v interface{}, ttl time.Duration) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return r.client.Set(r.ctx, key, b, ttl).Err()
}

func (r *RedisClient) GetJSON(key string, out interface{}) error {
	s, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(s), out)
}
