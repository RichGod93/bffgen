package routes

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
)

// Service represents a backend service configuration
type Service struct {
	BaseURL   string     `yaml:"baseUrl"`
	Endpoints []Endpoint `yaml:"endpoints"`
}

// Endpoint represents a single API endpoint
type Endpoint struct {
	Name     string `yaml:"name"`
	Path     string `yaml:"path"`
	Method   string `yaml:"method"`
	ExposeAs string `yaml:"exposeAs"`
}

// SetupProxyRoutes configures reverse proxy routes for all services
func SetupProxyRoutes(r chi.Router, services map[string]Service) {
	for _, service := range services {
		baseURL, err := url.Parse(service.BaseURL)
		if err != nil {
			// Skip invalid URLs silently in production, log in verbose mode if needed
			continue
		}

		for _, endpoint := range service.Endpoints {
			// Create reverse proxy
			proxy := httputil.NewSingleHostReverseProxy(baseURL)

			// Modify the request
			proxy.Director = func(req *http.Request) {
				req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
				req.Header.Set("X-Real-IP", req.RemoteAddr)
				req.URL.Scheme = baseURL.Scheme
				req.URL.Host = baseURL.Host
				req.URL.Path = endpoint.Path
			}

			// Add route based on method
			switch strings.ToUpper(endpoint.Method) {
			case "GET":
				r.Get(endpoint.ExposeAs, proxy.ServeHTTP)
			case "POST":
				r.Post(endpoint.ExposeAs, proxy.ServeHTTP)
			case "PUT":
				r.Put(endpoint.ExposeAs, proxy.ServeHTTP)
			case "DELETE":
				r.Delete(endpoint.ExposeAs, proxy.ServeHTTP)
			case "PATCH":
				r.Patch(endpoint.ExposeAs, proxy.ServeHTTP)
			default:
				// Skip unsupported methods silently
				// Note: Method %s not supported for endpoint %s", endpoint.Method, endpoint.Name
			}
		}
	}
}
