package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/RichGod93/bffgen/internal/scaffolding"
	"github.com/RichGod93/bffgen/internal/types"
	"github.com/RichGod93/bffgen/internal/utils"
)

// BackendService represents a backend service configuration
type BackendService struct {
	Name      string
	BaseURL   string
	Port      int
	Path      string
	Endpoints []string
}

// ProjectOptions holds options for project initialization
type ProjectOptions struct {
	MiddlewareFlag   string
	ControllerType   string
	SkipTests        bool
	SkipDocs         bool
	LanguageExplicit bool // True if language was explicitly set via flag
}

// initializeProject initializes a new BFF project
func initializeProject(projectName string, langType scaffolding.LanguageType, framework string) (scaffolding.LanguageType, string, []BackendService, error) {
	return initializeProjectWithOptions(projectName, langType, framework, ProjectOptions{})
}

// initializeProjectWithOptions initializes a new BFF project with custom options
func initializeProjectWithOptions(projectName string, langType scaffolding.LanguageType, framework string, opts ProjectOptions) (scaffolding.LanguageType, string, []BackendService, error) {
	if err := os.MkdirAll(projectName, 0755); err != nil {
		return langType, framework, nil, fmt.Errorf("failed to create project directory: %w", err)
	}

	if err := createProjectDirectories(projectName, langType); err != nil {
		return langType, framework, nil, fmt.Errorf("failed to create directories: %w", err)
	}

	config, err := utils.LoadBFFGenConfig()
	if err != nil {
		fmt.Printf("⚠️  Warning: Could not load config: %v\n", err)
		config = types.GetDefaultConfig()
	}

	reader := bufio.NewReader(os.Stdin)
	prompter := utils.NewPromptConfig(reader, config.Defaults)

	// Only prompt for language if not explicitly set via flags
	if !opts.LanguageExplicit {
		var err error
		langType, framework, err = prompter.PromptLanguageSelection()
		if err != nil {
			return langType, framework, nil, fmt.Errorf("failed to select language: %w", err)
		}
	}

	corsOriginsList, err := prompter.PromptCORSSetting()
	if err != nil {
		return langType, framework, nil, fmt.Errorf("failed to configure CORS: %w", err)
	}

	backendArch, err := prompter.PromptBackendArchitecture()
	if err != nil {
		return langType, framework, nil, fmt.Errorf("failed to select backend architecture: %w", err)
	}

	backendServices, err := configureBackendServices(backendArch, reader)
	if err != nil {
		return langType, framework, nil, fmt.Errorf("failed to configure backend services: %w", err)
	}

	routeOption, err := prompter.PromptRouteConfiguration()
	if err != nil {
		return langType, framework, nil, fmt.Errorf("failed to configure routes: %w", err)
	}

	if routeOption == "2" {
		if err := copyTemplateFiles(projectName); err != nil {
			fmt.Printf("⚠️  Warning: Could not copy templates: %v\n", err)
		}
	}

	// Handle middleware selection (Node.js only)
	var selectedMiddleware []string
	if langType == scaffolding.LanguageNodeExpress || langType == scaffolding.LanguageNodeFastify {
		if opts.MiddlewareFlag != "" {
			// Use flag value
			selectedMiddleware = parseMiddlewareFlag(opts.MiddlewareFlag)
		} else {
			// Interactive prompt
			selectedMiddleware, err = promptMiddlewareSelection(reader)
			if err != nil {
				return langType, framework, nil, fmt.Errorf("failed to select middleware: %w", err)
			}
		}
	}

	corsConfig := generateCORSConfigWithLang(corsOriginsList, framework, langType)

	if langType == scaffolding.LanguageGo {
		if err := copyAuthPackage(projectName); err != nil {
			fmt.Printf("⚠️  Warning: Could not copy auth package: %v\n", err)
		}
	}

	if err := createDependencyFiles(projectName, langType, framework); err != nil {
		return langType, framework, nil, fmt.Errorf("failed to create dependency files: %w", err)
	}

	if err := createMainFileWithOptions(projectName, langType, framework, corsConfig, backendServices, opts); err != nil {
		return langType, framework, nil, fmt.Errorf("failed to create main file: %w", err)
	}

	// Generate additional middleware files for Node.js projects
	if (langType == scaffolding.LanguageNodeExpress || langType == scaffolding.LanguageNodeFastify) && len(selectedMiddleware) > 0 {
		if err := createAdditionalMiddleware(projectName, langType, framework, selectedMiddleware); err != nil {
			fmt.Printf("⚠️  Warning: Could not create additional middleware: %v\n", err)
		}
	}

	if err := createBFFConfig(projectName, backendServices); err != nil {
		return langType, framework, nil, fmt.Errorf("failed to create bff.config.yaml: %w", err)
	}

	if err := createReadme(projectName, langType); err != nil {
		return langType, framework, nil, fmt.Errorf("failed to create README.md: %w", err)
	}

	// Save controller type preference for generate command
	if langType == scaffolding.LanguageNodeExpress || langType == scaffolding.LanguageNodeFastify {
		saveControllerTypePreference(projectName, opts.ControllerType)
	}

	showRouteConfigInstructions(routeOption, projectName)
	updateAndSaveConfig(config, framework, corsOriginsList, routeOption, projectName)

	return langType, framework, backendServices, nil
}

