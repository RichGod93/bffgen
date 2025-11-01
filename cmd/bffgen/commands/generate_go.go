// File: generate_go.go
// Purpose: Go-specific code generation for Chi, Echo, and Fiber frameworks
// Contains all logic for generating Go BFF projects

package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/RichGod93/bffgen/internal/scaffolding"
	"github.com/RichGod93/bffgen/internal/types"
	"github.com/RichGod93/bffgen/internal/utils"
)

// generateGo generates code for Go projects
func generateGo() error {
	LogVerboseCommand("Starting code generation from bff.config.yaml")

	// Load generation state
	state, err := utils.LoadState()
	if err != nil {
		return fmt.Errorf("failed to load generation state: %w", err)
	}

	// Create generator with regeneration-safe capabilities
	generator := scaffolding.NewGenerator()
	generator.SetCheckMode(checkMode)
	generator.SetDryRun(dryRun)
	generator.SetVerbose(verbose)
	generator.SetBackupDir(utils.GetBackupDir())

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
		LogWarning("No services configured in bff.config.yaml")
		fmt.Println("💡 Add services using 'bffgen add-route' or 'bffgen add-template'")
		return nil
	}

	LogVerboseCommand("Found %d services to generate", len(config.Services))

	// Track routes and check for duplicates
	newRoutes := 0
	skippedRoutes := 0
	for serviceName, service := range config.Services {
		for _, endpoint := range service.Endpoints {
			routeKey := fmt.Sprintf("%s:%s:%s", serviceName, endpoint.Method, endpoint.ExposeAs)

			// Check if route already generated (unless force mode)
			if !forceMode && state.IsRouteGenerated(serviceName, endpoint.Method, endpoint.ExposeAs) {
				LogVerboseCommand("Skipping already generated route: %s", routeKey)
				skippedRoutes++
				continue
			}

			// Track the route
			state.TrackRoute(serviceName, endpoint.Method, endpoint.Path, endpoint.ExposeAs)
			newRoutes++
		}
	}

	if !forceMode && skippedRoutes > 0 {
		fmt.Printf("ℹ️  Skipped %d already generated routes (use --force to regenerate)\n", skippedRoutes)
	}

	// Generate main.go with routes using regeneration-safe scaffolding
	if err := generateMainGoWithScaffolding(config, generator); err != nil {
		return fmt.Errorf("failed to generate main.go: %w", err)
	}
	state.TrackGeneratedFile("main.go", "", true)

	// Generate server entry point using regeneration-safe scaffolding
	if err := generateServerMainWithScaffolding(config, generator); err != nil {
		return fmt.Errorf("failed to generate server main: %w", err)
	}
	state.TrackGeneratedFile("cmd/server/main.go", "", true)

	// Save generation state
	if !checkMode && !dryRun {
		state.ProjectType = "go"
		if err := utils.SaveState(state); err != nil {
			fmt.Printf("⚠️  Warning: Failed to save generation state: %v\n", err)
		}
	}

	if !checkMode && !dryRun {
		fmt.Println("✅ Code generation completed!")
		fmt.Printf("📁 Updated files:\n")
		fmt.Println("   - main.go (with proxy routes)")
		fmt.Println("   - cmd/server/main.go (server entry point)")
		if newRoutes > 0 {
			fmt.Printf("   - Added %d new routes\n", newRoutes)
		}
		fmt.Println()
		fmt.Println("🚀 Run 'go run main.go' to start your BFF server")
		fmt.Println()
		fmt.Println("📮 Generate Postman collection: bffgen postman")
		fmt.Println("   This creates a ready-to-import collection for testing your BFF endpoints")
	}

	LogVerboseCommand("Code generation completed successfully")

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
}`
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
		// Parse the backend URL
		target, err := url.Parse(backendURL)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid backend URL: %%v", err), http.StatusInternalServerError)
			return
		}

		// Create reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(target)
		
		// Configure proxy behavior
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Proxy error: %%v", err)
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
		log.Printf("Proxying %%s %%s -> %%s%%s", r.Method, originalPath, backendURL, backendPath)

		// Serve the proxy request
		proxy.ServeHTTP(w, r)
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
