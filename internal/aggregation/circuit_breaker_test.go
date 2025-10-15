package aggregation

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCircuitBreaker(t *testing.T) {
	t.Run("InitialState", func(t *testing.T) {
		cb := NewCircuitBreaker()

		if !cb.IsClosed() {
			t.Error("Circuit should start in closed state")
		}

		if cb.IsOpen() {
			t.Error("Circuit should not be open initially")
		}
	})

	t.Run("SuccessfulExecution", func(t *testing.T) {
		cb := NewCircuitBreaker()

		result, err := cb.Execute(context.Background(), func(ctx context.Context) (interface{}, error) {
			return "success", nil
		})

		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if result != "success" {
			t.Errorf("Expected 'success', got %v", result)
		}

		if !cb.IsClosed() {
			t.Error("Circuit should remain closed after success")
		}
	})

	t.Run("FailureThreshold", func(t *testing.T) {
		cb := NewCircuitBreaker(WithFailureThreshold(3))

		// Fail 3 times
		for i := 0; i < 3; i++ {
			_, err := cb.Execute(context.Background(), func(ctx context.Context) (interface{}, error) {
				return nil, errors.New("service error")
			})

			if err == nil {
				t.Error("Expected error")
			}
		}

		// Circuit should now be open
		if !cb.IsOpen() {
			t.Error("Circuit should be open after reaching failure threshold")
		}

		// Next request should be rejected immediately
		_, err := cb.Execute(context.Background(), func(ctx context.Context) (interface{}, error) {
			t.Error("This function should not be called when circuit is open")
			return nil, nil
		})

		if !errors.Is(err, ErrCircuitOpen) {
			t.Errorf("Expected ErrCircuitOpen, got %v", err)
		}
	})

	t.Run("HalfOpenRecovery", func(t *testing.T) {
		cb := NewCircuitBreaker(
			WithFailureThreshold(2),
			WithResetTimeout(100*time.Millisecond),
		)

		// Open the circuit
		for i := 0; i < 2; i++ {
			cb.Execute(context.Background(), func(ctx context.Context) (interface{}, error) {
				return nil, errors.New("error")
			})
		}

		if !cb.IsOpen() {
			t.Error("Circuit should be open")
		}

		// Wait for reset timeout
		time.Sleep(150 * time.Millisecond)

		// Next request should succeed and close the circuit
		_, err := cb.Execute(context.Background(), func(ctx context.Context) (interface{}, error) {
			return "recovered", nil
		})

		if err != nil {
			t.Errorf("Expected success after timeout, got error: %v", err)
		}
	})

	t.Run("ExecuteWithFallback", func(t *testing.T) {
		cb := NewCircuitBreaker(WithFailureThreshold(1))

		// Open the circuit
		cb.Execute(context.Background(), func(ctx context.Context) (interface{}, error) {
			return nil, errors.New("error")
		})

		// Execute with fallback
		result, err := cb.ExecuteWithFallback(
			context.Background(),
			func(ctx context.Context) (interface{}, error) {
				return nil, errors.New("should not be called")
			},
			func() (interface{}, error) {
				return "fallback-value", nil
			},
		)

		if err != nil {
			t.Errorf("Expected fallback to succeed, got error: %v", err)
		}

		if result != "fallback-value" {
			t.Errorf("Expected 'fallback-value', got %v", result)
		}
	})

	t.Run("GetState", func(t *testing.T) {
		cb := NewCircuitBreaker()

		state := cb.GetState()

		if state["state"] != "CLOSED" {
			t.Errorf("Expected state CLOSED, got %v", state["state"])
		}

		if state["failures"] != 0 {
			t.Errorf("Expected 0 failures, got %v", state["failures"])
		}
	})

	t.Run("Reset", func(t *testing.T) {
		cb := NewCircuitBreaker(WithFailureThreshold(1))

		// Open the circuit
		cb.Execute(context.Background(), func(ctx context.Context) (interface{}, error) {
			return nil, errors.New("error")
		})

		if !cb.IsOpen() {
			t.Error("Circuit should be open")
		}

		// Reset
		cb.Reset()

		if !cb.IsClosed() {
			t.Error("Circuit should be closed after reset")
		}
	})

	t.Run("ForceOpen", func(t *testing.T) {
		cb := NewCircuitBreaker()

		if !cb.IsClosed() {
			t.Error("Circuit should start closed")
		}

		cb.ForceOpen()

		if !cb.IsOpen() {
			t.Error("Circuit should be open after ForceOpen")
		}
	})

	t.Run("GetStats", func(t *testing.T) {
		cb := NewCircuitBreaker()

		// Execute some requests
		cb.Execute(context.Background(), func(ctx context.Context) (interface{}, error) {
			return "ok", nil
		})

		cb.Execute(context.Background(), func(ctx context.Context) (interface{}, error) {
			return nil, errors.New("error")
		})

		stats := cb.GetStats()

		if stats.Total != 2 {
			t.Errorf("Expected 2 total requests, got %d", stats.Total)
		}

		if stats.Succeeded != 1 {
			t.Errorf("Expected 1 success, got %d", stats.Succeeded)
		}

		if stats.Failed != 1 {
			t.Errorf("Expected 1 failure, got %d", stats.Failed)
		}
	})
}
