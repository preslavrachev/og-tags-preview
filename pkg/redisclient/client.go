package redisclient

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	ctxTimeoutDuration = 4 * time.Second
)

type RedisConfig struct {
	Address  string
	Password string
	DB       int
}

func New(cfg RedisConfig) *redis.Client {
	rc := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	// check redis connection
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeoutDuration)
	defer cancel()
	if err := rc.Ping(ctx).Err(); err != nil {
		slog.Error("could not connect to Redis")
		os.Exit(1)
	}
	return rc
}
