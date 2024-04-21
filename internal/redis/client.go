package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

type Config struct {
	RedisURL string
}

type RedisClientInterface interface {
	HealthCheck() error
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, expire time.Duration) error
	Del(ctx context.Context, key string) error
}

func NewRedisClient(config Config) RedisClientInterface {
	opt, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		panic(err)
	}
	return &RedisClient{
		client: redis.NewClient(opt),
	}
}

func (r *RedisClient) HealthCheck() error {
	ctx := context.Background()
	_, err := r.client.Ping(ctx).Result()
	return err
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisClient) Set(ctx context.Context, key string, value string, expire time.Duration) error {
	return r.client.Set(ctx, key, value, expire).Err()
}

func (r *RedisClient) Del(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