// Helper functions
func copyTemplateFiles(projectName string) error {
	templateFiles := []string{"auth.yaml", "ecommerce.yaml", "content.yaml"}
	for _, templateFile := range templateFiles {
		srcPath := filepath.Join("internal", "templates", templateFile)
		dstPath := filepath.Join(projectName, "internal", "templates", templateFile)

		if _, err := os.Stat(srcPath); err == nil {
			if err := copyFile(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to copy %s: %w", templateFile, err)
			}
		}
	}
	return nil
}

func createBFFConfig(projectName string, backendServices []BackendService) error {
	configContent := generateEnhancedBFFConfig(backendServices, projectName)
	return os.WriteFile(filepath.Join(projectName, "bff.config.yaml"), []byte(configContent), 0644)
}

func createReadme(projectName string, langType scaffolding.LanguageType) error {
	var installCmd, runCmd string

	if langType == scaffolding.LanguageGo {
		installCmd = "go mod tidy"
		runCmd = "go run main.go"
	} else {
		installCmd = "npm install"
		runCmd = "npm start"
	}

	readmeContent := fmt.Sprintf(`# %s

A Backend-for-Frontend (BFF) service generated by bffgen.

## Getting Started

1. Install dependencies: %s
2. Configure your backend services in bff.config.yaml
3. Run the development server: %s

The server will start on http://localhost:8080
`, projectName, installCmd, runCmd)

	return os.WriteFile(filepath.Join(projectName, "README.md"), []byte(readmeContent), 0644)
}

func showRouteConfigInstructions(routeOption, projectName string) {
	fmt.Println()
	switch routeOption {
	case "1", "2", "3":
		fmt.Println("💡 To add routes later, run:")
		fmt.Printf("   cd %s\n", projectName)
		fmt.Println("   bffgen add-route")
	}
}

func updateAndSaveConfig(config *types.BFFGenConfig, framework string, corsOriginsList []string, routeOption, projectName string) {
	config.Defaults.Framework = framework
	config.Defaults.CORSOrigins = make([]string, len(corsOriginsList))
	for i, origin := range corsOriginsList {
		config.Defaults.CORSOrigins[i] = strings.TrimPrefix(strings.TrimPrefix(origin, "http://"), "https://")
	}
	config.Defaults.RouteOption = routeOption

	if err := utils.SaveBFFGenConfig(config); err != nil {
		fmt.Printf("⚠️  Warning: Could not save config: %v\n", err)
	}

	if err := utils.UpdateRecentProject(projectName); err != nil {
		fmt.Printf("⚠️  Warning: Could not update recent projects: %v\n", err)
	}
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	return err
}

func copyAuthPackage(projectName string) error {
	authDir := filepath.Join(projectName, "internal", "auth")
	if err := os.MkdirAll(authDir, 0755); err != nil {
		return fmt.Errorf("failed to create auth directory: %w", err)
	}

	authFiles := []string{
		"internal/auth/secure_auth.go",
		"internal/auth/secure_auth_test.go",
	}

	for _, srcFile := range authFiles {
		dstFile := filepath.Join(projectName, srcFile)

		if _, err := os.Stat(srcFile); os.IsNotExist(err) {
			continue
		}

		if err := copyFile(srcFile, dstFile); err != nil {
			return fmt.Errorf("failed to copy %s: %w", srcFile, err)
		}
	}

	return nil
}

func configureBackendServices(arch string, reader *bufio.Reader) ([]BackendService, error) {
	switch arch {
	case "1":
		return configureMicroservices(reader), nil
	case "2":
		return configureMonolithic(reader), nil
	case "3":
		return configureHybrid(reader), nil
	default:
		return configureMicroservices(reader), nil
	}
}

