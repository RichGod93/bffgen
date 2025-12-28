package aggregation

import (
	"context"
	"testing"
	"time"
)

func TestNewCacheManager(t *testing.T) {
	t.Run("creates cache manager with in-memory fallback", func(t *testing.T) {
		// Use empty redis URL to force in-memory fallback
		cm := NewCacheManager("", 5*time.Minute, "test:")

		if cm == nil {
			t.Fatal("NewCacheManager() returned nil")
		}

		if cm.ttl != 5*time.Minute {
			t.Errorf("Expected TTL 5m, got %v", cm.ttl)
		}

		if cm.prefix != "test:" {
			t.Errorf("Expected prefix 'test:', got %q", cm.prefix)
		}

		if cm.useRedis {
			t.Error("Expected useRedis to be false for empty URL")
		}
	})

	t.Run("fails gracefully with invalid redis URL", func(t *testing.T) {
		cm := NewCacheManager("invalid://url", 5*time.Minute, "test:")

		if cm.useRedis {
			t.Error("Expected useRedis to be false for invalid URL")
		}
	})
}

func TestCacheManager_SetAndGet(t *testing.T) {
	cm := NewCacheManager("", 5*time.Minute, "test:")
	ctx := context.Background()

	t.Run("set and get string value", func(t *testing.T) {
		err := cm.Set(ctx, "key1", "value1", 0)
		if err != nil {
			t.Fatalf("Set() failed: %v", err)
		}

		val, err := cm.Get(ctx, "key1")
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}

		if val != "value1" {
			t.Errorf("Expected 'value1', got %v", val)
		}
	})

	t.Run("set and get complex value", func(t *testing.T) {
		data := map[string]interface{}{
			"name":  "test",
			"count": float64(42),
		}

		err := cm.Set(ctx, "key2", data, 0)
		if err != nil {
			t.Fatalf("Set() failed: %v", err)
		}

		val, err := cm.Get(ctx, "key2")
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}

		valMap, ok := val.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map, got %T", val)
		}

		if valMap["name"] != "test" {
			t.Errorf("Expected name 'test', got %v", valMap["name"])
		}
	})

	t.Run("returns error for missing key", func(t *testing.T) {
		_, err := cm.Get(ctx, "nonexistent")
		if err == nil {
			t.Error("Expected error for missing key")
		}
	})
}

func TestCacheManager_Delete(t *testing.T) {
	cm := NewCacheManager("", 5*time.Minute, "test:")
	ctx := context.Background()

	// Set a value
	cm.Set(ctx, "delete-key", "value", 0)

	// Delete it
	err := cm.Delete(ctx, "delete-key")
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// Verify it's gone
	_, err = cm.Get(ctx, "delete-key")
	if err == nil {
		t.Error("Expected error after deletion")
	}
}

func TestCacheManager_Clear(t *testing.T) {
	cm := NewCacheManager("", 5*time.Minute, "test:")
	ctx := context.Background()

	// Set multiple values
	cm.Set(ctx, "clear1", "value1", 0)
	cm.Set(ctx, "clear2", "value2", 0)

	// Clear all
	err := cm.Clear(ctx)
	if err != nil {
		t.Fatalf("Clear() failed: %v", err)
	}

	// Verify both are gone
	_, err1 := cm.Get(ctx, "clear1")
	_, err2 := cm.Get(ctx, "clear2")

	if err1 == nil || err2 == nil {
		t.Error("Expected errors after clearing")
	}
}

func TestCacheManager_IsUsingRedis(t *testing.T) {
	cm := NewCacheManager("", 5*time.Minute, "test:")

	if cm.IsUsingRedis() {
		t.Error("Expected IsUsingRedis() to return false")
	}
}

func TestCacheManager_Close(t *testing.T) {
	cm := NewCacheManager("", 5*time.Minute, "test:")

	err := cm.Close()
	if err != nil {
		t.Fatalf("Close() failed: %v", err)
	}
}

func TestCacheManager_GetOrSet(t *testing.T) {
	cm := NewCacheManager("", 5*time.Minute, "test:")
	ctx := context.Background()

	callCount := 0

	t.Run("fetches and caches value", func(t *testing.T) {
		val, err := cm.GetOrSet(ctx, "getorset-key", 0, func() (interface{}, error) {
			callCount++
			return "fetched-value", nil
		})

		if err != nil {
			t.Fatalf("GetOrSet() failed: %v", err)
		}

		if val != "fetched-value" {
			t.Errorf("Expected 'fetched-value', got %v", val)
		}

		if callCount != 1 {
			t.Errorf("Expected fetch to be called once, got %d", callCount)
		}
	})

	t.Run("returns cached value on second call", func(t *testing.T) {
		val, err := cm.GetOrSet(ctx, "getorset-key", 0, func() (interface{}, error) {
			callCount++
			return "new-value", nil
		})

		if err != nil {
			t.Fatalf("GetOrSet() failed: %v", err)
		}

		if val != "fetched-value" {
			t.Errorf("Expected cached 'fetched-value', got %v", val)
		}

		if callCount != 1 {
			t.Errorf("Expected fetch to still be called once, got %d", callCount)
		}
	})
}

