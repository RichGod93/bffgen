package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/RichGod93/bffgen/internal/scaffolding"
	"github.com/RichGod93/bffgen/internal/types"
	"github.com/RichGod93/bffgen/internal/utils"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Go code for routes from config",
	Long:  `Generate Go code for routes from bff.config.yaml configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := generate(); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating code: %v\n", err)
			os.Exit(1)
		}
	},
}

var (
	checkMode bool
	dryRun    bool
	verbose   bool
)

func init() {
	generateCmd.Flags().BoolVar(&checkMode, "check", false, "Check mode: show what would be changed without making changes")
	generateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Dry run: show what would be changed without making changes")
	generateCmd.Flags().BoolVar(&verbose, "verbose", false, "Verbose output")
}

func generate() error {
	LogVerbose("Starting code generation")
	
	// Detect project type
	projectType := detectProjectType()
	
	if projectType == "unknown" {
		fmt.Println("❌ No BFF project found in current directory")
		fmt.Println("💡 Run 'bffgen init <project-name>' first or navigate to a BFF project directory")
		return fmt.Errorf("no project configuration found")
	}
	
	// Handle based on project type
	if projectType == "nodejs" {
		return generateNodeJS()
	}
	
	// Default: Go project
	return generateGo()
}

// generateGo generates code for Go projects
func generateGo() error {
	LogVerbose("Starting code generation from bff.config.yaml")
	
	// Create generator with regeneration-safe capabilities
	generator := scaffolding.NewGenerator()
	generator.SetCheckMode(checkMode)
	generator.SetDryRun(dryRun)
	generator.SetVerbose(verbose)

	if checkMode {
		fmt.Println("🔍 Check mode: Analyzing what would be changed")
	} else if dryRun {
		fmt.Println("🔍 Dry run: Showing what would be changed")
	} else {
		fmt.Println("🔧 Generating Go code from bff.config.yaml")
	}
	fmt.Println()

	// Check if config file exists
	if _, err := os.Stat("bff.config.yaml"); os.IsNotExist(err) {
		fmt.Println("❌ bff.config.yaml not found in current directory")
		fmt.Println("💡 Run 'bffgen init <project-name>' first or navigate to a BFF project directory")
		return fmt.Errorf("config file not found")
	}

	// Load configuration
	config, err := utils.LoadConfig("bff.config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(config.Services) == 0 {
		fmt.Println("⚠️  No services configured in bff.config.yaml")
		fmt.Println("💡 Add services using 'bffgen add-route' or 'bffgen add-template'")
		return nil
	}

	LogVerbose("Found %d services to generate", len(config.Services))

	// Generate main.go with routes using regeneration-safe scaffolding
	if err := generateMainGoWithScaffolding(config, generator); err != nil {
		return fmt.Errorf("failed to generate main.go: %w", err)
	}

	// Generate server entry point using regeneration-safe scaffolding
	if err := generateServerMainWithScaffolding(config, generator); err != nil {
		return fmt.Errorf("failed to generate server main: %w", err)
	}

	if !checkMode && !dryRun {
		fmt.Println("✅ Code generation completed!")
		fmt.Println("📁 Updated files:")
		fmt.Println("   - main.go (with proxy routes)")
		fmt.Println("   - cmd/server/main.go (server entry point)")
		fmt.Println()
		fmt.Println("🚀 Run 'go run main.go' to start your BFF server")
		fmt.Println()
		fmt.Println("📮 Generate Postman collection: bffgen postman")
		fmt.Println("   This creates a ready-to-import collection for testing your BFF endpoints")
	}

	LogVerbose("Code generation completed successfully")

	return nil
}

func generateMainGo(config *types.BFFConfig) error {
	// Check if main.go already exists (created by init)
	if _, err := os.Stat("main.go"); err == nil {
		// main.go exists, just add proxy routes to it
		return addProxyRoutesToMainGo(config)
	}

	// main.go doesn't exist, create a basic one
	return createBasicMainGo(config)
}

func addProxyRoutesToMainGo(config *types.BFFConfig) error {
	// Read existing main.go
	content, err := os.ReadFile("main.go")
	if err != nil {
		return fmt.Errorf("failed to read main.go: %w", err)
	}

	contentStr := string(content)

	// Check if proxy routes already exist
	if strings.Contains(contentStr, "// Generated proxy routes") {
		LogVerbose("Proxy routes already exist in main.go, skipping")
		return nil
	}

	// Find the TODO comment and replace it with actual proxy routes
	todoComment := "// TODO: Add your aggregated routes here\n\t// Run 'bffgen add-route' or 'bffgen add-template' to add routes\n\t// Then run 'bffgen generate' to generate the code"

	// Generate proxy routes
	proxyRoutes := generateProxyRoutesCode(config)

	// Replace TODO comment with actual routes
	newContent := strings.Replace(contentStr, todoComment, proxyRoutes, 1)

	// Add createProxyHandler function if it doesn't exist
	if !strings.Contains(newContent, "func createProxyHandler") {
		// Find the end of main() function and insert createProxyHandler after it
		mainEndPattern := "\n\tfmt.Println(\"🚀 BFF server starting on :8080\")\n\tlog.Fatal(http.ListenAndServe(\":8080\", r))\n}\n"
		proxyHandlerFunc := generateProxyHandlerFunction()

		// Insert createProxyHandler after the main function ends
		newContent = strings.Replace(newContent, mainEndPattern, mainEndPattern+"\n"+proxyHandlerFunc+"\n", 1)
	}

	// Write updated main.go
	if err := os.WriteFile("main.go", []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write main.go: %w", err)
	}

	LogVerbose("Added proxy routes to existing main.go")
	return nil
}

func generateProxyRoutesCode(config *types.BFFConfig) string {
	var routes strings.Builder
	routes.WriteString("\t// Generated proxy routes\n")

	for serviceName, service := range config.Services {
		routes.WriteString(fmt.Sprintf("\t// %s service routes\n", serviceName))
		for _, endpoint := range service.Endpoints {
			method := chiMethod(endpoint.Method)
			routes.WriteString(fmt.Sprintf("\tr.%s(\"%s\", createProxyHandler(\"%s\", \"%s\"))\n",
				method, endpoint.ExposeAs, service.BaseURL, endpoint.Path))
		}
		routes.WriteString("\n")
	}

	return routes.String()
}

func generateProxyHandlerFunction() string {
	return `// createProxyHandler creates a reverse proxy handler for the given backend URL and path
