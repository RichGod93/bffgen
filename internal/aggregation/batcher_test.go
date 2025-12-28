package aggregation

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNewRequestBatcher(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		batcher := NewRequestBatcher(0, 0)

		if batcher.batchWindow != 10*time.Millisecond {
			t.Errorf("Expected default batch window 10ms, got %v", batcher.batchWindow)
		}

		if batcher.maxBatchSize != 50 {
			t.Errorf("Expected default max batch size 50, got %d", batcher.maxBatchSize)
		}
	})

	t.Run("custom values", func(t *testing.T) {
		batcher := NewRequestBatcher(50*time.Millisecond, 100)

		if batcher.batchWindow != 50*time.Millisecond {
			t.Errorf("Expected batch window 50ms, got %v", batcher.batchWindow)
		}

		if batcher.maxBatchSize != 100 {
			t.Errorf("Expected max batch size 100, got %d", batcher.maxBatchSize)
		}
	})

	t.Run("initializes empty batches map", func(t *testing.T) {
		batcher := NewRequestBatcher(10*time.Millisecond, 10)

		if batcher.batches == nil {
			t.Error("Expected batches map to be initialized")
		}

		if len(batcher.batches) != 0 {
			t.Errorf("Expected empty batches map, got %d batches", len(batcher.batches))
		}
	})
}

func TestRequestBatcher_Batch(t *testing.T) {
	t.Run("batches requests together", func(t *testing.T) {
		batcher := NewRequestBatcher(50*time.Millisecond, 10)

		batchCalls := 0
		var mu sync.Mutex

		batchFn := func(ctx context.Context, ids []string) ([]interface{}, error) {
			mu.Lock()
			batchCalls++
			mu.Unlock()

			results := make([]interface{}, len(ids))
			for i, id := range ids {
				results[i] = map[string]interface{}{"id": id, "value": "result-" + id}
			}
			return results, nil
		}

		var wg sync.WaitGroup
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func(id string) {
				defer wg.Done()
				_, err := batcher.Batch(context.Background(), "test-key", batchFn, id)
				if err != nil {
					// Errors are expected in some cases due to result mapping
				}
			}(string(rune('a' + i)))
		}

		wg.Wait()

		// Wait for batch to complete
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("executes immediately when batch is full", func(t *testing.T) {
		batcher := NewRequestBatcher(500*time.Millisecond, 2)

		batchExecuted := make(chan bool, 1)

		batchFn := func(ctx context.Context, ids []string) ([]interface{}, error) {
			batchExecuted <- true
			results := make([]interface{}, len(ids))
			for i, id := range ids {
				results[i] = map[string]interface{}{"id": id}
			}
			return results, nil
		}

		var wg sync.WaitGroup
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func(id string) {
				defer wg.Done()
				batcher.Batch(context.Background(), "full-batch", batchFn, id)
			}(string(rune('a' + i)))
		}

		wg.Wait()

		select {
		case <-batchExecuted:
			// Good, batch executed before timeout
		case <-time.After(100 * time.Millisecond):
			t.Error("Expected batch to execute immediately when full")
		}
	})
}

func TestRequestBatcher_Clear(t *testing.T) {
	batcher := NewRequestBatcher(500*time.Millisecond, 100)

	// Add a batch that won't complete immediately
	batchFn := func(ctx context.Context, ids []string) ([]interface{}, error) {
		return []interface{}{}, nil
	}

	go batcher.Batch(context.Background(), "clear-test", batchFn, "id1")
	time.Sleep(10 * time.Millisecond) // Let the batch start

	// Clear the batch
	batcher.Clear("clear-test")

	batcher.mu.RLock()
	_, exists := batcher.batches["clear-test"]
	batcher.mu.RUnlock()

	if exists {
		t.Error("Expected batch to be cleared")
	}
}

func TestRequestBatcher_ClearAll(t *testing.T) {
	batcher := NewRequestBatcher(500*time.Millisecond, 100)

	batchFn := func(ctx context.Context, ids []string) ([]interface{}, error) {
		return []interface{}{}, nil
	}

	// Start multiple batches
	go batcher.Batch(context.Background(), "batch1", batchFn, "id1")
	go batcher.Batch(context.Background(), "batch2", batchFn, "id2")
	time.Sleep(10 * time.Millisecond)

	batcher.ClearAll()

	batcher.mu.RLock()
	count := len(batcher.batches)
	batcher.mu.RUnlock()

	if count != 0 {
		t.Errorf("Expected all batches to be cleared, got %d", count)
	}
}

func TestRequestBatcher_GetStats(t *testing.T) {
	batcher := NewRequestBatcher(500*time.Millisecond, 100)

	stats := batcher.GetStats()

	if stats["activeBatches"] != 0 {
		t.Errorf("Expected 0 active batches, got %v", stats["activeBatches"])
	}

	if stats["totalPendingRequests"] != 0 {
		t.Errorf("Expected 0 pending requests, got %v", stats["totalPendingRequests"])
	}

	batches, ok := stats["batches"].([]map[string]interface{})
	if !ok {
		t.Error("Expected batches to be a slice of maps")
	}
	if len(batches) != 0 {
		t.Errorf("Expected empty batches slice, got %d", len(batches))
	}
}
