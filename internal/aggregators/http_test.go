package aggregators

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHTTPClient(t *testing.T) {
	timeout := 5 * time.Second
	client := NewHTTPClient(timeout)

	if client == nil {
		t.Fatal("Expected HTTPClient, got nil")
	}

	if client.timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, client.timeout)
	}

	if client.client.Timeout != timeout {
		t.Errorf("Expected client timeout %v, got %v", timeout, client.client.Timeout)
	}
}

func TestHTTPClient_Get(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Expected Accept: application/json, got %s", r.Header.Get("Accept"))
		}
		if r.Header.Get("User-Agent") != "bffgen-aggregator/1.0" {
			t.Errorf("Expected User-Agent: bffgen-aggregator/1.0, got %s", r.Header.Get("User-Agent"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	client := NewHTTPClient(5 * time.Second)
	resp, err := client.Get(server.URL)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestHTTPClient_Post(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id": "123"}`))
	}))
	defer server.Close()

	client := NewHTTPClient(5 * time.Second)
	body := []byte(`{"name": "test"}`)
	resp, err := client.Post(server.URL, body)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}
}

func TestNewParallelCaller(t *testing.T) {
	timeout := 3 * time.Second
	caller := NewParallelCaller(timeout)

	if caller == nil {
		t.Fatal("Expected ParallelCaller, got nil")
	}

	if caller.client.timeout != timeout {
		t.Errorf("Expected client timeout %v, got %v", timeout, caller.client.timeout)
	}
}

func TestParallelCaller_CallServices(t *testing.T) {
	// Create test servers
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"service": "user"}`))
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"service": "orders"}`))
	}))
	defer server2.Close()

	caller := NewParallelCaller(5 * time.Second)
	calls := []ServiceCall{
		{
			ServiceName: "user-service",
			URL:         server1.URL,
			Method:      "GET",
		},
		{
			ServiceName: "orders-service",
			URL:         server2.URL,
			Method:      "GET",
		},
	}

	responses := caller.CallServices(calls)

	if len(responses) != 2 {
		t.Fatalf("Expected 2 responses, got %d", len(responses))
	}

	for _, resp := range responses {
		if resp.Error != nil {
			t.Errorf("Expected no error, got %v", resp.Error)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
		if resp.Duration <= 0 {
			t.Errorf("Expected positive duration, got %v", resp.Duration)
		}
	}
}

func TestNewCache(t *testing.T) {
	ttl := 1 * time.Second
	cache := NewCache(ttl)
	defer cache.Close()

	if cache == nil {
		t.Fatal("Expected Cache, got nil")
	}

	if cache.ttl != ttl {
		t.Errorf("Expected TTL %v, got %v", ttl, cache.ttl)
	}

	if cache.data == nil {
		t.Fatal("Expected data map, got nil")
	}
}

func TestCache_SetGet(t *testing.T) {
	cache := NewCache(1 * time.Second)
	defer cache.Close()

	key := "test-key"
	data := []byte("test-data")

	// Test Set
	cache.Set(key, data)

	// Test Get
	retrievedData, exists := cache.Get(key)
	if !exists {
		t.Fatal("Expected data to exist")
	}

	if string(retrievedData) != string(data) {
		t.Errorf("Expected %s, got %s", string(data), string(retrievedData))
	}
}

func TestCache_Expiration(t *testing.T) {
	cache := NewCache(100 * time.Millisecond)
	defer cache.Close()

	key := "test-key"
	data := []byte("test-data")

	cache.Set(key, data)

	// Data should exist immediately
	_, exists := cache.Get(key)
	if !exists {
		t.Fatal("Expected data to exist immediately")
	}

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Data should be expired
	_, exists = cache.Get(key)
	if exists {
		t.Fatal("Expected data to be expired")
	}
}

func TestCache_Delete(t *testing.T) {
	cache := NewCache(1 * time.Second)
	defer cache.Close()

	key := "test-key"
	data := []byte("test-data")

	cache.Set(key, data)
	cache.Delete(key)

	_, exists := cache.Get(key)
	if exists {
		t.Fatal("Expected data to be deleted")
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache(1 * time.Second)
	defer cache.Close()

	cache.Set("key1", []byte("data1"))
	cache.Set("key2", []byte("data2"))

	cache.Clear()

	_, exists1 := cache.Get("key1")
	_, exists2 := cache.Get("key2")

	if exists1 || exists2 {
		t.Fatal("Expected all data to be cleared")
	}
}

func TestNewUserDashboardAggregator(t *testing.T) {
	agg := NewUserDashboardAggregator()

	if agg == nil {
		t.Fatal("Expected UserDashboardAggregator, got nil")
	}

	if agg.GetName() != "user-dashboard" {
		t.Errorf("Expected name 'user-dashboard', got %s", agg.GetName())
	}

	if agg.GetPath() != "/api/user-dashboard/:id" {
		t.Errorf("Expected path '/api/user-dashboard/:id', got %s", agg.GetPath())
	}
}

func TestUserDashboardAggregator_Aggregate(t *testing.T) {
	agg := NewUserDashboardAggregator()

	req := httptest.NewRequest("GET", "/api/user-dashboard/123?id=123", nil)
	w := httptest.NewRecorder()

	err := agg.Aggregate(w, req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type: application/json, got %s", w.Header().Get("Content-Type"))
	}
}

func TestNewEcommerceAggregator(t *testing.T) {
	agg := NewEcommerceAggregator()

	if agg == nil {
		t.Fatal("Expected EcommerceAggregator, got nil")
	}

	if agg.GetName() != "ecommerce-catalog" {
		t.Errorf("Expected name 'ecommerce-catalog', got %s", agg.GetName())
	}
}

func TestEcommerceAggregator_Aggregate(t *testing.T) {
	agg := NewEcommerceAggregator()

	req := httptest.NewRequest("GET", "/api/catalog/electronics?category=electronics", nil)
	w := httptest.NewRecorder()

	err := agg.Aggregate(w, req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	if registry == nil {
		t.Fatal("Expected Registry, got nil")
	}

	if registry.aggregators == nil {
		t.Fatal("Expected aggregators map, got nil")
	}
}

func TestRegistry_RegisterGet(t *testing.T) {
	registry := NewRegistry()
	agg := NewUserDashboardAggregator()

	registry.Register(agg)

	retrievedAgg, exists := registry.Get("user-dashboard")
	if !exists {
		t.Fatal("Expected aggregator to exist")
	}

	if retrievedAgg != agg {
		t.Fatal("Expected same aggregator instance")
	}
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()
	agg1 := NewUserDashboardAggregator()
	agg2 := NewEcommerceAggregator()

	registry.Register(agg1)
	registry.Register(agg2)

	list := registry.List()

	if len(list) != 2 {
		t.Errorf("Expected 2 aggregators, got %d", len(list))
	}
}

func TestDefaultRegistry(t *testing.T) {
	registry := DefaultRegistry()

	if registry == nil {
		t.Fatal("Expected Registry, got nil")
	}

	// Should have default aggregators
	_, exists1 := registry.Get("user-dashboard")
	_, exists2 := registry.Get("ecommerce-catalog")

	if !exists1 || !exists2 {
		t.Fatal("Expected default aggregators to be registered")
	}
}
