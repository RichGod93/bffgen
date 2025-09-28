package commands

import (
	"fmt"
	"os"
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
