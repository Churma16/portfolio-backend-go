package util

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// DeleteCacheByPrefix deletes all Redis keys matching the given prefix.
func DeleteCacheByPrefix(redisClient *redis.Client, prefix string) {
	ctx := context.Background()
	// Find all keys matching the prefix
	keysSingleData, _ := redisClient.Keys(ctx, prefix+"list:*").Result()
	keysMultiData, _ := redisClient.Keys(ctx, prefix+"single:*").Result()
	if len(keysSingleData) > 0 {
		redisClient.Del(ctx, keysSingleData...)
	}
	if len(keysMultiData) > 0 {
		redisClient.Del(ctx, keysMultiData...)
	}
}
