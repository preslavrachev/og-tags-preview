package ogtags_cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	ctxTimeoutDuration = 4 * time.Second // timeout duration for each request
	ttl                = time.Hour * 1   // expire in 1h
	sessionKeyPrefix   = "ogtag"         // help namespace keys
)

type OGCacheClient interface {
	Set(url string, jsonByte []byte) error
	Get(url string) (string, error)
}

var (
	ErrKeyNotFound = errors.New("key not found or expired")
)

type OGCache struct {
	rc *redis.Client
}

func New(rc *redis.Client) *OGCache {
	return &OGCache{
		rc: rc,
	}
}

// cached og tags of a url
func (c *OGCache) Set(url string, jsonByte []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeoutDuration)
	defer cancel()
	k := createKey(url)
	err := c.rc.Set(ctx, k, jsonByte, ttl).Err()
	if err != nil {
		return fmt.Errorf("Set:redisClient.Set: %w", err)
	}
	slog.Info("cached og tags", "url", url)
	return nil
}

// check for cached url
func (c *OGCache) Get(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeoutDuration)
	defer cancel()
	k := createKey(url)
	jsonStr, err := c.rc.Get(ctx, k).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrKeyNotFound
		}
		return "", fmt.Errorf("Get:redisClient.Get: %w", err)
	}
	slog.Info("found cached url", "url", url)
	return jsonStr, nil
}

func createKey(url string) string {
	hash := sha256.Sum256([]byte(url))
	return fmt.Sprintf("%s:%s", sessionKeyPrefix, hex.EncodeToString(hash[:]))
}
