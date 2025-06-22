package ogtags_cache

import (
	"context"
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func Test_Set(t *testing.T) {
	t.Run("successful set value and ttl", func(t *testing.T) {
		redisServer := setup()
		defer redisServer.Close()

		redisClient := redis.NewClient(&redis.Options{
			Addr: redisServer.Addr(),
		})

		url := "test url"
		jsonBytes := []byte("test json")
		k := createKey(url)

		cache := New(redisClient)
		err := cache.Set(url, jsonBytes)
		assert.Nil(t, err)

		got, err := redisClient.Get(context.TODO(), k).Result()
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, string(jsonBytes), got)
		assert.Equal(t, redisServer.TTL(k), ttl)
	})

	t.Run("set failed", func(t *testing.T) {
		redisServer := setup()
		redisClient := redis.NewClient(&redis.Options{
			Addr: redisServer.Addr(),
		})

		redisServer.Close() // Close the connection to force an error

		url := "test url"
		jsonBytes := []byte("test json")

		cache := New(redisClient)
		err := cache.Set(url, jsonBytes)
		assert.Contains(t, err.Error(), "Set:redisClient.Set:")
	})

}

func Test_Get(t *testing.T) {
	t.Run("get added key", func(t *testing.T) {
		redisServer := setup()
		defer redisServer.Close()

		redisClient := redis.NewClient(&redis.Options{
			Addr: redisServer.Addr(),
		})

		url := "test url"
		jsonBytes := []byte("test json")
		k := createKey(url)

		err := redisClient.Set(context.TODO(), k, jsonBytes, ttl).Err()
		if err != nil {
			t.Fatal(err)
		}

		cache := New(redisClient)
		got, err := cache.Get(url)
		assert.Nil(t, err)
		assert.Equal(t, string(jsonBytes), got)
	})

	t.Run("get key not found", func(t *testing.T) {
		redisServer := setup()
		defer redisServer.Close()

		redisClient := redis.NewClient(&redis.Options{
			Addr: redisServer.Addr(),
		})

		url := "non-existent-url"
		cache := New(redisClient)

		result, err := cache.Get(url)

		assert.Empty(t, result)
		assert.True(t, errors.Is(err, ErrKeyNotFound))
	})

	t.Run("get failed", func(t *testing.T) {
		redisServer := setup()
		defer redisServer.Close()

		redisClient := redis.NewClient(&redis.Options{
			Addr: redisServer.Addr(),
		})

		url := "test url"
		cache := New(redisClient)

		// Close the client connection to force an error
		redisClient.Close()

		result, err := cache.Get(url)

		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "Get:redisClient.Get:")

		// it's not the "key not found" error
		assert.False(t, errors.Is(err, ErrKeyNotFound))
	})
}

func Test_SetGet(t *testing.T) {
	t.Run("Set then Get", func(t *testing.T) {
		redisServer := setup()
		defer redisServer.Close()

		redisClient := redis.NewClient(&redis.Options{
			Addr: redisServer.Addr(),
		})

		url := "test url"
		jsonBytes := []byte("test json")

		cache := New(redisClient)

		err := cache.Set(url, jsonBytes)
		assert.Nil(t, err)

		jsonStr, err := cache.Get(url)
		assert.Nil(t, err)

		assert.Equal(t, string(jsonBytes), jsonStr)
	})
}

func setup() *miniredis.Miniredis {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	return s
}