func createProxyHandler(backendURL, backendPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Simple proxy implementation - in production, use httputil.ReverseProxy
		// This is a placeholder for the actual proxy logic
		w.WriteHeader(http.StatusNotImplemented)
		fmt.Fprintf(w, "Proxy to %s%s not implemented yet", backendURL, backendPath)
	}
}`
}

func createBasicMainGo(config *types.BFFConfig) error {
	mainTemplate := `package main

import (
	"fmt"
	"log"
	"net/http"

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
{{range $serviceName, $service := .Services}}
	// {{$serviceName}} service routes
{{range $endpoint := $service.Endpoints}}
	r.{{chiMethod $endpoint.Method}}("{{$endpoint.ExposeAs}}", createProxyHandler("{{$service.BaseURL}}", "{{$endpoint.Path}}"))
{{end}}
{{end}}

	fmt.Println("🚀 BFF server starting on :{{.Settings.Port}}")
	log.Fatal(http.ListenAndServe(":{{.Settings.Port}}", r))
}

// createProxyHandler creates a reverse proxy handler for the given backend URL and path
func createProxyHandler(backendURL, backendPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Simple proxy implementation - in production, use httputil.ReverseProxy
		// This is a placeholder for the actual proxy logic
		w.WriteHeader(http.StatusNotImplemented)
		fmt.Fprintf(w, "Proxy to %s%s not implemented yet", backendURL, backendPath)
	}
}`

	tmpl, err := template.New("main").Funcs(template.FuncMap{
		"chiMethod": chiMethod,
	}).Parse(mainTemplate)
	if err != nil {
		return err
	}

	// Set default port if not specified
	if config.Settings.Port == 0 {
		config.Settings.Port = 8080
	}

	file, err := os.Create("main.go")
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, config)
}

func generateServerMain(config *types.BFFConfig) error {
	serverTemplate := `package main

