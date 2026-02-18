package database

import (
	"context"
	"time"

	"github.com/codepnw/go-starter-kit/internal/config"
	"github.com/redis/go-redis/v9"
)

func InitRedis(cfg *config.EnvConfig) (*redis.Client, error) {
	options := &redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       0, // Default DB
	}
	rdb := redis.NewClient(options)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return rdb, nil
}
