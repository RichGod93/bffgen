package commands

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/RichGod93/bffgen/internal/types"
	"github.com/RichGod93/bffgen/internal/utils"
	"github.com/spf13/cobra"
)

var addRouteCmd = &cobra.Command{
	Use:   "add-route",
	Short: "Interactively add a backend endpoint to your BFF",
	Long:  `Interactively add a backend endpoint to your BFF configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := addRoute(); err != nil {
			fmt.Fprintf(os.Stderr, "Error adding route: %v\n", err)
			os.Exit(1)
		}
	},
}

func addRoute() error {
	// Detect project type
	projectType := detectProjectType()

	if projectType == "unknown" {
		fmt.Println("❌ No BFF project found in current directory")
		fmt.Println("💡 Run 'bffgen init <project-name>' first or navigate to a BFF project directory")
		return fmt.Errorf("no project configuration found")
	}

	fmt.Println("🔧 Adding a new route to your BFF")
	fmt.Println()

	// Handle based on project type
	if projectType == "nodejs" {
		return addRouteNodeJS()
	}

	// Default: Go project
	return addRouteGo()
}

// addRouteGo handles adding routes for Go projects
func addRouteGo() error {
	// Check if config file exists
	if _, err := os.Stat("bff.config.yaml"); os.IsNotExist(err) {
		fmt.Println("❌ bff.config.yaml not found in current directory")
		fmt.Println("💡 Run 'bffgen init <project-name>' first or navigate to a BFF project directory")
		return fmt.Errorf("config file not found")
	}

	// Load existing config
	config, err := utils.LoadConfig("bff.config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize services map if nil
	if config.Services == nil {
		config.Services = make(map[string]types.Service)
	}

	reader := bufio.NewReader(os.Stdin)

	// Get service name
	fmt.Print("✔ Service name: ")
	serviceName, _ := reader.ReadString('\n')
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	// Check if service exists, if not ask for base URL
	var baseURL string
	if service, exists := config.Services[serviceName]; exists {
		baseURL = service.BaseURL
		fmt.Printf("✔ Using existing service '%s' with base URL: %s\n", serviceName, baseURL)
	} else {
		fmt.Print("✔ Base URL: ")
		baseURL, _ = reader.ReadString('\n')
		baseURL = strings.TrimSpace(baseURL)
		if baseURL == "" {
			return fmt.Errorf("base URL cannot be empty")
		}
		// Validate URL format
		if err := validateURL(baseURL); err != nil {
			return fmt.Errorf("invalid base URL: %w", err)
		}
	}

	// Get endpoint details
	fmt.Print("✔ Endpoint name: ")
	endpointName, _ := reader.ReadString('\n')
	endpointName = strings.TrimSpace(endpointName)
	if endpointName == "" {
		return fmt.Errorf("endpoint name cannot be empty")
	}

	fmt.Print("✔ Path: ")
	path, _ := reader.ReadString('\n')
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	// Validate path format
	if err := validatePath(path); err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	fmt.Print("✔ Method (GET/POST/PUT/DELETE/PATCH): ")
	method, _ := reader.ReadString('\n')
	method = strings.TrimSpace(strings.ToUpper(method))
	if method == "" {
		method = "GET"
	}

	fmt.Print("✔ Expose as: ")
	exposeAs, _ := reader.ReadString('\n')
	exposeAs = strings.TrimSpace(exposeAs)
	if exposeAs == "" {
		exposeAs = path
	}

	// Create new endpoint
	newEndpoint := types.Endpoint{
		Name:     endpointName,
		Path:     path,
		Method:   method,
		ExposeAs: exposeAs,
	}

	// Check for duplicate endpoints
	if service, exists := config.Services[serviceName]; exists {
		for _, ep := range service.Endpoints {
			if ep.Method == method && ep.ExposeAs == exposeAs {
				return fmt.Errorf("duplicate endpoint: %s %s already exists in service '%s'", method, exposeAs, serviceName)
			}
		}
		service.Endpoints = append(service.Endpoints, newEndpoint)
		config.Services[serviceName] = service
	} else {
		config.Services[serviceName] = types.Service{
			BaseURL:   baseURL,
			Endpoints: []types.Endpoint{newEndpoint},
		}
	}

	// Save updated config
	if err := utils.SaveConfig(config, "bff.config.yaml"); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Printf("✅ Added route: %s %s → %s%s\n", method, exposeAs, baseURL, path)
	fmt.Println("💡 Run 'bffgen generate' to update your Go code")

	return nil
}

// addRouteNodeJS handles adding routes for Node.js projects
func addRouteNodeJS() error {
	// Check if config file exists
	if _, err := os.Stat("bffgen.config.json"); os.IsNotExist(err) {
		fmt.Println("❌ bffgen.config.json not found in current directory")
		fmt.Println("💡 Run 'bffgen init <project-name>' first or navigate to a BFF project directory")
		return fmt.Errorf("config file not found")
	}

	// Load existing bffgen.config.json
	configData, err := os.ReadFile("bffgen.config.json")
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Get or create backends array
	backends, ok := config["backends"].([]interface{})
	if !ok {
		backends = []interface{}{}
	}

	reader := bufio.NewReader(os.Stdin)

	// Get service name
	fmt.Print("✔ Service name: ")
	serviceName, _ := reader.ReadString('\n')
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	// Find existing backend or create new
	var targetBackend map[string]interface{}
	var baseURL string
	backendExists := false

	for _, b := range backends {
		backend, ok := b.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := backend["name"].(string)
		if name == serviceName {
			targetBackend = backend
			baseURL, _ = backend["baseUrl"].(string)
			backendExists = true
			fmt.Printf("✔ Using existing service '%s' with base URL: %s\n", serviceName, baseURL)
			break
		}
	}

	if !backendExists {
		fmt.Print("✔ Base URL: ")
		baseURL, _ = reader.ReadString('\n')
		baseURL = strings.TrimSpace(baseURL)
		if baseURL == "" {
			return fmt.Errorf("base URL cannot be empty")
		}
		// Validate URL format
		if err := validateURL(baseURL); err != nil {
			return fmt.Errorf("invalid base URL: %w", err)
		}

		targetBackend = map[string]interface{}{
			"name":    serviceName,
			"baseUrl": baseURL,
			"timeout": 30000,
			"retries": 3,
			"healthCheck": map[string]interface{}{
				"enabled":  true,
				"path":     "/health",
				"interval": 60000,
			},
			"endpoints": []interface{}{},
		}
	}

	// Get endpoint details
	fmt.Print("✔ Endpoint name: ")
	endpointName, _ := reader.ReadString('\n')
	endpointName = strings.TrimSpace(endpointName)
	if endpointName == "" {
		return fmt.Errorf("endpoint name cannot be empty")
	}

	fmt.Print("✔ Path: ")
	path, _ := reader.ReadString('\n')
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	// Validate path format
	if err := validatePath(path); err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	fmt.Print("✔ Method (GET/POST/PUT/DELETE/PATCH): ")
	method, _ := reader.ReadString('\n')
	method = strings.TrimSpace(strings.ToUpper(method))
	if method == "" {
		method = "GET"
	}

	fmt.Print("✔ Expose as: ")
	exposeAs, _ := reader.ReadString('\n')
	exposeAs = strings.TrimSpace(exposeAs)
	if exposeAs == "" {
		exposeAs = path
	}

	fmt.Print("✔ Requires authentication? (y/N): ")
	authInput, _ := reader.ReadString('\n')
	requiresAuth := strings.TrimSpace(strings.ToLower(authInput)) == "y" || strings.TrimSpace(strings.ToLower(authInput)) == "yes"

	// Create new endpoint
	newEndpoint := map[string]interface{}{
		"name":         endpointName,
		"path":         path,
		"method":       method,
		"exposeAs":     exposeAs,
		"requiresAuth": requiresAuth,
	}

	// Get existing endpoints
	endpoints, ok := targetBackend["endpoints"].([]interface{})
	if !ok {
		endpoints = []interface{}{}
	}

	// Check for duplicate endpoints
	for _, ep := range endpoints {
		endpoint, ok := ep.(map[string]interface{})
		if !ok {
			continue
		}
		epMethod, _ := endpoint["method"].(string)
		epExposeAs, _ := endpoint["exposeAs"].(string)
		if epMethod == method && epExposeAs == exposeAs {
			return fmt.Errorf("duplicate endpoint: %s %s already exists in service '%s'", method, exposeAs, serviceName)
		}
	}

	// Add new endpoint
	endpoints = append(endpoints, newEndpoint)
	targetBackend["endpoints"] = endpoints

	// Add backend to backends array if new
	if !backendExists {
		backends = append(backends, targetBackend)
		config["backends"] = backends
	}

	// Save updated config as JSON with proper indentation
	updatedData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile("bffgen.config.json", updatedData, 0644); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Printf("✅ Added route: %s %s → %s%s\n", method, exposeAs, baseURL, path)
	fmt.Println("💡 Run 'bffgen generate' to update your route files")

	return nil
}

// validateURL validates that a URL has proper format and scheme
func validateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check for scheme
	if parsedURL.Scheme == "" {
		return fmt.Errorf("URL must include scheme (http:// or https://)")
	}

	// Check for valid scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https, got: %s", parsedURL.Scheme)
	}

	// Check for host
	if parsedURL.Host == "" {
		return fmt.Errorf("URL must include host")
	}

	return nil
}

// validatePath validates that a path has proper format
func validatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Path should start with /
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path must start with /")
	}

	// Path should not contain spaces
	if strings.Contains(path, " ") {
		return fmt.Errorf("path cannot contain spaces")
	}

	// Path should only contain valid characters
	validPathRegex := regexp.MustCompile(`^/[a-zA-Z0-9/_\-:.{}]*$`)
	if !validPathRegex.MatchString(path) {
		return fmt.Errorf("path contains invalid characters (allowed: a-z, A-Z, 0-9, /, _, -, :, ., {, })")
	}

	return nil
}