import (
	"fmt"
	"log"
	"net/http"

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
{{range $serviceName, $service := .Services}}
	// {{$serviceName}} service routes
{{range $endpoint := $service.Endpoints}}
	r.{{chiMethod $endpoint.Method}}("{{$endpoint.ExposeAs}}", createProxyHandler("{{$service.BaseURL}}", "{{$endpoint.Path}}"))
{{end}}
{{end}}

	fmt.Println("🚀 BFF server starting on :{{.Settings.Port}}")
	log.Fatal(http.ListenAndServe(":{{.Settings.Port}}", r))
}

// createProxyHandler creates a reverse proxy handler for the given backend URL and path
func createProxyHandler(backendURL, backendPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Simple proxy implementation - in production, use httputil.ReverseProxy
		// This is a placeholder for the actual proxy logic
		w.WriteHeader(http.StatusNotImplemented)
		fmt.Fprintf(w, "Proxy to %s%s not implemented yet", backendURL, backendPath)
	}
}`

	tmpl, err := template.New("server").Funcs(template.FuncMap{
		"chiMethod": chiMethod,
	}).Parse(serverTemplate)
	if err != nil {
		return err
	}

	// Ensure cmd/server directory exists
	if err := os.MkdirAll("cmd/server", 0755); err != nil {
		return err
	}

	// Set default port if not specified
	if config.Settings.Port == 0 {
		config.Settings.Port = 8080
	}

	file, err := os.Create("cmd/server/main.go")
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, config)
}

// generateMainGoWithScaffolding generates main.go using regeneration-safe scaffolding
func generateMainGoWithScaffolding(config *types.BFFConfig, generator *scaffolding.Generator) error {
	// Generate proxy routes content
	proxyRoutes := generateProxyRoutesCode(config)
	
	// Use scaffolding to generate/update the file
	return generator.GenerateFile("main.go", proxyRoutes)
}

// generateServerMainWithScaffolding generates server main using regeneration-safe scaffolding
func generateServerMainWithScaffolding(config *types.BFFConfig, generator *scaffolding.Generator) error {
	// Generate server content
	serverContent := generateServerContent(config)
	
	// Use scaffolding to generate/update the file
	return generator.GenerateFile("cmd/server/main.go", serverContent)
}

// generateServerContent generates the server main content
func generateServerContent(config *types.BFFConfig) string {
	var content strings.Builder
	
	content.WriteString(`package main

import (
	"fmt"
	"log"
	"net/http"

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
`)

	// Add proxy routes
	for serviceName, service := range config.Services {
		content.WriteString(fmt.Sprintf("\t// %s service routes\n", serviceName))
		for _, endpoint := range service.Endpoints {
			method := chiMethod(endpoint.Method)
			content.WriteString(fmt.Sprintf("\tr.%s(\"%s\", createProxyHandler(\"%s\", \"%s\"))\n",
				method, endpoint.ExposeAs, service.BaseURL, endpoint.Path))
		}
		content.WriteString("\n")
	}

	content.WriteString(fmt.Sprintf(`
	fmt.Println("🚀 BFF server starting on :%d")
	log.Fatal(http.ListenAndServe(":%d", r))
}