func TestCacheManager_SetMulti(t *testing.T) {
	cm := NewCacheManager("", 5*time.Minute, "test:")
	ctx := context.Background()

	items := map[string]interface{}{
		"multi1": "value1",
		"multi2": "value2",
		"multi3": "value3",
	}

	err := cm.SetMulti(ctx, items, 0)
	if err != nil {
		t.Fatalf("SetMulti() failed: %v", err)
	}

	// Verify all values are set
	for key, expected := range items {
		val, err := cm.Get(ctx, key)
		if err != nil {
			t.Errorf("Get(%q) failed: %v", key, err)
			continue
		}
		if val != expected {
			t.Errorf("Expected %q = %q, got %q", key, expected, val)
		}
	}
}

func TestCacheManager_GetMulti(t *testing.T) {
	cm := NewCacheManager("", 5*time.Minute, "test:")
	ctx := context.Background()

	// Set some values
	cm.Set(ctx, "getmulti1", "value1", 0)
	cm.Set(ctx, "getmulti2", "value2", 0)

	keys := []string{"getmulti1", "getmulti2", "nonexistent"}
	results, err := cm.GetMulti(ctx, keys)

	if err != nil {
		t.Fatalf("GetMulti() failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	if results["getmulti1"] != "value1" {
		t.Errorf("Expected 'value1', got %v", results["getmulti1"])
	}

	if results["getmulti2"] != "value2" {
		t.Errorf("Expected 'value2', got %v", results["getmulti2"])
	}
}

func TestCacheManager_Exists(t *testing.T) {
	cm := NewCacheManager("", 5*time.Minute, "test:")
	ctx := context.Background()

	t.Run("returns false for non-existent key", func(t *testing.T) {
		if cm.Exists(ctx, "nonexistent") {
			t.Error("Expected Exists() to return false")
		}
	})

	t.Run("returns true for existing key", func(t *testing.T) {
		cm.Set(ctx, "exists-key", "value", 0)

		if !cm.Exists(ctx, "exists-key") {
			t.Error("Expected Exists() to return true")
		}
	})
}

func TestCacheManager_TTL(t *testing.T) {
	cm := NewCacheManager("", 5*time.Minute, "test:")
	ctx := context.Background()

	t.Run("returns error for non-existent key", func(t *testing.T) {
		_, err := cm.TTL(ctx, "nonexistent")
		if err == nil {
			t.Error("Expected error for non-existent key")
		}
	})

	t.Run("returns remaining TTL for existing key", func(t *testing.T) {
		cm.Set(ctx, "ttl-key", "value", 5*time.Minute)

		ttl, err := cm.TTL(ctx, "ttl-key")
		if err != nil {
			t.Fatalf("TTL() failed: %v", err)
		}

		// TTL should be close to 5 minutes (allow some tolerance)
		if ttl < 4*time.Minute || ttl > 5*time.Minute {
			t.Errorf("Expected TTL around 5 minutes, got %v", ttl)
		}
	})
}

func TestCacheManager_Expiration(t *testing.T) {
	cm := NewCacheManager("", 5*time.Minute, "test:")
	ctx := context.Background()

	// Set with very short TTL
	cm.Set(ctx, "expire-key", "value", 50*time.Millisecond)

	// Value should exist initially
	_, err := cm.Get(ctx, "expire-key")
	if err != nil {
		t.Fatal("Expected value to exist initially")
	}

	// Wait for expiration
	time.Sleep(60 * time.Millisecond)

	// Value should be expired
	_, err = cm.Get(ctx, "expire-key")
	if err == nil {
		t.Error("Expected error after expiration")
	}
}

func TestCacheItem(t *testing.T) {
	item := CacheItem{
		Data:      []byte(`{"test": "value"}`),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	if len(item.Data) == 0 {
		t.Error("Expected Data to be set")
	}

	if item.ExpiresAt.Before(time.Now()) {
		t.Error("Expected ExpiresAt to be in the future")
	}
}
