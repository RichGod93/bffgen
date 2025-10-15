package aggregation

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheManager provides caching with Redis backend and in-memory fallback
type CacheManager struct {
	redisClient *redis.Client
	memCache    *sync.Map
	ttl         time.Duration
	prefix      string
	useRedis    bool
}

// CacheItem represents a cached item with expiration
type CacheItem struct {
	Data      []byte
	ExpiresAt time.Time
}

// NewCacheManager creates a new cache manager
func NewCacheManager(redisURL string, ttl time.Duration, prefix string) *CacheManager {
	cm := &CacheManager{
		memCache: &sync.Map{},
		ttl:      ttl,
		prefix:   prefix,
	}

	// Try to connect to Redis
	if redisURL != "" {
		opts, err := redis.ParseURL(redisURL)
		if err == nil {
			client := redis.NewClient(opts)
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			// Test connection
			if err := client.Ping(ctx).Err(); err == nil {
				cm.redisClient = client
				cm.useRedis = true
			}
		}
	}

	return cm
}

// Get retrieves a value from cache
func (cm *CacheManager) Get(ctx context.Context, key string) (interface{}, error) {
	fullKey := cm.prefix + key

	if cm.useRedis {
		// Try Redis first
		val, err := cm.redisClient.Get(ctx, fullKey).Bytes()
		if err == nil {
			var data interface{}
			if err := json.Unmarshal(val, &data); err == nil {
				return data, nil
			}
		}
	}

	// Fallback to in-memory cache
	if val, ok := cm.memCache.Load(fullKey); ok {
		item := val.(CacheItem)
		if time.Now().Before(item.ExpiresAt) {
			var data interface{}
			if err := json.Unmarshal(item.Data, &data); err == nil {
				return data, nil
			}
		} else {
			// Expired, delete
			cm.memCache.Delete(fullKey)
		}
	}

	return nil, fmt.Errorf("cache miss")
}

// Set stores a value in cache
func (cm *CacheManager) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	fullKey := cm.prefix + key

	// Use provided TTL or default
	if ttl == 0 {
		ttl = cm.ttl
	}

	// Serialize value
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if cm.useRedis {
		// Store in Redis
		if err := cm.redisClient.Set(ctx, fullKey, data, ttl).Err(); err == nil {
			return nil
		}
		// If Redis fails, fall through to memory cache
	}

	// Store in memory cache
	cm.memCache.Store(fullKey, CacheItem{
		Data:      data,
		ExpiresAt: time.Now().Add(ttl),
	})

	return nil
}

// Delete removes a value from cache
func (cm *CacheManager) Delete(ctx context.Context, key string) error {
	fullKey := cm.prefix + key

	if cm.useRedis {
		cm.redisClient.Del(ctx, fullKey)
	}

	cm.memCache.Delete(fullKey)
	return nil
}

// Clear removes all keys with the prefix
func (cm *CacheManager) Clear(ctx context.Context) error {
	if cm.useRedis {
		// Get all keys with prefix
		iter := cm.redisClient.Scan(ctx, 0, cm.prefix+"*", 0).Iterator()
		for iter.Next(ctx) {
			cm.redisClient.Del(ctx, iter.Val())
		}
		if err := iter.Err(); err != nil {
			return err
		}
	}

	// Clear memory cache
	cm.memCache.Range(func(key, value interface{}) bool {
		keyStr, ok := key.(string)
		if ok && len(keyStr) >= len(cm.prefix) && keyStr[:len(cm.prefix)] == cm.prefix {
			cm.memCache.Delete(key)
		}
		return true
	})

	return nil
}

// IsUsingRedis returns whether Redis is being used
func (cm *CacheManager) IsUsingRedis() bool {
	return cm.useRedis
}

// Close closes the cache manager and cleans up resources
func (cm *CacheManager) Close() error {
	if cm.useRedis && cm.redisClient != nil {
		return cm.redisClient.Close()
	}
	return nil
}

// GetOrSet gets a value from cache, or sets it if not found
func (cm *CacheManager) GetOrSet(ctx context.Context, key string, ttl time.Duration, fetchFn func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache
	if value, err := cm.Get(ctx, key); err == nil {
		return value, nil
	}

	// Fetch the value
	value, err := fetchFn()
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := cm.Set(ctx, key, value, ttl); err != nil {
		// Log error but don't fail
		fmt.Printf("Warning: Failed to cache value: %v\n", err)
	}

	return value, nil
}

// SetMulti sets multiple values at once
func (cm *CacheManager) SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	for key, value := range items {
		if err := cm.Set(ctx, key, value, ttl); err != nil {
			return fmt.Errorf("failed to set %s: %w", key, err)
		}
	}
	return nil
}

// GetMulti gets multiple values at once
func (cm *CacheManager) GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error) {
	results := make(map[string]interface{})

	for _, key := range keys {
		if value, err := cm.Get(ctx, key); err == nil {
			results[key] = value
		}
	}

	return results, nil
}

// Exists checks if a key exists in cache
func (cm *CacheManager) Exists(ctx context.Context, key string) bool {
	fullKey := cm.prefix + key

	if cm.useRedis {
		exists, err := cm.redisClient.Exists(ctx, fullKey).Result()
		if err == nil && exists > 0 {
			return true
		}
	}

	if _, ok := cm.memCache.Load(fullKey); ok {
		return true
	}

	return false
}

// TTL returns the time-to-live for a key
func (cm *CacheManager) TTL(ctx context.Context, key string) (time.Duration, error) {
	fullKey := cm.prefix + key

	if cm.useRedis {
		ttl, err := cm.redisClient.TTL(ctx, fullKey).Result()
		if err == nil {
			return ttl, nil
		}
	}

	// For memory cache, calculate from expiration
	if val, ok := cm.memCache.Load(fullKey); ok {
		item := val.(CacheItem)
		remaining := time.Until(item.ExpiresAt)
		if remaining > 0 {
			return remaining, nil
		}
	}

	return 0, fmt.Errorf("key not found or expired")
}
