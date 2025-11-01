package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// Status represents the health check response
type Status struct {
	Status       string            `json:"status"`
	Version      string            `json:"version"`
	Timestamp    time.Time         `json:"timestamp"`
	Dependencies map[string]bool   `json:"dependencies,omitempty"`
}

// Checker provides health check functionality
type Checker struct {
	version  string
	backends []string
	mu       sync.RWMutex
}

// NewChecker creates a new health checker
func NewChecker(version string, backends []string) *Checker {
	return &Checker{
		version:  version,
		backends: backends,
	}
}

// Liveness returns a basic health check for liveness probe
func (c *Checker) Liveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Status{
		Status:    "ok",
		Timestamp: time.Now(),
	})
}

// Readiness returns a health check with dependency validation for readiness probe
func (c *Checker) Readiness(w http.ResponseWriter, r *http.Request) {
	c.mu.RLock()
	backends := c.backends
	c.mu.RUnlock()

	deps := make(map[string]bool)
	allHealthy := true

	// Check backend services in parallel
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, backend := range backends {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			healthy := checkBackend(url)
			mu.Lock()
			deps[url] = healthy
			if !healthy {
				allHealthy = false
			}
			mu.Unlock()
		}(backend)
	}

	wg.Wait()

	status := "ok"
	statusCode := http.StatusOK
	if !allHealthy {
		status = "degraded"
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Status{
		Status:       status,
		Version:      c.version,
		Timestamp:    time.Now(),
		Dependencies: deps,
	})
}

// checkBackend verifies if a backend service is healthy
func checkBackend(url string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Try common health check endpoints
	healthPaths := []string{"/health", "/healthz", "/health/readiness"}
	
	for _, path := range healthPaths {
		req, err := http.NewRequestWithContext(ctx, "GET", url+path, nil)
		if err != nil {
			continue
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return true
		}
	}

	return false
}