func configureMicroservices(reader *bufio.Reader) []BackendService {
	var services []BackendService

	fmt.Println("\n🔧 Configuring Microservices Backend")
	fmt.Println("Enter your backend services (press Enter with empty name to finish):")

	for {
		fmt.Printf("✔ Service name (e.g., 'users', 'products', 'orders'): ")
		serviceName, _ := reader.ReadString('\n')
		serviceName = strings.TrimSpace(serviceName)
		if serviceName == "" {
			break
		}

		fmt.Printf("✔ Base URL for %s (e.g., 'http://localhost:4000/api'): ", serviceName)
		baseURL, _ := reader.ReadString('\n')
		baseURL = strings.TrimSpace(baseURL)
		if baseURL == "" {
			baseURL = fmt.Sprintf("http://localhost:400%d/api", len(services)+1)
			fmt.Printf("   Using default: %s\n", baseURL)
		}

		port := extractPortFromURL(baseURL, 4000+len(services))

		service := BackendService{
			Name:      serviceName,
			BaseURL:   baseURL,
			Port:      port,
			Path:      "",
			Endpoints: getDefaultEndpoints(serviceName),
		}

		services = append(services, service)
		fmt.Printf("✅ Added %s service on %s\n", serviceName, baseURL)
	}

	return services
}

func configureMonolithic(reader *bufio.Reader) []BackendService {
	fmt.Println("\n🔧 Configuring Monolithic Backend")

	fmt.Printf("✔ Backend base URL (e.g., 'http://localhost:3000/api'): ")
	baseURL, _ := reader.ReadString('\n')
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = "http://localhost:3000/api"
		fmt.Printf("   Using default: %s\n", baseURL)
	}

	port := extractPortFromURL(baseURL, 3000)

	serviceNames := []string{"users", "products", "orders", "cart", "auth"}
	var services []BackendService

	for _, serviceName := range serviceNames {
		service := BackendService{
			Name:      serviceName,
			BaseURL:   baseURL,
			Port:      port,
			Path:      "",
			Endpoints: getDefaultEndpoints(serviceName),
		}
		services = append(services, service)
	}

	fmt.Printf("✅ Configured monolithic backend on %s\n", baseURL)
	return services
}

func configureHybrid(reader *bufio.Reader) []BackendService {
	var services []BackendService

	fmt.Println("\n🔧 Configuring Hybrid Backend")
	fmt.Println("Enter your backend services (press Enter with empty name to finish):")

	for {
		fmt.Printf("✔ Service name (e.g., 'users', 'products', 'orders'): ")
		serviceName, _ := reader.ReadString('\n')
		serviceName = strings.TrimSpace(serviceName)
		if serviceName == "" {
			break
		}

		fmt.Printf("✔ Base URL for %s (e.g., 'http://localhost:3000/api/users'): ", serviceName)
		baseURL, _ := reader.ReadString('\n')
		baseURL = strings.TrimSpace(baseURL)
		if baseURL == "" {
			baseURL = fmt.Sprintf("http://localhost:3000/api/%s", serviceName)
			fmt.Printf("   Using default: %s\n", baseURL)
		}

		port := extractPortFromURL(baseURL, 3000)

		path := ""
		if strings.Contains(baseURL, "/api/") {
			pathParts := strings.Split(baseURL, "/api/")
			if len(pathParts) > 1 {
				path = "/" + pathParts[1]
			}
		}

		service := BackendService{
			Name:      serviceName,
			BaseURL:   baseURL,
			Port:      port,
			Path:      path,
			Endpoints: getDefaultEndpoints(serviceName),
		}

		services = append(services, service)
		fmt.Printf("✅ Added %s service on %s\n", serviceName, baseURL)
	}

	return services
}

func extractPortFromURL(url string, defaultPort int) int {
	if strings.Contains(url, ":") {
		parts := strings.Split(url, ":")
		if len(parts) >= 3 {
			portStr := strings.Split(parts[2], "/")[0]
			if p, err := strconv.Atoi(portStr); err == nil {
				return p
			}
		}
	}
	return defaultPort
}

func getDefaultEndpoints(serviceName string) []string {
	switch serviceName {
	case "users":
		return []string{"GET /users", "GET /users/:id", "POST /users", "PUT /users/:id", "DELETE /users/:id"}
	case "products":
		return []string{"GET /products", "GET /products/:id", "POST /products", "PUT /products/:id", "DELETE /products/:id"}
	case "orders":
		return []string{"GET /orders", "GET /orders/:id", "POST /orders", "PUT /orders/:id"}
	case "cart":
		return []string{"GET /cart", "POST /cart/items", "DELETE /cart/items/:id", "POST /cart/checkout"}
	case "auth":
		return []string{"POST /auth/login", "POST /auth/register", "POST /auth/refresh", "POST /auth/logout"}
	default:
		return []string{"GET /" + serviceName, "GET /" + serviceName + "/:id", "POST /" + serviceName, "PUT /" + serviceName + "/:id"}
	}
}

