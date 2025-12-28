package testgen

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/RichGod93/bffgen/internal/scaffolding"
)

// RouteConfig represents a single API route for testing
type RouteConfig struct {
	Path         string
	Method       string
	RequiresAuth bool
	Description  string
}

// TestConfig holds configuration for test generation
type TestConfig struct {
	ProjectName string
	Language    scaffolding.LanguageType
	Framework   string
	Routes      []RouteConfig
	OutputDir   string
}

// Generator generates test files
type Generator struct {
	config TestConfig
}

// NewGenerator creates a new test generator
func NewGenerator(config TestConfig) *Generator {
	return &Generator{config: config}
}

// GenerateIntegrationTests generates integration tests for the project
func (g *Generator) GenerateIntegrationTests() error {
	switch g.config.Language {
	case scaffolding.LanguageNodeExpress, scaffolding.LanguageNodeFastify:
		return g.generateNodeJSIntegrationTests()
	case scaffolding.LanguageGo:
		return g.generateGoIntegrationTests()
	case scaffolding.LanguagePythonFastAPI:
		return g.generatePythonIntegrationTests()
	default:
		return fmt.Errorf("unsupported language: %s", g.config.Language)
	}
}

// generateNodeJSIntegrationTests generates Jest integration tests for Node.js
func (g *Generator) generateNodeJSIntegrationTests() error {
	testDir := filepath.Join(g.config.OutputDir, "tests", "integration")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		return fmt.Errorf("failed to create test directory: %w", err)
	}

	// Generate test file for each route
	for _, route := range g.config.Routes {
		testFile := filepath.Join(testDir, fmt.Sprintf("%s.test.js", sanitizeFileName(route.Path)))

		tmpl := template.Must(template.New("integration").Parse(nodeJSIntegrationTemplate))

		f, err := os.Create(testFile)
		if err != nil {
			return fmt.Errorf("failed to create test file: %w", err)
		}
		defer f.Close()

		data := map[string]interface{}{
			"Route":       route,
			"ProjectName": g.config.ProjectName,
		}

		if err := tmpl.Execute(f, data); err != nil {
			return fmt.Errorf("failed to write test file: %w", err)
		}
	}

	return nil
}

// generateGoIntegrationTests generates table-driven integration tests for Go
func (g *Generator) generateGoIntegrationTests() error {
	testDir := filepath.Join(g.config.OutputDir, "tests", "integration")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		return fmt.Errorf("failed to create test directory: %w", err)
	}

	testFile := filepath.Join(testDir, "api_test.go")

	tmpl := template.Must(template.New("integration").Parse(goIntegrationTemplate))

	f, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("failed to create test file: %w", err)
	}
	defer f.Close()

	data := map[string]interface{}{
		"Routes":      g.config.Routes,
		"ProjectName": g.config.ProjectName,
		"Framework":   g.config.Framework,
	}

	return tmpl.Execute(f, data)
}

// generatePythonIntegrationTests generates pytest integration tests
func (g *Generator) generatePythonIntegrationTests() error {
	testDir := filepath.Join(g.config.OutputDir, "tests", "integration")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		return fmt.Errorf("failed to create test directory: %w", err)
	}

	// Create __init__.py
	initFile := filepath.Join(testDir, "__init__.py")
	if err := os.WriteFile(initFile, []byte(""), 0644); err != nil {
		return err
	}

	testFile := filepath.Join(testDir, "test_api.py")

	tmpl := template.Must(template.New("integration").Parse(pythonIntegrationTemplate))

	f, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("failed to create test file: %w", err)
	}
	defer f.Close()

	data := map[string]interface{}{
		"Routes":      g.config.Routes,
		"ProjectName": g.config.ProjectName,
	}

	return tmpl.Execute(f, data)
}

// sanitizeFileName converts a path like "/api/users" to "api-users"
func sanitizeFileName(path string) string {
	result := path
	result = filepath.Clean(result)
	result = filepath.Base(result)
	if result == "." || result == "/" {
		result = "root"
	}
	return result
}

// Template constants
const nodeJSIntegrationTemplate = `import request from 'supertest';
import app from '../../src/index.js';

describe('{{ .Route.Path }} Integration Tests', () => {
  {{ if .Route.RequiresAuth }}
  let authToken;

  beforeAll(() => {
    // TODO: Set up authentication token
    authToken = 'test-token';
  });
  {{ end }}

  describe('{{ .Route.Method }} {{ .Route.Path }}', () => {
    it('should {{ .Route.Description }}', async () => {
      const response = await request(app)
        .{{ .Route.Method | ToLower }}('{{ .Route.Path }}')
        {{ if .Route.RequiresAuth }}.set('Authorization', ` + "`Bearer ${authToken}`" + `){{ end }}
        .expect(200);

      expect(response.body).toBeDefined();
      // TODO: Add specific assertions
    });

    {{  if .Route.RequiresAuth }}
    it('should return 401 without authentication', async () => {
      await request(app)
        .{{ .Route.Method | ToLower }}('{{ .Route.Path }}')
        .expect(401);
    });
    {{ end }}

    it('should handle errors gracefully', async () => {
      // TODO: Test error scenarios
    });
  });
});
`

const goIntegrationTemplate = `package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIEndpoints(t *testing.T) {
	// Set up test server
	// TODO: Initialize your {{ .Framework }} server here

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{{ range .Routes }}
		{
			name:       "{{ .Description }}",
			method:     "{{ .Method }}",
			path:       "{{ .Path }}",
			wantStatus: http.StatusOK,
		},
		{{ end }}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			{{ if .RequiresAuth }}
			req.Header.Set("Authorization", "Bearer test-token")
			{{ end }}
			
			w := httptest.NewRecorder()
			// TODO: Call your handler here
			// handler.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
`

const pythonIntegrationTemplate = `import pytest
from fastapi.testclient import TestClient
from main import app

client = TestClient(app)

{{ if .RequiresAuth }}
@pytest.fixture
def auth_headers():
    """Create authentication headers for testing"""
    # TODO: Generate test JWT token
    return {"Authorization": "Bearer test-token"}
{{ end }}

{{ range .Routes }}
def test_{{ .Path | SanitizePython }}_{{ .Method | ToLower }}({{ if .RequiresAuth }}auth_headers{{ end }}):
    """Test {{ .Description }}"""
    response = client.{{ .Method | ToLower }}(
        "{{ .Path }}",
        {{ if .RequiresAuth }}headers=auth_headers{{ end }}
    )
    
    assert response.status_code == 200
    # TODO: Add specific assertions
    assert response.json() is not None

{{ if .RequiresAuth }}
def test_{{ .Path | SanitizePython }}_{{ .Method | ToLower }}_unauthorized():
    """Test {{ .Path }} without authentication"""
    response = client.{{ .Method | ToLower }}("{{ .Path }}")
    assert response.status_code == 401
{{ end }}
{{ end }}
`