// createProxyHandler creates a reverse proxy handler for the given backend URL and path
func createProxyHandler(backendURL, backendPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Simple proxy implementation - in production, use httputil.ReverseProxy
		// This is a placeholder for the actual proxy logic
		w.WriteHeader(http.StatusNotImplemented)
		fmt.Fprintf(w, "Proxy to %%s%%s not implemented yet", backendURL, backendPath)
	}
}`, config.Settings.Port, config.Settings.Port))

	return content.String()
}

// chiMethod converts HTTP method to Chi router method name
func chiMethod(method string) string {
	switch strings.ToUpper(method) {
	case "GET":
		return "Get"
	case "POST":
		return "Post"
	case "PUT":
		return "Put"
	case "DELETE":
		return "Delete"
	case "PATCH":
		return "Patch"
	case "HEAD":
		return "Head"
	case "OPTIONS":
		return "Options"
	default:
		return "Get" // Default to Get for unknown methods
	}
}

// generateNodeJS generates code for Node.js projects
func generateNodeJS() error {
	if checkMode {
		fmt.Println("🔍 Check mode: Analyzing what would be changed")
	} else if dryRun {
		fmt.Println("🔍 Dry run: Showing what would be changed")
	} else {
		fmt.Println("🔧 Generating Node.js routes from bffgen.config.json")
	}
	fmt.Println()

	// Check if config file exists
	if _, err := os.Stat("bffgen.config.json"); os.IsNotExist(err) {
		fmt.Println("❌ bffgen.config.json not found in current directory")
		fmt.Println("💡 Run 'bffgen init <project-name>' first or navigate to a BFF project directory")
		return fmt.Errorf("config file not found")
	}

	// Load bffgen.config.json
	configData, err := os.ReadFile("bffgen.config.json")
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Get backends
	backends, ok := config["backends"].([]interface{})
	if !ok || len(backends) == 0 {
		fmt.Println("⚠️  No backends configured in bffgen.config.json")
		fmt.Println("💡 Add backends using 'bffgen add-route' or 'bffgen add-template'")
		return nil
	}

	LogVerbose("Found %d backends to generate", len(backends))

	// Detect framework (Express or Fastify)
	framework := "express"
	if project, ok := config["project"].(map[string]interface{}); ok {
		if fw, ok := project["framework"].(string); ok {
			framework = fw
		}
	}

	fmt.Printf("📝 Generating routes for %s\n", framework)

	// Generate route files for each backend
	routesGenerated := 0
	for _, backend := range backends {
		backendMap, ok := backend.(map[string]interface{})
		if !ok {
			continue
		}

		serviceName, _ := backendMap["name"].(string)
		if serviceName == "" {
			continue
		}

		// Skip backends with no endpoints
		endpoints, ok := backendMap["endpoints"].([]interface{})
		if !ok || len(endpoints) == 0 {
			LogVerbose("Skipping %s (no endpoints defined)", serviceName)
			fmt.Printf("⏭️  Skipped service '%s' (no endpoints defined)\n", serviceName)
			continue
		}

		// Generate route file
		if err := generateNodeJSRouteFile(serviceName, backendMap, framework); err != nil {
			fmt.Printf("⚠️  Warning: Failed to generate routes for %s: %v\n", serviceName, err)
			continue
		}

		routesGenerated++
		fmt.Printf("✅ Generated routes for service: %s\n", serviceName)
	}

	if !checkMode && !dryRun {
		fmt.Println()
		fmt.Printf("✅ Code generation completed! Generated %d route files.\n", routesGenerated)
		fmt.Println("📁 Route files created in: src/routes/")
		fmt.Println()
		fmt.Println("🚀 Routes are automatically loaded by src/index.js")
		fmt.Println()
		fmt.Println("💡 Next steps:")
		fmt.Println("   1. Review generated route files in src/routes/")
		fmt.Println("   2. Start your backend services")
		fmt.Println("   3. Run: npm run dev")
	}

	LogVerbose("Code generation completed successfully")

	return nil
}

// generateNodeJSRouteFile generates a route file for a Node.js service
func generateNodeJSRouteFile(serviceName string, backend map[string]interface{}, framework string) error {
	// Create routes directory if it doesn't exist
	routesDir := filepath.Join("src", "routes")
	if err := os.MkdirAll(routesDir, 0755); err != nil {
		return err
	}

	baseURL, _ := backend["baseUrl"].(string)
	endpoints, ok := backend["endpoints"].([]interface{})
	if !ok {
		endpoints = []interface{}{}
	}

	var content strings.Builder

	if framework == "fastify" {
		// Generate Fastify route file
		content.WriteString(fmt.Sprintf(`/**
 * %s Service Routes (Fastify)
 * Auto-generated by bffgen
 */