func generateEnhancedBFFConfig(backendServices []BackendService, projectName string) string {
	var configContent strings.Builder

	configContent.WriteString("# BFF Configuration\n")
	configContent.WriteString("services:\n")

	for _, service := range backendServices {
		configContent.WriteString(fmt.Sprintf("  %s:\n", service.Name))
		configContent.WriteString(fmt.Sprintf("    baseUrl: %s\n", service.BaseURL))
		configContent.WriteString("    endpoints:\n")

		for _, endpoint := range service.Endpoints {
			parts := strings.Split(endpoint, " ")
			if len(parts) >= 2 {
				method := parts[0]
				path := parts[1]
				name := strings.ReplaceAll(strings.TrimPrefix(path, "/"), "/", "_")

				configContent.WriteString(fmt.Sprintf("      - name: %s\n", name))
				configContent.WriteString(fmt.Sprintf("        path: %s\n", path))
				configContent.WriteString(fmt.Sprintf("        method: %s\n", method))
				configContent.WriteString(fmt.Sprintf("        exposeAs: %s\n", path))
			}
		}
	}

	configContent.WriteString("\nsettings:\n")
	configContent.WriteString("  port: 8080\n")
	configContent.WriteString("  timeout: 30s\n")

	return configContent.String()
}

func showBackendConfigSummary(backendServices []BackendService) {
	fmt.Println("\n📋 Backend Configuration Summary:")

	portGroups := make(map[int][]BackendService)
	for _, service := range backendServices {
		portGroups[service.Port] = append(portGroups[service.Port], service)
	}

	if len(portGroups) == 1 {
		var services []BackendService
		for _, s := range portGroups {
			services = s
		}

		if len(services) > 3 {
			fmt.Println("   Architecture: Monolithic")
			fmt.Printf("   - Backend: %s\n", services[0].BaseURL)
		} else {
			fmt.Println("   Architecture: Hybrid")
		}
	} else {
		fmt.Println("   Architecture: Microservices")
		for _, service := range backendServices {
			fmt.Printf("   - %s: %s\n", service.Name, service.BaseURL)
		}
	}
}

func showSetupInstructions(backendServices []BackendService, projectName string) {
	fmt.Println("\n🔧 Setup Instructions:")
	fmt.Println("   1. Start your backend services")
	fmt.Println("   2. Run the BFF server:")
	fmt.Printf("      cd %s && bffgen dev\n", projectName)
	fmt.Println("   3. Test: curl http://localhost:8080/health")
}

// promptMiddlewareSelection prompts user to select additional middleware
func promptMiddlewareSelection(reader *bufio.Reader) ([]string, error) {
	fmt.Println("\n🔧 Which additional middleware would you like to include?")
	fmt.Println("   (Authentication and Error Handling are always included)")
	fmt.Println("\n   1) Request Validation")
	fmt.Println("   2) Request Logging")
	fmt.Println("   3) Request ID Tracking")
	fmt.Println("   4) All of the above")
	fmt.Println("   5) None (minimal setup)")
	fmt.Print("\n✔ Select option (1-5) [4]: ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		input = "4" // Default to all middleware
	}

	var selected []string
	switch input {
	case "1":
		selected = []string{"validation"}
	case "2":
		selected = []string{"logger"}
	case "3":
		selected = []string{"requestId"}
	case "4":
		selected = []string{"validation", "logger", "requestId"}
	case "5":
		selected = []string{}
	default:
		fmt.Println("⚠️  Invalid option, defaulting to all middleware")
		selected = []string{"validation", "logger", "requestId"}
	}

	if len(selected) > 0 {
		fmt.Printf("✅ Selected middleware: %s\n", strings.Join(selected, ", "))
	} else {
		fmt.Println("✅ Minimal middleware setup selected")
	}

	return selected, nil
}

// parseMiddlewareFlag parses the middleware flag value
func parseMiddlewareFlag(flag string) []string {
	if flag == "none" || flag == "" {
		return []string{}
	}

	if flag == "all" {
		return []string{"validation", "logger", "requestId"}
	}

	// Split by comma and trim spaces
	parts := strings.Split(flag, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// saveControllerTypePreference saves controller type preference to project config
func saveControllerTypePreference(projectName, controllerType string) {
	configPath := filepath.Join(projectName, ".bffgen-config")
	content := fmt.Sprintf("controller_type=%s\n", controllerType)
	os.WriteFile(configPath, []byte(content), 0644)
}
