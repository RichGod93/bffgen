package commands

import (
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
		issues = append(issues, "Not in a BFF project directory (bff.config.yaml not found)")
		fmt.Println("❌ Not in a BFF project directory")
		fmt.Println("💡 Run 'bffgen init <project-name>' to create a new project")
		return nil
	}

	// Check configuration file
	if err := checkConfigFile(); err != nil {
		issues = append(issues, fmt.Sprintf("Configuration file issue: %v", err))
	} else {
		fmt.Println("✅ Configuration file is valid")
	}

	// Check Go module
	if err := checkGoModule(); err != nil {
		warnings = append(warnings, fmt.Sprintf("Go module issue: %v", err))
	} else {
		fmt.Println("✅ Go module is properly configured")
	}

	// Check dependencies
	if err := checkDependencies(); err != nil {
		warnings = append(warnings, fmt.Sprintf("Dependency issue: %v", err))
	} else {
		fmt.Println("✅ Dependencies are up to date")
	}

	// Check generated files
	if err := checkGeneratedFiles(); err != nil {
		warnings = append(warnings, fmt.Sprintf("Generated files issue: %v", err))
	} else {
		fmt.Println("✅ Generated files are present")
	}

	// Check environment setup
	if err := checkEnvironment(); err != nil {
		warnings = append(warnings, fmt.Sprintf("Environment issue: %v", err))
	} else {
		fmt.Println("✅ Environment is properly configured")
	}

	// Check for placeholder code (informational only)
	if placeholderInfo := checkPlaceholderCode(); placeholderInfo != "" {
		info = append(info, placeholderInfo)
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
		fmt.Println("💡 Recommendations:")
		if len(issues) > 0 {
			fmt.Println("   - Fix the issues above before proceeding")
		}
		if len(warnings) > 0 {
			fmt.Println("   - Consider addressing the warnings for optimal performance")
		}
		fmt.Println("   - Run 'bffgen generate' to regenerate code if needed")
		fmt.Println("   - Run 'go mod tidy' to clean up dependencies")
	}

	return nil
}

func isBFFProject() bool {
	_, err := os.Stat("bff.config.yaml")
	return err == nil
}

func checkConfigFile() error {
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

func checkGoModule() error {
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		return fmt.Errorf("go.mod file not found")
	}

	if _, err := os.Stat("go.sum"); os.IsNotExist(err) {
		return fmt.Errorf("go.sum file not found (run 'go mod tidy')")
	}

	return nil
}

func checkDependencies() error {
	// Check if main.go exists and has required imports
	if _, err := os.Stat("main.go"); os.IsNotExist(err) {
		return fmt.Errorf("main.go not found (run 'bffgen generate')")
	}

	return nil
}

func checkPlaceholderCode() string {
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

func checkGeneratedFiles() error {
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

func checkEnvironment() error {
	// Check for common environment variables
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
