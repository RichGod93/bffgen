package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/RichGod93/bffgen/internal/utils"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose and fix common BFF project issues",
	Long: `Doctor command performs health checks on your BFF project and suggests fixes
for common issues like missing dependencies, configuration problems, and
outdated files.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runDoctor(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running doctor: %v\n", err)
			os.Exit(1)
		}
	},
}

func runDoctor() error {
	fmt.Println("🔍 Running BFF project health check...")
	fmt.Println()

	var issues []string
	var warnings []string
	var info []string

	// Check if we're in a BFF project directory
	if !isBFFProject() {
		fmt.Println("❌ Not in a BFF project directory")
		fmt.Println("💡 Run 'bffgen init <project-name>' to create a new project")
		return nil
	}

	// Detect project type
	projectType := getProjectType()
	if projectType == "nodejs" {
		fmt.Printf("📦 Detected Node.js project\n\n")
	} else if projectType == "go" {
		fmt.Printf("🔷 Detected Go project\n\n")
	}

	// Check configuration file
	if err := checkConfigFile(projectType); err != nil {
		issues = append(issues, fmt.Sprintf("Configuration file issue: %v", err))
	} else {
		fmt.Println("✅ Configuration file is valid")
	}

	// Check project-specific files
	if projectType == "go" {
		// Check Go module
		if err := checkGoModule(); err != nil {
			warnings = append(warnings, fmt.Sprintf("Go module issue: %v", err))
		} else {
			fmt.Println("✅ Go module is properly configured")
		}

		// Check Go dependencies
		if err := checkGoDependencies(); err != nil {
			warnings = append(warnings, fmt.Sprintf("Dependency issue: %v", err))
		} else {
			fmt.Println("✅ Dependencies are up to date")
		}

		// Check generated Go files
		if err := checkGoGeneratedFiles(); err != nil {
			warnings = append(warnings, fmt.Sprintf("Generated files issue: %v", err))
		} else {
			fmt.Println("✅ Generated files are present")
		}

		// Check Go environment
		if err := checkGoEnvironment(); err != nil {
			warnings = append(warnings, fmt.Sprintf("Environment issue: %v", err))
		} else {
			fmt.Println("✅ Environment is properly configured")
		}

		// Check for placeholder code
		if placeholderInfo := checkGoPlaceholderCode(); placeholderInfo != "" {
			info = append(info, placeholderInfo)
		}
	} else if projectType == "nodejs" {
		// Check Node.js dependencies
		if err := checkNodeDependencies(); err != nil {
			warnings = append(warnings, fmt.Sprintf("Dependency issue: %v", err))
		} else {
			fmt.Println("✅ Dependencies are properly configured")
		}

		// Check generated Node.js files
		if err := checkNodeGeneratedFiles(); err != nil {
			warnings = append(warnings, fmt.Sprintf("Generated files issue: %v", err))
		} else {
			fmt.Println("✅ Generated files are present")
		}

		// Check Node.js environment
		if err := checkNodeEnvironment(); err != nil {
			warnings = append(warnings, fmt.Sprintf("Environment issue: %v", err))
		} else {
			fmt.Println("✅ Environment is properly configured")
		}

		// Check for placeholder code in routes
		if placeholderInfo := checkNodePlaceholderCode(); placeholderInfo != "" {
			info = append(info, placeholderInfo)
		}
	}

	fmt.Println()

	// Report issues
	if len(issues) > 0 {
		fmt.Println("🚨 Issues found:")
		for _, issue := range issues {
			fmt.Printf("   - %s\n", issue)
		}
		fmt.Println()
	}

	// Report warnings
	if len(warnings) > 0 {
		fmt.Println("⚠️  Warnings:")
		for _, warning := range warnings {
			fmt.Printf("   - %s\n", warning)
		}
		fmt.Println()
	}

	// Report informational notes
	if len(info) > 0 {
		fmt.Println("ℹ️  Notes:")
		for _, note := range info {
			fmt.Printf("   - %s\n", note)
		}
		fmt.Println()
	}

	// Provide recommendations
	if len(issues) == 0 && len(warnings) == 0 {
		fmt.Println("🎉 Your BFF project is healthy!")
	} else {
		projType := getProjectType()
		fmt.Println("💡 Recommendations:")
		if len(issues) > 0 {
			fmt.Println("   - Fix the issues above before proceeding")
		}
		if len(warnings) > 0 {
			fmt.Println("   - Consider addressing the warnings for optimal performance")
		}
		fmt.Println("   - Run 'bffgen generate' to regenerate code if needed")

		if projType == "go" {
			fmt.Println("   - Run 'go mod tidy' to clean up dependencies")
		} else if projType == "nodejs" {
			fmt.Println("   - Run 'npm install' to install dependencies")
			fmt.Println("   - Copy .env.example to .env and configure variables")
		}
	}

	return nil
}

func isBFFProject() bool {
	// Check for either Go or Node.js config files
	_, goErr := os.Stat("bff.config.yaml")
	_, nodeErr := os.Stat("bffgen.config.json")
	return goErr == nil || nodeErr == nil
}

func getProjectType() string {
	// Check for Node.js project
	if _, err := os.Stat("bffgen.config.json"); err == nil {
		return "nodejs"
	}
	// Check for Go project
	if _, err := os.Stat("bff.config.yaml"); err == nil {
		return "go"
	}
	return "unknown"
}

func checkConfigFile(projectType string) error {
	if projectType == "nodejs" {
		return checkNodeConfigFile()
	}
	return checkGoConfigFile()
}

func checkGoConfigFile() error {
	config, err := utils.LoadConfig("bff.config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(config.Services) == 0 {
		return fmt.Errorf("no services configured")
	}

	// Check for common configuration issues
	for name, service := range config.Services {
		if service.BaseURL == "" {
			return fmt.Errorf("service %s has empty base URL", name)
		}

		if len(service.Endpoints) == 0 {
			return fmt.Errorf("service %s has no endpoints", name)
		}

		for _, endpoint := range service.Endpoints {
			if endpoint.Path == "" {
				return fmt.Errorf("service %s has endpoint with empty path", name)
			}
			if endpoint.ExposeAs == "" {
				return fmt.Errorf("service %s has endpoint with empty expose path", name)
			}
		}
	}

	return nil
}

func checkNodeConfigFile() error {
	// Check if bffgen.config.json exists
	if _, err := os.Stat("bffgen.config.json"); os.IsNotExist(err) {
		return fmt.Errorf("bffgen.config.json not found")
	}

	// Try to parse it as valid JSON
	configData, err := os.ReadFile("bffgen.config.json")
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Check for required fields
	if _, ok := config["project"]; !ok {
		return fmt.Errorf("missing 'project' section in config")
	}

	if _, ok := config["backends"]; !ok {
		return fmt.Errorf("missing 'backends' section in config")
	}

	return nil
}

func checkGoModule() error {
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		return fmt.Errorf("go.mod file not found")
	}

	if _, err := os.Stat("go.sum"); os.IsNotExist(err) {
		return fmt.Errorf("go.sum file not found (run 'go mod tidy')")
	}

	return nil
}

func checkGoDependencies() error {
	// Check if main.go exists and has required imports
	if _, err := os.Stat("main.go"); os.IsNotExist(err) {
		return fmt.Errorf("main.go not found (run 'bffgen generate')")
	}

	return nil
}

func checkNodeDependencies() error {
	// Check if package.json exists
	if _, err := os.Stat("package.json"); os.IsNotExist(err) {
		return fmt.Errorf("package.json not found")
	}

	// Check if node_modules exists
	if _, err := os.Stat("node_modules"); os.IsNotExist(err) {
		return fmt.Errorf("node_modules not found (run 'npm install')")
	}

	return nil
}

func checkGoPlaceholderCode() string {
	// Read main.go and check for placeholder proxy handlers
	content, err := os.ReadFile("main.go")
	if err != nil {
		return ""
	}

	contentStr := string(content)

	// Count placeholder proxy handlers
	placeholderCount := strings.Count(contentStr, "not implemented yet")
	if placeholderCount > 0 {
		return fmt.Sprintf("Route handlers contain %d placeholder(s). Implement proxy logic before production deployment.", placeholderCount)
	}

	return ""
}

func checkNodePlaceholderCode() string {
	// Check for TODO comments in generated files
	todoCount := 0

	// Check routes directory
	if files, err := os.ReadDir("src/routes"); err == nil {
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".js") {
				content, err := os.ReadFile("src/routes/" + file.Name())
				if err == nil {
					todoCount += strings.Count(string(content), "TODO:")
				}
			}
		}
	}

	if todoCount > 0 {
		return fmt.Sprintf("Found %d TODO comment(s) in route files. Review and implement before production.", todoCount)
	}

	return ""
}

func checkGoGeneratedFiles() error {
	// Check for required generated files
	requiredFiles := []string{
		"main.go",
		"cmd/server/main.go",
	}

	for _, file := range requiredFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("required file %s not found", file)
		}
	}

	return nil
}

func checkNodeGeneratedFiles() error {
	// Check for required Node.js files
	requiredFiles := []string{
		"src/index.js",
		"src/services/httpClient.js",
		"src/utils/logger.js",
	}

	var missing []string
	for _, file := range requiredFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			missing = append(missing, file)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing files: %s", strings.Join(missing, ", "))
	}

	return nil
}

func checkGoEnvironment() error {
	// Check for common Go environment variables
	envVars := []string{
		"ENCRYPTION_KEY",
		"JWT_SECRET",
		"REDIS_URL",
	}

	var missingVars []string
	for _, envVar := range envVars {
		if os.Getenv(envVar) == "" {
			missingVars = append(missingVars, envVar)
		}
	}

	if len(missingVars) > 0 {
		return fmt.Errorf("missing environment variables: %s", strings.Join(missingVars, ", "))
	}

	return nil
}

func checkNodeEnvironment() error {
	// Check for .env file or .env.example
	hasEnv := false
	if _, err := os.Stat(".env"); err == nil {
		hasEnv = true
	}
	if _, err := os.Stat(".env.example"); err == nil {
		hasEnv = true
	}

	if !hasEnv {
		return fmt.Errorf(".env or .env.example file not found")
	}

	// Check for common Node.js environment variables (optional check)
	recommendedVars := []string{"JWT_SECRET", "NODE_ENV"}
	var missing []string
	for _, envVar := range recommendedVars {
		if os.Getenv(envVar) == "" {
			missing = append(missing, envVar)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("recommended environment variables not set: %s (check .env.example)", strings.Join(missing, ", "))
	}

	return nil
}
