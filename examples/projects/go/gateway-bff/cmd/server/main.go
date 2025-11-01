// bffgen:begin
package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "BFF server is running!")
	})

	// Generated proxy routes
	// users service routes
	r.Get("/api/users", createProxyHandler("http://localhost:4000/api", "/users"))
	r.Get("/api/users/{id}", createProxyHandler("http://localhost:4000/api", "/users/{id}"))
	r.Post("/api/users", createProxyHandler("http://localhost:4000/api", "/users"))
	r.Put("/api/users/{id}", createProxyHandler("http://localhost:4000/api", "/users/{id}"))
	r.Delete("/api/users/{id}", createProxyHandler("http://localhost:4000/api", "/users/{id}"))

	// analytics service routes
	r.Get("/api/analytics/metrics", createProxyHandler("http://localhost:4001/api", "/metrics"))
	r.Get("/api/analytics/events", createProxyHandler("http://localhost:4001/api", "/events"))
	r.Post("/api/analytics/events", createProxyHandler("http://localhost:4001/api", "/events"))

	// notifications service routes
	r.Get("/api/notifications", createProxyHandler("http://localhost:4002/api", "/notifications"))
	r.Get("/api/notifications/{id}", createProxyHandler("http://localhost:4002/api", "/notifications/{id}"))
	r.Post("/api/notifications/{id}/read", createProxyHandler("http://localhost:4002/api", "/notifications/{id}/read"))
	r.Post("/api/notifications/read-all", createProxyHandler("http://localhost:4002/api", "/notifications/read-all"))


	fmt.Println("🚀 BFF server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// createProxyHandler creates a reverse proxy handler for the given backend URL and path
func createProxyHandler(backendURL, backendPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the backend URL
		target, err := url.Parse(backendURL)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid backend URL: %v", err), http.StatusInternalServerError)
			return
		}

		// Create reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(target)
		
		// Configure proxy behavior
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Proxy error: %v", err)
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
		}

		// Modify the request to use the backend path
		originalPath := r.URL.Path
		r.URL.Path = backendPath
		r.URL.Host = target.Host
		r.URL.Scheme = target.Scheme
		r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
		r.Host = target.Host

		// Log the proxy request
		log.Printf("Proxying %s %s -> %s%s", r.Method, originalPath, backendURL, backendPath)

		// Serve the proxy request
		proxy.ServeHTTP(w, r)
	}
}
// bffgen:end