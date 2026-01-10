package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CacheEntry represents a cached version with timestamp
type CacheEntry struct {
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

// VersionCache handles caching of fetched package versions
type VersionCache struct {
	cacheDir string
	ttl      time.Duration
}

// NewVersionCache creates a new version cache
// Cache is stored in ~/.bffgen/cache/versions/
func NewVersionCache() (*VersionCache, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	cacheDir := filepath.Join(homeDir, ".bffgen", "cache", "versions")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &VersionCache{
		cacheDir: cacheDir,
		ttl:      24 * time.Hour, // 24 hour TTL
	}, nil
}

// Get retrieves a cached version if it exists and is not expired
func (c *VersionCache) Get(runtime, packageName string) (string, bool) {
	cacheFile := filepath.Join(c.cacheDir, fmt.Sprintf("%s_%s.json", runtime, packageName))

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return "", false // Cache miss
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return "", false
	}

	// Check if expired
	if time.Since(entry.Timestamp) > c.ttl {
		return "", false // Expired
	}

	return entry.Version, true
}

// Set stores a version in the cache
func (c *VersionCache) Set(runtime, packageName, version string) error {
	entry := CacheEntry{
		Version:   version,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	cacheFile := filepath.Join(c.cacheDir, fmt.Sprintf("%s_%s.json", runtime, packageName))
	if err := os.WriteFile(cacheFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// Clear removes all cached entries
func (c *VersionCache) Clear() error {
	return os.RemoveAll(c.cacheDir)
}

// ClearExpired removes expired cache entries
func (c *VersionCache) ClearExpired() error {
	entries, err := os.ReadDir(c.cacheDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(c.cacheDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var cacheEntry CacheEntry
		if err := json.Unmarshal(data, &cacheEntry); err != nil {
			continue
		}

		if time.Since(cacheEntry.Timestamp) > c.ttl {
			os.Remove(filePath)
		}
	}

	return nil
}
