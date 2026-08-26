package util

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// DeleteCacheByPrefix deletes all Redis keys matching the given prefix and invalidates Express cache.
func DeleteCacheByPrefix(redisClient *redis.Client, prefix string) {
	if redisClient == nil {
		return
	}
	ctx := context.Background()
	// Find all keys matching the prefix
	keys, _ := redisClient.Keys(ctx, prefix+"*").Result()
	if len(keys) > 0 {
		redisClient.Del(ctx, keys...)
	}

	// Trigger cache invalidation in Express API asynchronously to avoid blocking the response
	go InvalidateExpressCache(prefix)
}

// InvalidateExpressCache sends an HTTP POST request to Express API to invalidate its cache.
func InvalidateExpressCache(prefix string) {
	expressURL := os.Getenv("EXPRESS_API_URL")
	if expressURL == "" {
		expressURL = "http://localhost:4000" // Default fallback
	}

	token := os.Getenv("TOKEN_SYMMETRIC_KEY")

	url := fmt.Sprintf("%s/api/internal/cache/invalidate", expressURL)
	payload := map[string]string{"prefix": prefix}
	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshalling cache invalidation payload: %v\n", err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error creating cache invalidation request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Token", token)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending cache invalidation request to Express: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Express cache invalidation failed. Status: %s\n", resp.Status)
	} else {
		fmt.Printf("Express cache successfully invalidated for prefix: %s\n", prefix)
	}
}

