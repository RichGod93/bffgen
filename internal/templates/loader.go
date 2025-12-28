package templates

import (
	"bytes"
	"embed"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/RichGod93/bffgen/internal/scaffolding"
)

//go:embed node/**/*
var nodeTemplates embed.FS

// TemplateData holds data for template rendering
type TemplateData struct {
	ProjectName        string
	Runtime            string // nodejs-express, nodejs-fastify
	Framework          string // express, fastify
	CORSOrigins        string // JSON array format with double quotes: ["http://localhost:3000"]
	CORSOriginsJS      string // JS array format with single quotes: ['http://localhost:3000']
	CORSOriginsEnv     string // Comma-separated for .env: http://localhost:3000,http://localhost:3001
	BackendRoutes      string // Generated route code
	BackendServicesEnv string // Environment variables for backend services
	BackendsJSON       string // JSON array of backend services
	BackendServices    []BackendServiceData
}

// BackendServiceData represents a backend service for templates
type BackendServiceData struct {
	Name    string
	BaseURL string
	Port    int
}

// ServiceTemplateData holds data for service template rendering
type ServiceTemplateData struct {
	ServiceName       string
	ServiceNamePascal string
	BaseURL           string
	EnvKey            string
	Endpoints         []EndpointData
}

// ControllerTemplateData holds data for controller template rendering
type ControllerTemplateData struct {
	ServiceName       string
	ServiceNamePascal string
	Endpoints         []EndpointData
}

// EndpointData represents an API endpoint
type EndpointData struct {
	Name              string
	Path              string
	Method            string
	BackendPath       string
	ExposeAs          string
	RequiresAuth      bool
	HandlerName       string
	HandlerNamePascal string
	Description       string
}

// TemplateLoader handles loading and rendering templates
type TemplateLoader struct {
	langType scaffolding.LanguageType
	fs       embed.FS
}

// NewTemplateLoader creates a new template loader
func NewTemplateLoader(langType scaffolding.LanguageType) *TemplateLoader {
	return &TemplateLoader{
		langType: langType,
		fs:       nodeTemplates,
	}
}

// LoadTemplate loads a template file
func (tl *TemplateLoader) LoadTemplate(framework, filename string) (string, error) {
	var path string

	// Common templates (shared between Express, Fastify, Apollo, and Yoga)
	commonFiles := []string{
		"env.tmpl", "env.test.tmpl", "gitignore.tmpl", "bffgen.config.json.tmpl",
		"service-base.js.tmpl", "service-template.js.tmpl",
		"jest.config.js.tmpl", "setup-tests.js.tmpl",
		"test-fixtures.js.tmpl", "test-integration.template.js.tmpl",
		"swagger-config.js.tmpl", "logger.js.tmpl",
		// New aggregation utilities
		"aggregator.js.tmpl", "cache-manager.js.tmpl", "circuit-breaker.js.tmpl",
		"response-transformer.js.tmpl", "request-batcher.js.tmpl", "field-selector.js.tmpl",
		// GraphQL utilities
		"graphql/rest-datasource.js.tmpl", "graphql/schema-stitching.js.tmpl",
		// Docker and scripts
		"docker-compose.yml.tmpl", "scripts/clear-cache.js.tmpl",
		// Health and shutdown
		"health.js.tmpl", "graceful-shutdown.js.tmpl",
	}

	isCommon := false
	for _, cf := range commonFiles {
		if filename == cf {
			isCommon = true
			break
		}
	}

	if isCommon {
		path = filepath.Join("node", "common", filename)
	} else if strings.HasPrefix(filename, "examples/") || strings.HasPrefix(filename, "scripts/") {
		// Examples and scripts are always in common
		path = filepath.Join("node", "common", filename)
	} else {
		// Framework-specific templates
		switch tl.langType {
		case scaffolding.LanguageNodeExpress:
			path = filepath.Join("node", "express", filename)
		case scaffolding.LanguageNodeFastify:
			path = filepath.Join("node", "fastify", filename)
		case scaffolding.LanguageNodeApollo:
			path = filepath.Join("node", "apollo", filename)
		case scaffolding.LanguageNodeYoga:
			path = filepath.Join("node", "yoga", filename)
		default:
			return "", fmt.Errorf("unsupported language type: %s", tl.langType)
		}
	}

	content, err := tl.fs.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read template %s: %w", path, err)
	}

	return string(content), nil
}

// RenderTemplate renders a template with data
func (tl *TemplateLoader) RenderTemplate(framework, filename string, data *TemplateData) (string, error) {
	tmplContent, err := tl.LoadTemplate(framework, filename)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New(filename).Parse(tmplContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", filename, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", filename, err)
	}

	return buf.String(), nil
}

// FormatCORSOriginsForJS formats CORS origins for JavaScript array
func FormatCORSOriginsForJS(origins []string) string {
	if len(origins) == 0 {
		return "['http://localhost:3000']"
	}

	formatted := make([]string, len(origins))
	for i, origin := range origins {
		// Ensure proper protocol
		if !strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://") {
			origin = "http://" + origin
		}
		formatted[i] = fmt.Sprintf("'%s'", origin)
	}

	return "[" + strings.Join(formatted, ", ") + "]"
}

