package commands

import (
	"testing"
)

func TestGenerateCORSConfig(t *testing.T) {
	tests := []struct {
		name     string
		origins  []string
		framework string
		expected string
	}{
		{
			name:     "Chi framework with single origin",
			origins:  []string{"http://localhost:3000"},
			framework: "chi",
			expected: `r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))`,
		},
		{
			name:     "Echo framework with multiple origins",
			origins:  []string{"http://localhost:3000", "https://example.com"},
			framework: "echo",
			expected: `e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000", "https://example.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposeHeaders:    []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))`,
		},
		{
			name:     "Fiber framework with single origin",
			origins:  []string{"http://localhost:3000"},
			framework: "fiber",
			expected: `app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Accept,Authorization,Content-Type,X-CSRF-Token",
		ExposeHeaders:    "Link",
		AllowCredentials: true,
		MaxAge:           300,
	}))`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateCORSConfig(tt.origins, tt.framework)
			if result != tt.expected {
				t.Errorf("generateCORSConfig() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGenerateGoMod(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
		framework   string
		expected    string
	}{
		{
			name:        "Chi framework",
			projectName: "test-project",
			framework:   "chi",
			expected: `module test-project

go 1.21

require (
	github.com/go-chi/chi/v5 v5.2.3
	github.com/go-chi/cors v1.2.2
	github.com/golang-jwt/jwt/v5 v5.2.1
	gopkg.in/yaml.v3 v3.0.1
)`,
		},
		{
			name:        "Echo framework",
			projectName: "test-project",
			framework:   "echo",
			expected: `module test-project

go 1.21

require (
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/labstack/echo/v4 v4.11.4
	gopkg.in/yaml.v3 v3.0.1
)`,
		},
		{
			name:        "Fiber framework",
			projectName: "test-project",
			framework:   "fiber",
			expected: `module test-project

go 1.21

require (
	github.com/gofiber/fiber/v2 v2.52.9
	github.com/golang-jwt/jwt/v5 v5.2.1
	gopkg.in/yaml.v3 v3.0.1
)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateGoMod(tt.projectName, tt.framework)
			if result != tt.expected {
				t.Errorf("generateGoMod() = %v, want %v", result, tt.expected)
			}
		})
	}
}
