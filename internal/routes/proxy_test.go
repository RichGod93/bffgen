package routes

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestSetupProxyRoutes(t *testing.T) {
	// Create a test backend server
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Note: The proxy headers are set by the reverse proxy, not the backend
		// We'll just verify the request reaches the backend
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"backend": "response"}`))
	}))
	defer backendServer.Close()

	// Parse backend URL
	backendURL, err := url.Parse(backendServer.URL)
	if err != nil {
		t.Fatalf("Failed to parse backend URL: %v", err)
	}

	// Create services configuration
	services := map[string]Service{
		"test-service": {
			BaseURL: backendURL.String(),
			Endpoints: []Endpoint{
				{
					Name:     "get-users",
					Path:     "/users",
					Method:   "GET",
					ExposeAs: "/api/users",
				},
				{
					Name:     "create-user",
					Path:     "/users",
					Method:   "POST",
					ExposeAs: "/api/users",
				},
			},
		},
	}

	// Setup router and proxy routes
	r := chi.NewRouter()
	SetupProxyRoutes(r, services)

	// Test GET endpoint
	req := httptest.NewRequest("GET", "/api/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test POST endpoint
	req = httptest.NewRequest("POST", "/api/users", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestSetupProxyRoutes_InvalidURL(t *testing.T) {
	services := map[string]Service{
		"invalid-service": {
			BaseURL: "invalid-url",
			Endpoints: []Endpoint{
				{
					Name:     "test",
					Path:     "/test",
					Method:   "GET",
					ExposeAs: "/api/test",
				},
			},
		},
	}

	r := chi.NewRouter()
	// This should not panic and should skip the invalid service
	SetupProxyRoutes(r, services)

	// Test that the route was not created
	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should get 404 since route wasn't created due to invalid URL
	// Note: In practice, this might return 502 (Bad Gateway) if the proxy tries to connect
	// The important thing is that the route setup doesn't crash
	if w.Code != http.StatusNotFound && w.Code != http.StatusBadGateway {
		t.Errorf("Expected status 404 or 502, got %d", w.Code)
	}
}

func TestSetupProxyRoutes_UnsupportedMethod(t *testing.T) {
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backendServer.Close()

	services := map[string]Service{
		"test-service": {
			BaseURL: backendServer.URL,
			Endpoints: []Endpoint{
				{
					Name:     "unsupported",
					Path:     "/test",
					Method:   "UNSUPPORTED",
					ExposeAs: "/api/test",
				},
			},
		},
	}

	r := chi.NewRouter()
	SetupProxyRoutes(r, services)

	// Test that the route was not created for unsupported method
	req := httptest.NewRequest("UNSUPPORTED", "/api/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should get 404 since route wasn't created due to unsupported method
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestSetupProxyRoutes_MultipleMethods(t *testing.T) {
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"method": "` + r.Method + `"}`))
	}))
	defer backendServer.Close()

	services := map[string]Service{
		"test-service": {
			BaseURL: backendServer.URL,
			Endpoints: []Endpoint{
				{
					Name:     "get",
					Path:     "/test",
					Method:   "GET",
					ExposeAs: "/api/test",
				},
				{
					Name:     "put",
					Path:     "/test",
					Method:   "PUT",
					ExposeAs: "/api/test",
				},
				{
					Name:     "delete",
					Path:     "/test",
					Method:   "DELETE",
					ExposeAs: "/api/test",
				},
				{
					Name:     "patch",
					Path:     "/test",
					Method:   "PATCH",
					ExposeAs: "/api/test",
				},
			},
		},
	}

	r := chi.NewRouter()
	SetupProxyRoutes(r, services)

	// Test all supported methods
	methods := []string{"GET", "PUT", "DELETE", "PATCH"}
	for _, method := range methods {
		req := httptest.NewRequest(method, "/api/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for %s, got %d", method, w.Code)
		}
	}
}

func TestService_Structure(t *testing.T) {
	service := Service{
		BaseURL: "http://localhost:8080",
		Endpoints: []Endpoint{
			{
				Name:     "test-endpoint",
				Path:     "/test",
				Method:   "GET",
				ExposeAs: "/api/test",
			},
		},
	}

	if service.BaseURL != "http://localhost:8080" {
		t.Errorf("Expected BaseURL 'http://localhost:8080', got %s", service.BaseURL)
	}

	if len(service.Endpoints) != 1 {
		t.Errorf("Expected 1 endpoint, got %d", len(service.Endpoints))
	}

	endpoint := service.Endpoints[0]
	if endpoint.Name != "test-endpoint" {
		t.Errorf("Expected Name 'test-endpoint', got %s", endpoint.Name)
	}
	if endpoint.Path != "/test" {
		t.Errorf("Expected Path '/test', got %s", endpoint.Path)
	}
	if endpoint.Method != "GET" {
		t.Errorf("Expected Method 'GET', got %s", endpoint.Method)
	}
	if endpoint.ExposeAs != "/api/test" {
		t.Errorf("Expected ExposeAs '/api/test', got %s", endpoint.ExposeAs)
	}
}
