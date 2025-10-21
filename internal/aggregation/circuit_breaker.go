package aggregation

import (
	"context"
	"errors"
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	// StateClosed allows requests to pass through
	StateClosed CircuitState = iota
	// StateOpen blocks requests and returns errors
	StateOpen
	// StateHalfOpen allows a test request to check if service recovered
	StateHalfOpen
)

func (cs CircuitState) String() string {
	switch cs {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// ErrCircuitOpen is returned when circuit is open
var ErrCircuitOpen = errors.New("circuit breaker is open")

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu               sync.RWMutex
	state            CircuitState
	failureCount     int
	successCount     int
	failureThreshold int
	successThreshold int
	resetTimeout     time.Duration
	monitoringPeriod time.Duration
	nextAttemptTime  time.Time
	stats            *CircuitStats
}

// CircuitStats tracks circuit breaker statistics
type CircuitStats struct {
	Total     int
	Succeeded int
	Failed    int
	Opened    int
	Closed    int
}

// CircuitBreakerOption is a functional option for circuit breaker
type CircuitBreakerOption func(*CircuitBreaker)

// NewCircuitBreaker creates a new circuit breaker with options
func NewCircuitBreaker(options ...CircuitBreakerOption) *CircuitBreaker {
	cb := &CircuitBreaker{
		state:            StateClosed,
		failureThreshold: 5,
		successThreshold: 2,
		resetTimeout:     60 * time.Second,
		monitoringPeriod: 10 * time.Second,
		stats:            &CircuitStats{},
	}

	for _, option := range options {
		option(cb)
	}

	return cb
}

// WithFailureThreshold sets the failure threshold
func WithFailureThreshold(threshold int) CircuitBreakerOption {
	return func(cb *CircuitBreaker) {
		cb.failureThreshold = threshold
	}
}

// WithResetTimeout sets the reset timeout
func WithResetTimeout(timeout time.Duration) CircuitBreakerOption {
	return func(cb *CircuitBreaker) {
		cb.resetTimeout = timeout
	}
}

// WithMonitoringPeriod sets the monitoring period
func WithMonitoringPeriod(period time.Duration) CircuitBreakerOption {
	return func(cb *CircuitBreaker) {
		cb.monitoringPeriod = period
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	cb.mu.Lock()

	cb.stats.Total++

	// Check current state
	switch cb.state {
	case StateOpen:
		// Check if it's time to try again
		if time.Now().After(cb.nextAttemptTime) {
			cb.state = StateHalfOpen
			cb.mu.Unlock()
		} else {
			cb.mu.Unlock()
			return nil, ErrCircuitOpen
		}

	case StateClosed:
		cb.mu.Unlock()

	case StateHalfOpen:
		cb.mu.Unlock()
	}

	// Execute the function
	result, err := fn(ctx)

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.onFailure()
		return nil, err
	}

	cb.onSuccess()
	return result, nil
}

// ExecuteWithFallback executes with a fallback function if circuit is open
func (cb *CircuitBreaker) ExecuteWithFallback(ctx context.Context, fn func(ctx context.Context) (interface{}, error), fallback func() (interface{}, error)) (interface{}, error) {
	result, err := cb.Execute(ctx, fn)

	if errors.Is(err, ErrCircuitOpen) {
		// Use fallback
		return fallback()
	}

	return result, err
}

// onFailure handles a failed request
func (cb *CircuitBreaker) onFailure() {
	cb.failureCount++
	cb.stats.Failed++

	switch cb.state {
	case StateClosed:
		if cb.failureCount >= cb.failureThreshold {
			cb.state = StateOpen
			cb.nextAttemptTime = time.Now().Add(cb.resetTimeout)
			cb.failureCount = 0
			cb.stats.Opened++
		}

	case StateHalfOpen:
		// Failed during half-open, go back to open
		cb.state = StateOpen
		cb.nextAttemptTime = time.Now().Add(cb.resetTimeout)
		cb.failureCount = 0
		cb.successCount = 0
		cb.stats.Opened++
	}
}

// onSuccess handles a successful request
func (cb *CircuitBreaker) onSuccess() {
	cb.stats.Succeeded++

	switch cb.state {
	case StateClosed:
		cb.failureCount = 0

	case StateHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.state = StateClosed
			cb.failureCount = 0
			cb.successCount = 0
			cb.stats.Closed++
		}
	}
}

// GetState returns the current state and statistics
func (cb *CircuitBreaker) GetState() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	var nextAttempt *time.Time
	if cb.state == StateOpen {
		nextAttempt = &cb.nextAttemptTime
	}

	return map[string]interface{}{
		"state":       cb.state.String(),
		"failures":    cb.failureCount,
		"successes":   cb.successCount,
		"stats":       cb.stats,
		"nextAttempt": nextAttempt,
	}
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failureCount = 0
	cb.successCount = 0
	cb.stats.Closed++
}

// ForceOpen forces the circuit breaker to open state
func (cb *CircuitBreaker) ForceOpen() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateOpen
	cb.nextAttemptTime = time.Now().Add(cb.resetTimeout)
	cb.stats.Opened++
}

// GetStats returns the current statistics
func (cb *CircuitBreaker) GetStats() *CircuitStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return &CircuitStats{
		Total:     cb.stats.Total,
		Succeeded: cb.stats.Succeeded,
		Failed:    cb.stats.Failed,
		Opened:    cb.stats.Opened,
		Closed:    cb.stats.Closed,
	}
}

// IsOpen returns whether the circuit is open
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return cb.state == StateOpen
}

// IsClosed returns whether the circuit is closed
func (cb *CircuitBreaker) IsClosed() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return cb.state == StateClosed
}

// IsHalfOpen returns whether the circuit is half-open
func (cb *CircuitBreaker) IsHalfOpen() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return cb.state == StateHalfOpen
}
