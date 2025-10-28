package aggregators

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// HTTPClient provides HTTP client functionality for aggregators
type HTTPClient struct {
	client  *http.Client
	timeout time.Duration
}

// NewHTTPClient creates a new HTTP client for aggregators
func NewHTTPClient(timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

// Get makes a GET request to the specified URL
func (hc *HTTPClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add common headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "bffgen-aggregator/1.0")

	resp, err := hc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	return resp, nil
}

// Post makes a POST request to the specified URL
func (hc *HTTPClient) Post(url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add common headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "bffgen-aggregator/1.0")

	resp, err := hc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	return resp, nil
}

// ServiceCall represents a call to a backend service
type ServiceCall struct {
	ServiceName string
	URL         string
	Method      string
	Headers     map[string]string
	Body        []byte
	Timeout     time.Duration
}

// ServiceResponse represents the response from a service call
type ServiceResponse struct {
	ServiceName string
	StatusCode  int
	Headers     map[string][]string
	Body        []byte
	Error       error
	Duration    time.Duration
}

// ParallelCaller allows making multiple service calls in parallel
type ParallelCaller struct {
	client *HTTPClient
}

// NewParallelCaller creates a new parallel caller
func NewParallelCaller(timeout time.Duration) *ParallelCaller {
	return &ParallelCaller{
		client: NewHTTPClient(timeout),
	}
}

// CallServices makes multiple service calls in parallel
func (pc *ParallelCaller) CallServices(calls []ServiceCall) []ServiceResponse {
	responses := make([]ServiceResponse, len(calls))
	responseChan := make(chan ServiceResponse, len(calls))

	// Start all calls in parallel
	for i, call := range calls {
		go func(index int, serviceCall ServiceCall) {
			start := time.Now()

			var resp *http.Response
			var err error

			switch serviceCall.Method {
			case "GET":
				resp, err = pc.client.Get(serviceCall.URL)
			case "POST":
				resp, err = pc.client.Post(serviceCall.URL, serviceCall.Body)
			default:
				err = fmt.Errorf("unsupported method: %s", serviceCall.Method)
			}

			duration := time.Since(start)

			response := ServiceResponse{
				ServiceName: serviceCall.ServiceName,
				Duration:    duration,
				Error:       err,
			}

			if err == nil && resp != nil {
				response.StatusCode = resp.StatusCode
				response.Headers = resp.Header

				// Read response body
				if resp.Body != nil {
					defer func() {
						if closeErr := resp.Body.Close(); closeErr != nil {
							response.Error = fmt.Errorf("failed to close response body: %w", closeErr)
						}
					}()
					body, readErr := io.ReadAll(resp.Body)
					if readErr != nil {
						response.Error = fmt.Errorf("failed to read response body: %w", readErr)
					} else {
						response.Body = body
					}
				}
			}

			responseChan <- response
		}(i, call)
	}

	// Collect all responses
	for i := 0; i < len(calls); i++ {
		responses[i] = <-responseChan
	}

	return responses
}

// Cache provides caching functionality for aggregators
type Cache struct {
	data    map[string]CacheEntry
	ttl     time.Duration
	cleanup *time.Ticker
	mutex   sync.RWMutex
}

// CacheEntry represents a cached item
type CacheEntry struct {
	Data      []byte
	Timestamp time.Time
	TTL       time.Duration
}

// NewCache creates a new cache with the specified TTL
func NewCache(ttl time.Duration) *Cache {
	cache := &Cache{
		data: make(map[string]CacheEntry),
		ttl:  ttl,
	}

	// Start cleanup routine
	cache.cleanup = time.NewTicker(ttl / 2)
	go cache.cleanupRoutine()

	return cache
}

// Get retrieves data from cache
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mutex.RLock()
	entry, exists := c.data[key]
	c.mutex.RUnlock()

	if !exists {
		return nil, false
	}

	// Check if entry has expired
	if time.Since(entry.Timestamp) > entry.TTL {
		c.mutex.Lock()
		delete(c.data, key)
		c.mutex.Unlock()
		return nil, false
	}

	return entry.Data, true
}

// Set stores data in cache
func (c *Cache) Set(key string, data []byte) {
	c.mutex.Lock()
	c.data[key] = CacheEntry{
		Data:      data,
		Timestamp: time.Now(),
		TTL:       c.ttl,
	}
	c.mutex.Unlock()
}

// Delete removes data from cache
func (c *Cache) Delete(key string) {
	c.mutex.Lock()
	delete(c.data, key)
	c.mutex.Unlock()
}

// Clear removes all data from cache
func (c *Cache) Clear() {
	c.mutex.Lock()
	c.data = make(map[string]CacheEntry)
	c.mutex.Unlock()
}

// cleanupRoutine periodically removes expired entries
func (c *Cache) cleanupRoutine() {
	for range c.cleanup.C {
		c.mutex.Lock()
		now := time.Now()
		for key, entry := range c.data {
			if now.Sub(entry.Timestamp) > entry.TTL {
				delete(c.data, key)
			}
		}
		c.mutex.Unlock()
	}
}

// Close stops the cleanup routine
func (c *Cache) Close() {
	if c.cleanup != nil {
		c.cleanup.Stop()
	}
}
