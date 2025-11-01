package templates

import (
	"strings"
	"testing"

	"github.com/RichGod93/bffgen/internal/scaffolding"
)

func TestNewTemplateLoader(t *testing.T) {
	tests := []struct {
		name     string
		langType scaffolding.LanguageType
		wantErr  bool
	}{
		{
			name:     "Express loader",
			langType: scaffolding.LanguageNodeExpress,
			wantErr:  false,
		},
		{
			name:     "Fastify loader",
			langType: scaffolding.LanguageNodeFastify,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewTemplateLoader(tt.langType)
			if loader == nil {
				t.Error("Expected non-nil loader")
				return
			}
			if loader.langType != tt.langType {
				t.Errorf("Expected langType %s, got %s", tt.langType, loader.langType)
			}
		})
	}
}

func TestLoadTemplate(t *testing.T) {
	tests := []struct {
		name      string
		langType  scaffolding.LanguageType
		framework string
		filename  string
		wantErr   bool
	}{
		{
			name:      "Express index.js template",
			langType:  scaffolding.LanguageNodeExpress,
			framework: "express",
			filename:  "index.js.tmpl",
			wantErr:   false,
		},
		{
			name:      "Fastify index.js template",
			langType:  scaffolding.LanguageNodeFastify,
			framework: "fastify",
			filename:  "index.js.tmpl",
			wantErr:   false,
		},
		{
			name:      "Express package.json template",
			langType:  scaffolding.LanguageNodeExpress,
			framework: "express",
			filename:  "package.json.tmpl",
			wantErr:   false,
		},
		{
			name:      "Fastify package.json template",
			langType:  scaffolding.LanguageNodeFastify,
			framework: "fastify",
			filename:  "package.json.tmpl",
			wantErr:   false,
		},
		{
			name:      "Common env template",
			langType:  scaffolding.LanguageNodeExpress,
			framework: "express",
			filename:  "env.tmpl",
			wantErr:   false,
		},
		{
			name:      "Common gitignore template",
			langType:  scaffolding.LanguageNodeExpress,
			framework: "express",
			filename:  "gitignore.tmpl",
			wantErr:   false,
		},
		{
			name:      "Nonexistent template",
			langType:  scaffolding.LanguageNodeExpress,
			framework: "express",
			filename:  "nonexistent.tmpl",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewTemplateLoader(tt.langType)
			content, err := loader.LoadTemplate(tt.framework, tt.filename)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(content) == 0 {
				t.Error("Expected non-empty template content")
			}
		})
	}
}

func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		name      string
		langType  scaffolding.LanguageType
		framework string
		filename  string
		data      *TemplateData
		wantErr   bool
		contains  []string
	}{
		{
			name:      "Express index.js with project name",
			langType:  scaffolding.LanguageNodeExpress,
			framework: "express",
			filename:  "index.js.tmpl",
			data: &TemplateData{
				ProjectName:   "test-app",
				CORSOrigins:   `["http://localhost:3000"]`,
				CORSOriginsJS: "['http://localhost:3000']",
			},
			wantErr:  false,
			contains: []string{"test-app", "express", "localhost:3000"},
		},
		{
			name:      "Fastify index.js with project name",
			langType:  scaffolding.LanguageNodeFastify,
			framework: "fastify",
			filename:  "index.js.tmpl",
			data: &TemplateData{
				ProjectName:   "test-fastify",
				CORSOrigins:   `["http://localhost:5000"]`,
				CORSOriginsJS: "['http://localhost:5000']",
			},
			wantErr:  false,
			contains: []string{"test-fastify", "fastify", "localhost:5000"},
		},
		{
			name:      "Express package.json",
			langType:  scaffolding.LanguageNodeExpress,
			framework: "express",
			filename:  "package.json.tmpl",
			data: &TemplateData{
				ProjectName: "my-bff",
			},
			wantErr:  false,
			contains: []string{"my-bff", "express", "nodemon"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewTemplateLoader(tt.langType)
			content, err := loader.RenderTemplate(tt.framework, tt.filename, tt.data)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(content) == 0 {
				t.Error("Expected non-empty rendered content")
			}

			// Check for expected strings
			for _, expected := range tt.contains {
				if !strings.Contains(content, expected) {
					t.Errorf("Expected rendered content to contain '%s'", expected)
				}
			}
		})
	}
}

