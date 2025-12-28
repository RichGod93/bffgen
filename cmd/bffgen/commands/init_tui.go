package commands

import (
	"fmt"
	"os"

	"github.com/RichGod93/bffgen/internal/scaffolding"
	"github.com/RichGod93/bffgen/internal/tui"
	"github.com/RichGod93/bffgen/internal/types"
	tea "github.com/charmbracelet/bubbletea"
)

// runTUI launches the interactive TUI and returns collected configuration
func runTUI(projectName string) (scaffolding.LanguageType, string, []string, string, []types.BackendService, string, bool) {
	// Check if running in a TTY
	if !isTerminal() {
		fmt.Println("⚠️  TUI mode requires a terminal. Falling back to prompts...")
		return scaffolding.LanguageGo, "chi", nil, "", nil, "", false
	}

	model := tui.NewModel(projectName)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("⚠️  TUI error: %v. Falling back to prompts...\n", err)
		return scaffolding.LanguageGo, "chi", nil, "", nil, "", false
	}

	// Extract results from final model
	if m, ok := finalModel.(tui.Model); ok {
		return m.GetResults()
	}

	return scaffolding.LanguageGo, "chi", nil, "", nil, "", false
}

// isTerminal checks if we're running in a real terminal
func isTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// initializeProjectWithTUIConfig initializes project using TUI-collected config
func initializeProjectWithTUIConfig(
	projectName string,
	langType scaffolding.LanguageType,
	framework string,
	corsOrigins []string,
	services []types.BackendService,
	opts ProjectOptions,
) error {
	// Create project directory
	if err := os.MkdirAll(projectName, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create project directories
	if err := createProjectDirectories(projectName, langType); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Generate CORS config
	corsConfig := generateCORSConfigWithLang(corsOrigins, framework, langType)

	// Copy auth package for Go
	if langType == scaffolding.LanguageGo {
		if err := copyAuthPackage(projectName); err != nil {
			fmt.Printf("⚠️  Warning: Could not copy auth package: %v\n", err)
		}
	}

	// Create dependency files
	if err := createDependencyFilesWithOptions(projectName, langType, framework, opts); err != nil {
		return fmt.Errorf("failed to create dependency files: %w", err)
	}

	// Create main file
	if err := createMainFileWithOptions(projectName, langType, framework, corsConfig, services, opts); err != nil {
		return fmt.Errorf("failed to create main file: %w", err)
	}

	// Language-specific setup
	if langType == scaffolding.LanguagePythonFastAPI {
		// Python-specific files
		if err := createFastAPIConfig(projectName); err != nil {
			return fmt.Errorf("failed to create config.py: %w", err)
		}
		if err := createFastAPIDependencies(projectName); err != nil {
			return fmt.Errorf("failed to create dependencies.py: %w", err)
		}
		if err := createPythonEnvFile(projectName); err != nil {
			return fmt.Errorf("failed to create .env: %w", err)
		}
		if err := createPythonGitignore(projectName); err != nil {
			return fmt.Errorf("failed to create .gitignore: %w", err)
		}
		if err := createPythonLogger(projectName); err != nil {
			return fmt.Errorf("failed to create logger: %w", err)
		}
		if err := createPythonCacheManager(projectName); err != nil {
			fmt.Printf("⚠️  Warning: Could not create cache manager: %v\n", err)
		}
		if err := createPythonCircuitBreaker(projectName); err != nil {
			fmt.Printf("⚠️  Warning: Could not create circuit breaker: %v\n", err)
		}
		if err := createPythonMiddleware(projectName); err != nil {
			return fmt.Errorf("failed to create middleware: %w", err)
		}
		if err := createPythonTestFiles(projectName, opts); err != nil {
			fmt.Printf("⚠️  Warning: Could not create test files: %v\n", err)
		}
		if err := createPythonBFFGenConfig(projectName, opts); err != nil {
			return fmt.Errorf("failed to create bffgen.config.py.json: %w", err)
		}
		if err := createPythonSetupScript(projectName, opts); err != nil {
			return fmt.Errorf("failed to create setup.sh: %w", err)
		}
		if err := createPythonREADME(projectName, opts); err != nil {
			return fmt.Errorf("failed to create README.md: %w", err)
		}
	} else {
		// Create BFF config and README
		if err := createBFFConfig(projectName, services); err != nil {
			return fmt.Errorf("failed to create bff.config.yaml: %w", err)
		}
		if err := createReadme(projectName, langType); err != nil {
			return fmt.Errorf("failed to create README.md: %w", err)
		}
	}

	// Generate infrastructure files based on flags
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
		if err := generateHealthChecks(projectName, langType, framework, services); err != nil {
			fmt.Printf("⚠️  Warning: Failed to generate health checks: %v\n", err)
		} else {
			fmt.Println("✅ Generated enhanced health check endpoints")
		}

		if err := generateGracefulShutdown(projectName, langType, framework); err != nil {
			fmt.Printf("⚠️  Warning: Failed to generate graceful shutdown: %v\n", err)
		} else {
			fmt.Println("✅ Generated graceful shutdown handler")
		}
	}

	if opts.IncludeCompose {
		if err := generateDockerCompose(projectName, langType, services, 8080); err != nil {
			fmt.Printf("⚠️  Warning: Failed to generate docker-compose: %v\n", err)
		} else {
			fmt.Println("✅ Generated development docker-compose.yml")
		}
	}

	return nil
}