async function %sRoutes(fastify, options) {
`, serviceName, serviceName))

		for _, ep := range endpoints {
			endpoint, ok := ep.(map[string]interface{})
			if !ok {
				continue
			}

			path, _ := endpoint["exposeAs"].(string)
			method, _ := endpoint["method"].(string)
			backendPath, _ := endpoint["path"].(string)
			requiresAuth, _ := endpoint["requiresAuth"].(bool)

			if path == "" || method == "" {
				continue
			}

			authMiddleware := ""
			if requiresAuth {
				authMiddleware = "\n    onRequest: [fastify.authenticate],"
			}

			content.WriteString(fmt.Sprintf(`
  fastify.%s('%s', {%s
    schema: {
      description: '%s %s',
      tags: ['%s']
    }
  }, async (request, reply) => {
    try {
      const baseURL = process.env.%s_URL || '%s';
      const response = await fetch(` + "`${baseURL}%s`" + `, {
        method: '%s',
        headers: {
          'Content-Type': 'application/json',
          ...(request.headers.authorization && { 'Authorization': request.headers.authorization })
        },
        ...(request.body && { body: JSON.stringify(request.body) })
      });
      
      if (!response.ok) {
        reply.status(response.status);
        return { error: '%s service error', message: ` + "`Backend returned status ${response.status}`" + ` };
      }
      
      const data = await response.json();
      return data;
    } catch (error) {
      fastify.log.error('%s service error:', error);
      reply.status(500);
      return { error: 'Internal Server Error', message: 'Failed to fetch from %s service' };
    }
  });
`, strings.ToLower(method), path, authMiddleware, method, path, serviceName,
				strings.ToUpper(serviceName), baseURL, backendPath, method,
				serviceName, serviceName, serviceName))
		}

		content.WriteString(fmt.Sprintf(`}

module.exports = %sRoutes;
`, serviceName))

	} else {
		// Generate Express route file
		content.WriteString(fmt.Sprintf(`/**
 * %s Service Routes (Express)
 * Auto-generated by bffgen
 */

const express = require('express');
const router = express.Router();
const { asyncHandler } = require('../middleware/errorHandler');
const { authenticate } = require('../middleware/auth');

`, serviceName))

		for _, ep := range endpoints {
			endpoint, ok := ep.(map[string]interface{})
			if !ok {
				continue
			}

			path, _ := endpoint["exposeAs"].(string)
			method, _ := endpoint["method"].(string)
			backendPath, _ := endpoint["path"].(string)
			requiresAuth, _ := endpoint["requiresAuth"].(bool)

			if path == "" || method == "" {
				continue
			}

			authMiddleware := ""
			if requiresAuth {
				authMiddleware = "authenticate, "
			}

			content.WriteString(fmt.Sprintf(`
router.%s('%s', %sasyncHandler(async (req, res) => {
  try {
    const baseURL = process.env.%s_URL || '%s';
    const response = await fetch(` + "`${baseURL}%s`" + `, {
      method: '%s',
      headers: {
        'Content-Type': 'application/json',
        ...(req.headers.authorization && { 'Authorization': req.headers.authorization })
      },
      ...(req.body && { body: JSON.stringify(req.body) })
    });
    
    if (!response.ok) {
      return res.status(response.status).json({
        error: '%s service error',
        message: ` + "`Backend returned status ${response.status}`" + `
      });
    }
    
    const data = await response.json();
    res.json(data);
  } catch (error) {
    console.error('%s service error:', error);
    res.status(500).json({
      error: 'Internal Server Error',
      message: 'Failed to fetch from %s service'
    });
  }
}));
`, strings.ToLower(method), path, authMiddleware,
				strings.ToUpper(serviceName), baseURL, backendPath, method,
				serviceName, serviceName, serviceName))
		}

		content.WriteString("\nmodule.exports = router;\n")
	}

	// Write route file
	filename := filepath.Join(routesDir, fmt.Sprintf("%s.js", serviceName))
	if err := os.WriteFile(filename, []byte(content.String()), 0644); err != nil {
		return err
	}

	return nil
}
