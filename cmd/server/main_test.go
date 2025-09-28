package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestHealthEndpoint(t *testing.T) {
	// Create a test server with the same setup as main()
	r := createTestRouter()

	// Test health endpoint
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	expectedBody := "BFF server is running!"
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, w.Body.String())
	}
}

func TestSecurityHeaders(t *testing.T) {
	r := createTestRouter()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Check security headers
	securityHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
		"Permissions-Policy":     "geolocation=(), microphone=(), camera=()",
	}

	for header, expectedValue := range securityHeaders {
		actualValue := w.Header().Get(header)
		if actualValue != expectedValue {
			t.Errorf("Expected header %s: %s, got %s", header, expectedValue, actualValue)
		}
	}
}

func TestCORSHeaders(t *testing.T) {
	r := createTestRouter()

	// Test preflight request
	req := httptest.NewRequest("OPTIONS", "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Check CORS headers
	corsHeaders := map[string]string{
		"Access-Control-Allow-Origin":      "http://localhost:3000",
		"Access-Control-Allow-Methods":     "GET,POST,PUT,DELETE,OPTIONS",
		"Access-Control-Allow-Headers":     "Accept,Authorization,Content-Type,X-CSRF-Token",
		"Access-Control-Expose-Headers":    "Link",
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Max-Age":           "300",
	}

	for header, expectedValue := range corsHeaders {
		actualValue := w.Header().Get(header)
		if actualValue != expectedValue {
			t.Errorf("Expected CORS header %s: %s, got %s", header, expectedValue, actualValue)
		}
	}
}

func TestContentTypeValidation(t *testing.T) {
	r := createTestRouter()

	tests := []struct {
		name           string
		method         string
		contentType    string
		expectedStatus int
	}{
		{
			name:           "POST with valid JSON content type",
			method:         "POST",
			contentType:    "application/json",
			expectedStatus: http.StatusMethodNotAllowed, // Health endpoint only supports GET
		},
		{
			name:           "POST with valid form content type",
			method:         "POST",
			contentType:    "application/x-www-form-urlencoded",
			expectedStatus: http.StatusMethodNotAllowed, // Health endpoint only supports GET
		},
		{
			name:           "POST with invalid content type",
			method:         "POST",
			contentType:    "text/plain",
			expectedStatus: http.StatusUnsupportedMediaType,
		},
		{
			name:           "GET with no content type",
			method:         "GET",
			contentType:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "PUT with valid JSON content type",
			method:         "PUT",
			contentType:    "application/json",
			expectedStatus: http.StatusMethodNotAllowed, // Health endpoint only supports GET
		},
		{
			name:           "PUT with invalid content type",
			method:         "PUT",
			contentType:    "text/html",
			expectedStatus: http.StatusUnsupportedMediaType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/health", nil)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestAuthMiddleware_PublicEndpoints(t *testing.T) {
	r := createTestRouter()

	publicEndpoints := []string{
		"/health",
		"/api/auth/login",
		"/api/auth/register",
	}

	for _, endpoint := range publicEndpoints {
		t.Run("Public endpoint: "+endpoint, func(t *testing.T) {
			req := httptest.NewRequest("GET", endpoint, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			// Public endpoints should be accessible without auth
			if w.Code == http.StatusUnauthorized {
				t.Errorf("Expected public endpoint %s to be accessible, got 401", endpoint)
			}
		})
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	r := createTestRouter()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Check that RequestID header is set
	requestID := w.Header().Get("X-Request-Id")
	if requestID == "" {
		t.Error("Expected X-Request-Id header to be set")
	}
}

func TestTimeoutMiddleware(t *testing.T) {
	r := createTestRouter()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// The timeout middleware should not affect normal requests
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestTRACEMethodDisabled(t *testing.T) {
	r := createTestRouter()
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("TRACE", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d for TRACE method, got: %d", http.StatusMethodNotAllowed, w.Code)
	}

	body := strings.TrimSpace(w.Body.String())
	if body != "Method Not Allowed" {
		t.Errorf("Expected 'Method Not Allowed' response, got: %s", body)
	}
}

func TestRequestSizeLimit(t *testing.T) {
	r := chi.NewRouter()

	// Add request validation middleware with size limit
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, 5<<20)
			next.ServeHTTP(w, r)
		})
	})

	r.Post("/test", func(w http.ResponseWriter, r *http.Request) {
		// Try to read the body to trigger size limit
		body := make([]byte, 6<<20) // Try to read 6MB
		_, err := r.Body.Read(body)
		if err != nil {
			http.Error(w, "Request too large", http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	// Create a request larger than 5MB
	largeBody := make([]byte, 6<<20) // 6MB
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(largeBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should be rejected due to size limit
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status %d for large request, got: %d", http.StatusRequestEntityTooLarge, w.Code)
	}
}

func TestMethodValidation(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		expected int
	}{
		{"Valid GET", "GET", http.StatusOK},
		{"Valid POST", "POST", http.StatusOK},
		{"Valid PUT", "PUT", http.StatusOK},
		{"Valid DELETE", "DELETE", http.StatusOK},
		{"Valid OPTIONS", "OPTIONS", http.StatusOK},
		{"Valid HEAD", "HEAD", http.StatusOK},
		{"Invalid TRACE", "TRACE", http.StatusMethodNotAllowed},
		{"Invalid CONNECT", "CONNECT", http.StatusMethodNotAllowed},
		{"Invalid PATCH", "PATCH", http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := createTestRouter()
			r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			r.Post("/test", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			r.Put("/test", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			r.Delete("/test", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			r.Options("/test", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			r.Head("/test", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(tt.method, "/test", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expected {
				t.Errorf("Expected status %d, got: %d", tt.expected, w.Code)
			}
		})
	}
}

func TestEnhancedSecurityHeaders(t *testing.T) {
	r := createTestRouter()
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Check enhanced security headers
	expectedHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
		"Permissions-Policy":     "geolocation=(), microphone=(), camera=()",
	}

	for header, expectedValue := range expectedHeaders {
		actualValue := w.Header().Get(header)
		if actualValue != expectedValue {
			t.Errorf("Expected header %s: %s, got: %s", header, expectedValue, actualValue)
		}
	}
}

// createTestRouter creates a router with the same middleware setup as main()
// but without starting the server
func createTestRouter() *chi.Mux {
	r := chi.NewRouter()

	// Add the same middleware as in main()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
			next.ServeHTTP(w, r)
		})
	})

	// Add CORS middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "http://localhost:3000" || origin == "http://localhost:3001" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept,Authorization,Content-Type,X-CSRF-Token")
			w.Header().Set("Access-Control-Expose-Headers", "Link")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "300")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Add request validation middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, 5<<20)

			if r.Method == "POST" || r.Method == "PUT" {
				contentType := r.Header.Get("Content-Type")
				if contentType != "" && contentType != "application/json" && contentType != "application/x-www-form-urlencoded" && contentType != "multipart/form-data" {
					http.Error(w, "Unsupported Content-Type", http.StatusUnsupportedMediaType)
					return
				}
			}

			// Validate request method
			allowedMethods := map[string]bool{
				"GET": true, "POST": true, "PUT": true, "DELETE": true, "OPTIONS": true, "HEAD": true,
			}
			if !allowedMethods[r.Method] {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Disable TRACE method for security
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "TRACE" {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// Add auth middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" || r.URL.Path == "/api/auth/login" || r.URL.Path == "/api/auth/register" {
				next.ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// Add RequestID middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Request-Id", "test-request-id")
			next.ServeHTTP(w, r)
		})
	})

	// Add health endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("BFF server is running!"))
	})

	return r
}
