// File: generate_go.go
// Purpose: Go-specific code generation for Chi, Echo, and Fiber frameworks
// Contains all logic for generating Go BFF projects

package commands

import (
	"fmt"
	"os"
	"strings"
	"text/template"

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
		LogVerboseCommand("Proxy routes already exist in main.go, skipping")
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
	if err := os.WriteFile("main.go", []byte(newContent), utils.ProjectFilePerm); err != nil {
		return fmt.Errorf("failed to write main.go: %w", err)
	}

	LogVerboseCommand("Added proxy routes to existing main.go")
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

func createBasicMainGo(config *types.BFFConfig) error {
	mainTemplate := `package main

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
	defer func() { _ = file.Close() }()

	return tmpl.Execute(file, config)
}

func generateServerMain(config *types.BFFConfig) error {
	serverTemplate := `package main

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

	tmpl, err := template.New("server").Funcs(template.FuncMap{
		"chiMethod": chiMethod,
	}).Parse(serverTemplate)
	if err != nil {
		return err
	}

	// Ensure cmd/server directory exists
	if err := os.MkdirAll("cmd/server", utils.ProjectDirPerm); err != nil {
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
	defer func() { _ = file.Close() }()

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

