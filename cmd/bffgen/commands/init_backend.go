package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/RichGod93/bffgen/internal/scaffolding"
	"github.com/RichGod93/bffgen/internal/types"
	"github.com/RichGod93/bffgen/internal/utils"
)

// ProjectOptions holds options for project initialization
type ProjectOptions struct {
	MiddlewareFlag   string
	ControllerType   string
	SkipTests        bool
	SkipDocs         bool
	LanguageExplicit bool // True if language was explicitly set via flag
	// Infrastructure options
	IncludeCI      bool
	IncludeDocker  bool
	IncludeHealth  bool
	IncludeCompose bool
	// Python-specific options
	PkgManager     string
	AsyncEndpoints bool
}

// initializeProject initializes a new BFF project
// initializeProjectWithOptions initializes a new BFF project with custom options
func initializeProjectWithOptions(projectName string, langType scaffolding.LanguageType, framework string, opts ProjectOptions) (scaffolding.LanguageType, string, []types.BackendService, error) {
	if err := os.MkdirAll(projectName, utils.ProjectDirPerm); err != nil {
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

	// Only copy Go template files for Go projects
	if routeOption == "2" && langType == scaffolding.LanguageGo {
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

	if err := createDependencyFilesWithOptions(projectName, langType, framework, opts); err != nil {
		return langType, framework, nil, fmt.Errorf("failed to create dependency files: %w", err)
	}

	if err := createMainFileWithOptions(projectName, langType, framework, corsConfig, backendServices, opts); err != nil {
		return langType, framework, nil, fmt.Errorf("failed to create main file: %w", err)
	}

	// Generate additional files based on language
	if langType == scaffolding.LanguagePythonFastAPI {
		// Python-specific files
		if err := createFastAPIConfig(projectName); err != nil {
			return langType, framework, nil, fmt.Errorf("failed to create config.py: %w", err)
		}
		if err := createFastAPIDependencies(projectName); err != nil {
			return langType, framework, nil, fmt.Errorf("failed to create dependencies.py: %w", err)
		}
		if err := createPythonEnvFile(projectName); err != nil {
			return langType, framework, nil, fmt.Errorf("failed to create .env: %w", err)
		}
		if err := createPythonGitignore(projectName); err != nil {
			return langType, framework, nil, fmt.Errorf("failed to create .gitignore: %w", err)
		}
		if err := createPythonLogger(projectName); err != nil {
			return langType, framework, nil, fmt.Errorf("failed to create logger: %w", err)
		}
		if err := createPythonCacheManager(projectName); err != nil {
			fmt.Printf("⚠️  Warning: Could not create cache manager: %v\n", err)
		}
		if err := createPythonCircuitBreaker(projectName); err != nil {
			fmt.Printf("⚠️  Warning: Could not create circuit breaker: %v\n", err)
		}
		if err := createPythonMiddleware(projectName); err != nil {
			return langType, framework, nil, fmt.Errorf("failed to create middleware: %w", err)
		}
		if err := createPythonTestFiles(projectName, opts); err != nil {
			fmt.Printf("⚠️  Warning: Could not create test files: %v\n", err)
		}
		if err := createPythonBFFGenConfig(projectName, opts); err != nil {
			return langType, framework, nil, fmt.Errorf("failed to create bffgen.config.py.json: %w", err)
		}
		if err := createPythonSetupScript(projectName, opts); err != nil {
			return langType, framework, nil, fmt.Errorf("failed to create setup.sh: %w", err)
		}
		if err := createPythonREADME(projectName, opts); err != nil {
			return langType, framework, nil, fmt.Errorf("failed to create README.md: %w", err)
		}
		// Skip createBFFConfig and createReadme for Python as we have Python-specific versions
	} else {
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
	}

	// Save controller type preference for generate command
	if langType == scaffolding.LanguageNodeExpress || langType == scaffolding.LanguageNodeFastify {
		saveControllerTypePreference(projectName, opts.ControllerType)
	}

	// Generate infrastructure files based on flags
	fmt.Println()
	if opts.IncludeCI {
		if err := generateCIWorkflow(projectName, langType, opts.IncludeDocker); err != nil {
			fmt.Printf("⚠️  Warning: Failed to generate CI workflow: %v\n", err)
		} else {
			fmt.Println("✅ Generated GitHub Actions CI/CD workflow")
		}
	}

	if opts.IncludeDocker {
		if err := generateDockerfile(projectName, langType, framework, 8080); err != nil {
			fmt.Printf("⚠️  Warning: Failed to generate Dockerfile: %v\n", err)
		} else {
			fmt.Println("✅ Generated production Dockerfile and .dockerignore")
		}
	}

	if opts.IncludeHealth {
		if err := generateHealthChecks(projectName, langType, framework, backendServices); err != nil {
			fmt.Printf("⚠️  Warning: Failed to generate health checks: %v\n", err)
		} else {
			fmt.Println("✅ Generated enhanced health check endpoints")
		}

		// Also generate graceful shutdown when health checks are included
		if err := generateGracefulShutdown(projectName, langType, framework); err != nil {
			fmt.Printf("⚠️  Warning: Failed to generate graceful shutdown: %v\n", err)
		} else {
			fmt.Println("✅ Generated graceful shutdown handler")
		}
	}

	if opts.IncludeCompose {
		if err := generateDockerCompose(projectName, langType, backendServices, 8080); err != nil {
			fmt.Printf("⚠️  Warning: Failed to generate docker-compose: %v\n", err)
		} else {
			fmt.Println("✅ Generated development docker-compose.yml")
		}
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

func createBFFConfig(projectName string, backendServices []types.BackendService) error {
	configContent := generateEnhancedBFFConfig(backendServices, projectName)
	return os.WriteFile(filepath.Join(projectName, "bff.config.yaml"), []byte(configContent), utils.ProjectFilePerm)
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

	return os.WriteFile(filepath.Join(projectName, "README.md"), []byte(readmeContent), utils.ProjectFilePerm)
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
	defer func() { _ = sourceFile.Close() }()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = destFile.Close() }()

	_, err = destFile.ReadFrom(sourceFile)
	return err
}

func copyAuthPackage(projectName string) error {
	authDir := filepath.Join(projectName, "internal", "auth")
	if err := os.MkdirAll(authDir, utils.ProjectDirPerm); err != nil {
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

func configureBackendServices(arch string, reader *bufio.Reader) ([]types.BackendService, error) {
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

func configureMicroservices(reader *bufio.Reader) []types.BackendService {
	var services []types.BackendService

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

		service := types.BackendService{
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

func configureMonolithic(reader *bufio.Reader) []types.BackendService {
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
	var services []types.BackendService

	for _, serviceName := range serviceNames {
		service := types.BackendService{
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

func configureHybrid(reader *bufio.Reader) []types.BackendService {
	var services []types.BackendService

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

		service := types.BackendService{
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

func generateEnhancedBFFConfig(backendServices []types.BackendService, projectName string) string {
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

func showBackendConfigSummary(backendServices []types.BackendService) {
	fmt.Println("\n📋 Backend Configuration Summary:")

	portGroups := make(map[int][]types.BackendService)
	for _, service := range backendServices {
		portGroups[service.Port] = append(portGroups[service.Port], service)
	}

	if len(portGroups) == 1 {
		var services []types.BackendService
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

func showSetupInstructions(backendServices []types.BackendService, projectName string) {
	fmt.Println("\n🔧 Setup Instructions:")
	fmt.Println("   1. Start your backend services")
	fmt.Println("   2. Run the BFF server:")
	fmt.Printf("      cd %s && bffgen dev\n", projectName)
	fmt.Println("   3. Test: curl http://localhost:8080/health")
}

// showPostInitGuidance shows personalized post-initialization guidance
func showPostInitGuidance(projectName, runtime, framework string, backendServices []types.BackendService) {
	// Check required tools
	required, optional := utils.CheckRequiredTools(runtime)

	missingRequired := []utils.ToolInfo{}
	for _, tool := range required {
		if !tool.Installed {
			missingRequired = append(missingRequired, tool)
		}
	}

	// Show tool status
	if len(missingRequired) > 0 {
		fmt.Println("⚠️  Missing Required Tools:")
		for _, tool := range missingRequired {
			fmt.Printf("   ❌ %s: Not installed\n", tool.Name)
		}
		fmt.Println()
		fmt.Println("📖 Installation Instructions:")
		for _, tool := range missingRequired {
			fmt.Println(utils.GetToolInstallInstructions(tool.Name))
			fmt.Println()
		}
		fmt.Println("💡 After installing tools, navigate to the project:")
		fmt.Printf("   cd %s\n", projectName)
		fmt.Println()
		return
	}

	fmt.Println("✅ All required tools are installed")
	fmt.Println()

	// Show optional tools status
	if len(optional) > 0 {
		installedOptional := []string{}
		missingOptional := []string{}

		for _, tool := range optional {
			if tool.Installed {
				installedOptional = append(installedOptional, tool.Name)
			} else {
				missingOptional = append(missingOptional, tool.Name)
			}
		}

		if len(installedOptional) > 0 {
			fmt.Printf("ℹ️  Optional tools available: %s\n", strings.Join(installedOptional, ", "))
		}
		if len(missingOptional) > 0 {
			fmt.Printf("ℹ️  Optional tools not installed: %s\n", strings.Join(missingOptional, ", "))
		}
		fmt.Println()
	}

	// Personalized next steps based on runtime
	fmt.Println("📋 Next Steps:")
	fmt.Printf("   1. Navigate to project: [bold]cd %s[reset]\n", projectName)

	if strings.HasPrefix(runtime, "nodejs") {
		fmt.Println("   2. Install dependencies: npm install")

		// Check if user wants to run npm install now
		fmt.Print("\n💡 Run 'npm install' now? (Y/n): ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response == "" || response == "y" || response == "yes" {
			fmt.Println()
			fmt.Println("📦 Running npm install...")

			// Run npm install
			cmd := exec.Command("npm", "install")
			cmd.Dir = projectName
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				fmt.Printf("⚠️  npm install failed: %v\n", err)
				fmt.Printf("💡 You can run it manually: cd %s && npm install\n", projectName)
			} else {
				fmt.Println("✅ Dependencies installed successfully")
			}
			fmt.Println()
		}

		fmt.Println("   3. Configure environment: cp .env.example .env")
		fmt.Println("   4. Edit .env with your backend URLs")
		fmt.Println("   5. Start development: npm run dev")
		fmt.Println("   6. View API docs: http://localhost:8080/api-docs")

	} else {
		// Go project
		fmt.Println("   2. Download dependencies: go mod download")
		fmt.Println("   3. Set environment variables:")
		fmt.Println("      export JWT_SECRET=your-secret-key")
		fmt.Println("      export ENCRYPTION_KEY=your-encryption-key")
		fmt.Println("   4. Start development: go run main.go")
		fmt.Println("   5. Or use: bffgen dev")
	}

	fmt.Println()
	fmt.Println("🎯 Quick Commands:")
	fmt.Printf("   Add routes:        cd %s && bffgen add-route\n", projectName)
	fmt.Printf("   Use template:      cd %s && bffgen add-template auth\n", projectName)
	fmt.Printf("   Generate code:     cd %s && bffgen generate\n", projectName)
	fmt.Printf("   Validate config:   cd %s && bffgen config validate\n", projectName)
	fmt.Printf("   Check health:      cd %s && bffgen doctor\n", projectName)

	fmt.Println()
	fmt.Println("📚 Documentation:")
	fmt.Println("   - README.md in your project")
	fmt.Println("   - GitHub: https://github.com/RichGod93/bffgen")
	if strings.HasPrefix(runtime, "nodejs") {
		fmt.Println("   - Node.js Guide: docs/NODEJS_AGGREGATION.md")
	}

	fmt.Println()
	fmt.Println("🔍 Troubleshooting:")
	fmt.Println("   - If dependencies fail: delete node_modules and re-run npm install")
	fmt.Println("   - If ports conflict: set PORT environment variable")
	fmt.Println("   - For issues: https://github.com/RichGod93/bffgen/issues")
	fmt.Println()
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
	_ = os.WriteFile(configPath, []byte(content), utils.ProjectFilePerm)
}
