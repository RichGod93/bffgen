package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/RichGod93/bffgen/internal/scaffolding"
	"github.com/RichGod93/bffgen/internal/testgen"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test generation and management",
	Long:  `Generate comprehensive test suites for your BFF project`,
}

var testGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate test suites",
	Long: `Generate integration tests, E2E tests, and contract tests for your BFF.

Examples:
  bffgen test generate                    # Generate all test types
  bffgen test generate --type=integration # Only integration tests
  bffgen test generate --type=e2e         # Only E2E tests
  bffgen test generate --type=contracts   # Only contract tests`,
	RunE: runTestGenerate,
}

var (
	testType    string
	testVerbose bool
)

func init() {
	testGenerateCmd.Flags().StringVar(&testType, "type", "all", "Test type to generate (integration, e2e, contracts, all)")
	testGenerateCmd.Flags().BoolVarP(&testVerbose, "verbose", "v", false, "Verbose output")

	testCmd.AddCommand(testGenerateCmd)
}

func runTestGenerate(cmd *cobra.Command, args []string) error {
	// Detect project type
	projectDir := "."
	language, framework, err := detectTestProjectType(projectDir)
	if err != nil {
		return fmt.Errorf("failed to detect project type: %w", err)
	}

	fmt.Println("🧪 BFFGen Test Generator")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📁 Project: %s (%s)\n", framework, language)
	fmt.Printf("🎯 Generating: %s tests\n\n", testType)

	var testsGenerated int

	// Generate integration tests
	if testType == "all" || testType == "integration" {
		if err := generateIntegrationTests(projectDir, language, framework); err != nil {
			return fmt.Errorf("failed to generate integration tests: %w", err)
		}
		testsGenerated++
		fmt.Println("✅ Integration tests generated")
	}

	// Generate E2E tests
	if testType == "all" || testType == "e2e" {
		if err := generateE2ETests(projectDir, language, framework); err != nil {
			fmt.Printf("⚠️  E2E test generation: %v\n", err)
		} else {
			testsGenerated++
			fmt.Println("✅ E2E tests generated")
		}
	}

	// Generate contract tests
	if testType == "all" || testType == "contracts" {
		if err := generateContractTests(projectDir, language, framework); err != nil {
			fmt.Printf("⚠️  Contract test generation: %v\n", err)
		} else {
			testsGenerated++
			fmt.Println("✅ Contract tests generated")
		}
	}

	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("✨ Generated %d test suite(s)\n\n", testsGenerated)

	// Show next steps based on language
	showTestInstructions(language)

	return nil
}

func generateIntegrationTests(projectDir string, language scaffolding.LanguageType, framework string) error {
	// TODO: Parse existing routes from project
	sampleRoutes := []testgen.RouteConfig{
		{Path: "/api/users", Method: "GET", RequiresAuth: true},
		{Path: "/api/users/:id", Method: "GET", RequiresAuth: true},
		{Path: "/api/products", Method: "GET", RequiresAuth: false},
	}

	config := testgen.TestConfig{
		ProjectName: filepath.Base(projectDir),
		Language:    language,
		Framework:   framework,
		Routes:      sampleRoutes,
		OutputDir:   projectDir,
	}

	generator := testgen.NewGenerator(config)
	return generator.GenerateIntegrationTests()
}

func generateE2ETests(projectDir string, language scaffolding.LanguageType, framework string) error {
	// TODO: Implement E2E test generation
	return fmt.Errorf("not yet implemented")
}

func generateContractTests(projectDir string, language scaffolding.LanguageType, framework string) error {
	// TODO: Implement contract test generation
	return fmt.Errorf("not yet implemented")
}

func detectTestProjectType(projectDir string) (scaffolding.LanguageType, string, error) {
	// Check for package.json (Node.js)
	if fileExists(filepath.Join(projectDir, "package.json")) {
		return scaffolding.LanguageNodeExpress, "express", nil
	}

	// Check for go.mod (Go)
	if fileExists(filepath.Join(projectDir, "go.mod")) {
		return scaffolding.LanguageGo, "chi", nil
	}

	// Check for requirements.txt (Python)
	if fileExists(filepath.Join(projectDir, "requirements.txt")) {
		return scaffolding.LanguagePythonFastAPI, "fastapi", nil
	}

	return "", "", fmt.Errorf("could not detect project type")
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func showTestInstructions(language scaffolding.LanguageType) {
	fmt.Println("📚 Next Steps:")

	switch language {
	case scaffolding.LanguageNodeExpress, scaffolding.LanguageNodeFastify:
		fmt.Println("   1. Install dependencies: npm install --save-dev jest supertest @types/jest @types/supertest")
		fmt.Println("   2. Run tests: npm test")
		fmt.Println("   3. Watch mode: npm test -- --watch")
	case scaffolding.LanguageGo:
		fmt.Println("   1. Install testify: go get github.com/stretchr/testify")
		fmt.Println("   2. Run tests: go test ./tests/...")
		fmt.Println("   3. With coverage: go test -cover ./tests/...")
	case scaffolding.LanguagePythonFastAPI:
		fmt.Println("   1. Install pytest: pip install pytest pytest-asyncio")
		fmt.Println("   2. Run tests: pytest tests/")
		fmt.Println("   3. With coverage: pytest --cov=app tests/")
	}
}
