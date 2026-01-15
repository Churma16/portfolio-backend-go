package util

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// DeleteCacheByPrefix deletes all Redis keys matching the given prefix.
func DeleteCacheByPrefix(redisClient *redis.Client, prefix string) {
	ctx := context.Background()
	// Find all keys matching the prefix
	keys, _ := redisClient.Keys(ctx, prefix+"*").Result()
	if len(keys) > 0 {
		redisClient.Del(ctx, keys...)
	}
}