func TestFormatCORSOriginsForJS(t *testing.T) {
	tests := []struct {
		name    string
		origins []string
		want    string
	}{
		{
			name:    "Single origin",
			origins: []string{"localhost:3000"},
			want:    "['http://localhost:3000']",
		},
		{
			name:    "Multiple origins",
			origins: []string{"localhost:3000", "localhost:3001"},
			want:    "['http://localhost:3000', 'http://localhost:3001']",
		},
		{
			name:    "Empty origins",
			origins: []string{},
			want:    "['http://localhost:3000']", // Default
		},
		{
			name:    "Origins with protocol",
			origins: []string{"http://example.com", "https://api.example.com"},
			want:    "['http://example.com', 'https://api.example.com']",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatCORSOriginsForJS(tt.origins)
			if got != tt.want {
				t.Errorf("FormatCORSOriginsForJS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatCORSOriginsForEnv(t *testing.T) {
	tests := []struct {
		name    string
		origins []string
		want    string
	}{
		{
			name:    "Single origin",
			origins: []string{"localhost:3000"},
			want:    "http://localhost:3000",
		},
		{
			name:    "Multiple origins",
			origins: []string{"localhost:3000", "localhost:3001"},
			want:    "http://localhost:3000,http://localhost:3001",
		},
		{
			name:    "Empty origins",
			origins: []string{},
			want:    "http://localhost:3000", // Default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatCORSOriginsForEnv(tt.origins)
			if got != tt.want {
				t.Errorf("FormatCORSOriginsForEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateBackendServicesEnv(t *testing.T) {
	tests := []struct {
		name     string
		services []BackendServiceData
		want     []string
	}{
		{
			name: "Single service",
			services: []BackendServiceData{
				{Name: "users", BaseURL: "http://localhost:5001/api"},
			},
			want: []string{"USERS_SERVICE_URL=http://localhost:5001/api"},
		},
		{
			name: "Multiple services",
			services: []BackendServiceData{
				{Name: "users", BaseURL: "http://localhost:5001/api"},
				{Name: "products", BaseURL: "http://localhost:5002/api"},
			},
			want: []string{
				"USERS_SERVICE_URL=http://localhost:5001/api",
				"PRODUCTS_SERVICE_URL=http://localhost:5002/api",
			},
		},
		{
			name:     "Empty services",
			services: []BackendServiceData{},
			want:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateBackendServicesEnv(tt.services)
			for _, expected := range tt.want {
				if !strings.Contains(got, expected) {
					t.Errorf("Expected env to contain '%s'", expected)
				}
			}
		})
	}
}

func TestGenerateExpressRoutes(t *testing.T) {
	services := []BackendServiceData{
		{Name: "users", BaseURL: "http://localhost:5001/api"},
	}

	routes := GenerateExpressRoutes(services)

	if !strings.Contains(routes, "app.get") {
		t.Error("Expected Express route declaration (app.get)")
	}

	if !strings.Contains(routes, "users") {
		t.Error("Expected service name 'users'")
	}

	if !strings.Contains(routes, "USERS_SERVICE_URL") {
		t.Error("Expected environment variable USERS_SERVICE_URL")
	}
}

func TestGenerateFastifyRoutes(t *testing.T) {
	services := []BackendServiceData{
		{Name: "products", BaseURL: "http://localhost:5002/api"},
	}

	routes := GenerateFastifyRoutes(services)

	if !strings.Contains(routes, "fastify.get") {
		t.Error("Expected Fastify route declaration (fastify.get)")
	}

	if !strings.Contains(routes, "products") {
		t.Error("Expected service name 'products'")
	}

	if !strings.Contains(routes, "PRODUCTS_SERVICE_URL") {
		t.Error("Expected environment variable PRODUCTS_SERVICE_URL")
	}
}

func TestGenerateBackendsJSON(t *testing.T) {
	tests := []struct {
		name     string
		services []BackendServiceData
		contains []string
	}{
		{
			name: "Single backend",
			services: []BackendServiceData{
				{Name: "users", BaseURL: "http://localhost:5001/api"},
			},
			contains: []string{
				`"name": "users"`,
				`"baseUrl": "http://localhost:5001/api"`,
				`"timeout": 30000`,
				`"healthCheck"`,
			},
		},
		{
			name: "Multiple backends",
			services: []BackendServiceData{
				{Name: "users", BaseURL: "http://localhost:5001/api"},
				{Name: "products", BaseURL: "http://localhost:5002/api"},
			},
			contains: []string{
				`"name": "users"`,
				`"name": "products"`,
			},
		},
		{
			name:     "Empty backends",
			services: []BackendServiceData{},
			contains: []string{"[]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateBackendsJSON(tt.services)

			for _, expected := range tt.contains {
				if !strings.Contains(got, expected) {
					t.Errorf("Expected JSON to contain '%s', got: %s", expected, got)
				}
			}
		})
	}
}

func TestMiddlewareTemplates(t *testing.T) {
	tests := []struct {
		name      string
		langType  scaffolding.LanguageType
		framework string
		filename  string
		contains  []string
	}{
		{
			name:      "Express auth middleware",
			langType:  scaffolding.LanguageNodeExpress,
			framework: "express",
			filename:  "middleware-auth.js.tmpl",
			contains:  []string{"authenticate", "jwt", "Bearer", "req.user"},
		},
		{
			name:      "Express error middleware",
			langType:  scaffolding.LanguageNodeExpress,
			framework: "express",
			filename:  "middleware-error.js.tmpl",
			contains:  []string{"errorHandler", "notFoundHandler", "asyncHandler"},
		},
		{
			name:      "Fastify auth plugin",
			langType:  scaffolding.LanguageNodeFastify,
			framework: "fastify",
			filename:  "middleware-auth.js.tmpl",
			contains:  []string{"authPlugin", "fastify.decorate", "authenticate", "request.user"},
		},
		{
			name:      "Fastify error plugin",
			langType:  scaffolding.LanguageNodeFastify,
			framework: "fastify",
			filename:  "middleware-error.js.tmpl",
			contains:  []string{"errorHandlerPlugin", "setErrorHandler", "setNotFoundHandler"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewTemplateLoader(tt.langType)
			content, err := loader.LoadTemplate(tt.framework, tt.filename)

			if err != nil {
				t.Fatalf("Failed to load template: %v", err)
			}

			for _, expected := range tt.contains {
				if !strings.Contains(content, expected) {
					t.Errorf("Expected template to contain '%s'", expected)
				}
			}
		})
	}
}

func TestCommonTemplates(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		contains []string
	}{
		{
			name:     "Environment template",
			filename: "env.tmpl",
			contains: []string{"NODE_ENV", "PORT", "JWT_SECRET", "CORS_ORIGINS"},
		},
		{
			name:     "Gitignore template",
			filename: "gitignore.tmpl",
			contains: []string{"node_modules", ".env", "*.log"},
		},
		{
			name:     "BFFGen config template",
			filename: "bffgen.config.json.tmpl",
			contains: []string{"project", "server", "backends", "features"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewTemplateLoader(scaffolding.LanguageNodeExpress)
			content, err := loader.LoadTemplate("express", tt.filename)

			if err != nil {
				t.Fatalf("Failed to load template: %v", err)
			}

			for _, expected := range tt.contains {
				if !strings.Contains(content, expected) {
					t.Errorf("Expected template to contain '%s'", expected)
				}
			}
		})
	}
}

func TestTemplateRendering(t *testing.T) {
	tests := []struct {
		name         string
		langType     scaffolding.LanguageType
		framework    string
		filename     string
		data         *TemplateData
		wantContains []string
		wantErr      bool
	}{
		{
			name:      "Express index with backend routes",
			langType:  scaffolding.LanguageNodeExpress,
			framework: "express",
			filename:  "index.js.tmpl",
			data: &TemplateData{
				ProjectName:   "test-app",
				CORSOrigins:   `["http://localhost:3000"]`,
				CORSOriginsJS: "['http://localhost:3000']",
				BackendRoutes: "// Test routes",
			},
			wantContains: []string{"test-app", "express", "// Test routes"},
			wantErr:      false,
		},
		{
			name:      "Fastify index with backend routes",
			langType:  scaffolding.LanguageNodeFastify,
			framework: "fastify",
			filename:  "index.js.tmpl",
			data: &TemplateData{
				ProjectName:   "test-fastify",
				CORSOrigins:   `["http://localhost:5000"]`,
				CORSOriginsJS: "['http://localhost:5000']",
				BackendRoutes: "// Fastify routes",
			},
			wantContains: []string{"test-fastify", "fastify", "// Fastify routes"},
			wantErr:      false,
		},
		{
			name:      "Package.json with project name",
			langType:  scaffolding.LanguageNodeExpress,
			framework: "express",
			filename:  "package.json.tmpl",
			data: &TemplateData{
				ProjectName: "my-project",
			},
			wantContains: []string{`"name": "my-project"`, `"express"`},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewTemplateLoader(tt.langType)
			content, err := loader.RenderTemplate(tt.framework, tt.filename, tt.data)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			for _, expected := range tt.wantContains {
				if !strings.Contains(content, expected) {
					t.Errorf("Expected rendered content to contain '%s'", expected)
				}
			}
		})
	}
}
