package aggregation

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// BatchFunc is a function that processes a batch of IDs
type BatchFunc func(ctx context.Context, ids []string) ([]interface{}, error)

// RequestBatcher batches requests to prevent N+1 queries
type RequestBatcher struct {
	mu           sync.RWMutex
	batches      map[string]*batch
	batchWindow  time.Duration
	maxBatchSize int
}

// batch represents a pending batch of requests
type batch struct {
	mu       sync.Mutex
	ids      []string
	promises map[string]*promise
	timer    *time.Timer
	batchFn  BatchFunc
}

// promise represents a pending request result
type promise struct {
	result interface{}
	err    error
	done   chan struct{}
}

// NewRequestBatcher creates a new request batcher
func NewRequestBatcher(batchWindow time.Duration, maxBatchSize int) *RequestBatcher {
	if batchWindow == 0 {
		batchWindow = 10 * time.Millisecond
	}
	if maxBatchSize == 0 {
		maxBatchSize = 50
	}

	return &RequestBatcher{
		batches:      make(map[string]*batch),
		batchWindow:  batchWindow,
		maxBatchSize: maxBatchSize,
	}
}

// Batch adds a request to a batch and returns the result when ready
func (rb *RequestBatcher) Batch(ctx context.Context, key string, batchFn BatchFunc, id string) (interface{}, error) {
	rb.mu.Lock()

	// Get or create batch for this key
	b, exists := rb.batches[key]
	if !exists {
		b = &batch{
			ids:      make([]string, 0),
			promises: make(map[string]*promise),
			batchFn:  batchFn,
		}
		rb.batches[key] = b

		// Start timer for this batch
		b.timer = time.AfterFunc(rb.batchWindow, func() {
			rb.executeBatch(ctx, key)
		})
	}

	b.mu.Lock()
	rb.mu.Unlock()

	// Check if this ID is already in the batch
	if p, ok := b.promises[id]; ok {
		b.mu.Unlock()
		// Wait for result
		<-p.done
		return p.result, p.err
	}

	// Create promise for this request
	p := &promise{
		done: make(chan struct{}),
	}
	b.promises[id] = p
	b.ids = append(b.ids, id)

	// Check if batch is full
	if len(b.ids) >= rb.maxBatchSize {
		// Stop timer and execute immediately
		b.timer.Stop()
		b.mu.Unlock()
		rb.executeBatch(ctx, key)
	} else {
		b.mu.Unlock()
	}

	// Wait for result
	<-p.done
	return p.result, p.err
}

// executeBatch executes a batch of requests
func (rb *RequestBatcher) executeBatch(ctx context.Context, key string) {
	rb.mu.Lock()
	b, exists := rb.batches[key]
	if !exists {
		rb.mu.Unlock()
		return
	}

	// Remove batch from map
	delete(rb.batches, key)
	rb.mu.Unlock()

	b.mu.Lock()
	ids := b.ids
	promises := b.promises
	batchFn := b.batchFn
	b.mu.Unlock()

	// Execute batch function
	results, err := batchFn(ctx, ids)

	// Resolve promises
	if err != nil {
		// All requests failed
		for _, p := range promises {
			p.err = err
			close(p.done)
		}
		return
	}

	// Create result map by ID
	resultMap := make(map[string]interface{})
	for _, result := range results {
		// Assume result has "id" field
		if resultMap, ok := result.(map[string]interface{}); ok {
			if id, ok := resultMap["id"].(string); ok {
				resultMap[id] = result
			}
		}
	}

	// Resolve each promise
	for id, p := range promises {
		if result, ok := resultMap[id]; ok {
			p.result = result
		} else {
			p.err = fmt.Errorf("no result found for ID: %s", id)
		}
		close(p.done)
	}
}

// Clear clears a specific batch
func (rb *RequestBatcher) Clear(key string) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if b, exists := rb.batches[key]; exists {
		b.timer.Stop()
		delete(rb.batches, key)
	}
}

// ClearAll clears all batches
func (rb *RequestBatcher) ClearAll() {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	for _, b := range rb.batches {
		b.timer.Stop()
	}
	rb.batches = make(map[string]*batch)
}

// GetStats returns statistics about pending batches
func (rb *RequestBatcher) GetStats() map[string]interface{} {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	stats := map[string]interface{}{
		"activeBatches":        len(rb.batches),
		"totalPendingRequests": 0,
	}

	batches := make([]map[string]interface{}, 0, len(rb.batches))
	for key, b := range rb.batches {
		b.mu.Lock()
		batchInfo := map[string]interface{}{
			"key":     key,
			"pending": len(b.ids),
		}
		stats["totalPendingRequests"] = stats["totalPendingRequests"].(int) + len(b.ids)
		batches = append(batches, batchInfo)
		b.mu.Unlock()
	}

	stats["batches"] = batches
	return stats
}
