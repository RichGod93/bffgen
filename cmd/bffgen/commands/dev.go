package commands

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/RichGod93/bffgen/internal/types"
	"github.com/RichGod93/bffgen/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/spf13/cobra"
)

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Run local BFF server with proxying",
	Long:  `Run a local BFF server with proxying to defined backend services.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runDevServer(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running dev server: %v\n", err)
			os.Exit(1)
		}
	},
}

func runDevServer() error {
	// Check if we're in a BFF project directory
	configPath := "bff.config.yaml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Println("❌ bff.config.yaml not found in current directory")
		fmt.Println("💡 Run 'bffgen init <project-name>' first or navigate to a BFF project directory")
		return fmt.Errorf("config file not found")
	}

	// Load configuration
	config, err := utils.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Set default port if not specified
	port := config.Settings.Port
	if port == 0 {
		port = 8080
	}

	// Create router
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

	// Setup proxy routes
	setupProxyRoutes(r, config)

	fmt.Printf("🚀 BFF server starting on :%d\n", port)
	fmt.Println("📋 Aggregated routes:")

	// Print configured routes
	for _, service := range config.Services {
		for _, endpoint := range service.Endpoints {
			fmt.Printf("   %s  %s  → %s%s\n",
				endpoint.Method,
				endpoint.ExposeAs,
				service.BaseURL,
				endpoint.Path)
		}
	}

	fmt.Printf("\n🌐 Server running at http://localhost:%d\n", port)
	fmt.Println("💡 Health check: http://localhost:8080/health")

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
	return nil
}

func setupProxyRoutes(r *chi.Mux, config *types.BFFConfig) {
	for serviceName, service := range config.Services {
		baseURL, err := url.Parse(service.BaseURL)
		if err != nil {
			fmt.Printf("⚠️  Invalid base URL for service %s: %s\n", serviceName, service.BaseURL)
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
				fmt.Printf("⚠️  Unsupported method %s for endpoint %s\n", endpoint.Method, endpoint.Name)
			}
		}
	}
}