// FormatCORSOriginsForJSON formats CORS origins for JSON (double quotes)
func FormatCORSOriginsForJSON(origins []string) string {
	if len(origins) == 0 {
		return `["http://localhost:3000"]`
	}

	formatted := make([]string, len(origins))
	for i, origin := range origins {
		// Ensure proper protocol
		if !strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://") {
			origin = "http://" + origin
		}
		formatted[i] = fmt.Sprintf(`"%s"`, origin)
	}

	return "[" + strings.Join(formatted, ", ") + "]"
}

// FormatCORSOriginsForEnv formats CORS origins for .env file
func FormatCORSOriginsForEnv(origins []string) string {
	if len(origins) == 0 {
		return "http://localhost:3000"
	}

	formatted := make([]string, len(origins))
	for i, origin := range origins {
		// Ensure proper protocol
		if !strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://") {
			origin = "http://" + origin
		}
		formatted[i] = origin
	}

	return strings.Join(formatted, ",")
}

// GenerateBackendServicesEnv generates environment variables for backend services
func GenerateBackendServicesEnv(services []BackendServiceData) string {
	if len(services) == 0 {
		return "# No backend services configured"
	}

	var lines []string
	for _, service := range services {
		envKey := strings.ToUpper(service.Name) + "_SERVICE_URL"
		lines = append(lines, fmt.Sprintf("%s=%s", envKey, service.BaseURL))
	}

	return strings.Join(lines, "\n")
}

// GenerateExpressRoutes generates Express.js route code
func GenerateExpressRoutes(services []BackendServiceData) string {
	if len(services) == 0 {
		return `
// TODO: Add your BFF routes here
// Example:
// app.get('/api/users', async (req, res) => {
//   try {
//     const response = await fetch(process.env.USERS_SERVICE_URL + '/users');
//     const data = await response.json();
//     res.json(data);
//   } catch (error) {
//     res.status(500).json({ error: 'Failed to fetch users' });
//   }
// });`
	}

	var routes []string
	routes = append(routes, "\n// Generated backend service routes")

	for _, service := range services {
		envKey := strings.ToUpper(service.Name) + "_SERVICE_URL"
		routes = append(routes, fmt.Sprintf(`
// %s service routes
app.get('/api/%s', async (req, res) => {
  try {
    const baseURL = process.env.%s || '%s';
    const response = await fetch(baseURL + '/%s');
    if (!response.ok) throw new Error('Backend service error');
    const data = await response.json();
    res.json(data);
  } catch (error) {
    console.error('%s service error:', error);
    res.status(500).json({ error: 'Failed to fetch %s data' });
  }
});`, service.Name, service.Name, envKey, service.BaseURL, service.Name, service.Name, service.Name))
	}

	return strings.Join(routes, "\n")
}

// GenerateFastifyRoutes generates Fastify route code
func GenerateFastifyRoutes(services []BackendServiceData) string {
	if len(services) == 0 {
		return `
    // TODO: Add your BFF routes here
    // Example:
    // fastify.get('/api/users', async (request, reply) => {
    //   try {
    //     const response = await fetch(process.env.USERS_SERVICE_URL + '/users');
    //     const data = await response.json();
    //     return data;
    //   } catch (error) {
    //     reply.status(500);
    //     return { error: 'Failed to fetch users' };
    //   }
    // });`
	}

	var routes []string
	routes = append(routes, "\n    // Generated backend service routes")

	for _, service := range services {
		envKey := strings.ToUpper(service.Name) + "_SERVICE_URL"
		routes = append(routes, fmt.Sprintf(`
    // %s service routes
    fastify.get('/api/%s', async (request, reply) => {
      try {
        const baseURL = process.env.%s || '%s';
        const response = await fetch(baseURL + '/%s');
        if (!response.ok) throw new Error('Backend service error');
        const data = await response.json();
        return data;
      } catch (error) {
        fastify.log.error('%s service error:', error);
        reply.status(500);
        return { error: 'Failed to fetch %s data' };
      }
    });`, service.Name, service.Name, envKey, service.BaseURL, service.Name, service.Name, service.Name))
	}

	return strings.Join(routes, "\n")
}

// GenerateBackendsJSON generates JSON array for bffgen.config.json
func GenerateBackendsJSON(services []BackendServiceData) string {
	if len(services) == 0 {
		return "[]"
	}

	var backends []string
	for _, service := range services {
		backend := fmt.Sprintf(`    {
      "name": "%s",
      "baseUrl": "%s",
      "timeout": 30000,
      "retries": 3,
      "healthCheck": {
        "enabled": true,
        "path": "/health",
        "interval": 60000
      }
    }`, service.Name, service.BaseURL)
		backends = append(backends, backend)
	}

	return "[\n" + strings.Join(backends, ",\n") + "\n  ]"
}

// ToPascalCase converts a string to PascalCase
func ToPascalCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_' || r == ' '
	})

	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, "")
}

// ToCamelCase converts a string to camelCase
func ToCamelCase(s string) string {
	pascal := ToPascalCase(s)
	if len(pascal) > 0 {
		return strings.ToLower(string(pascal[0])) + pascal[1:]
	}
	return pascal
}
