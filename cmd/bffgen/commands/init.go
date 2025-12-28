package commands

import (
	"fmt"
	"os"

	"github.com/RichGod93/bffgen/internal/scaffolding"
	"github.com/RichGod93/bffgen/internal/templates"
	"github.com/RichGod93/bffgen/internal/types"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new BFF project",
	Long:  `Initialize a new BFF project with support for Go, Node.js (Express/Fastify/Apollo/Yoga), Python (FastAPI), and GraphQL.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]

		// Validate project name early
		validator := NewProjectNameValidator()
		if err := validator.Validate(projectName); err != nil {
			ValidateError(err, "project-name")
			os.Exit(1)
		}

		// Get flags
		langFlag, _ := cmd.Flags().GetString("lang")
		runtimeFlag, _ := cmd.Flags().GetString("runtime")
		frameworkFlag, _ := cmd.Flags().GetString("framework")
		middlewareFlag, _ := cmd.Flags().GetString("middleware")
		controllerType, _ := cmd.Flags().GetString("controller-type")
		skipTests, _ := cmd.Flags().GetBool("skip-tests")
		skipDocs, _ := cmd.Flags().GetBool("skip-docs")
		noTUI, _ := cmd.Flags().GetBool("no-tui")
		templateFlag, _ := cmd.Flags().GetString("template")

		// Python-specific flags
		pkgManager, _ := cmd.Flags().GetString("pkg-manager")
		asyncFlag, _ := cmd.Flags().GetBool("async")

		// Infrastructure flags
		includeCI, _ := cmd.Flags().GetBool("include-ci")
		includeDocker, _ := cmd.Flags().GetBool("include-docker")
		includeHealth, _ := cmd.Flags().GetBool("include-health")
		includeCompose, _ := cmd.Flags().GetBool("include-compose")
		includeAllInfra, _ := cmd.Flags().GetBool("include-all-infra")

		// If include-all-infra is set, enable all infrastructure features
		if includeAllInfra {
			includeCI = true
			includeDocker = true
			includeHealth = true
			includeCompose = true
		}

		languageType := scaffolding.LanguageGo
		framework := "chi"
		var backendServices []types.BackendService
		var corsOriginsList []string
		var architecture string // Used by TUI
		var routeOption string  // Used by TUI
		_ = architecture        // Suppress unused warning when not using TUI
		_ = routeOption         // Suppress unused warning when not using TUI

		// Check if template flag is provided
		if templateFlag != "" {
			if err := initializeFromTemplate(projectName, templateFlag); err != nil {
				fmt.Printf("Error initializing from template: %v\n", err)
				os.Exit(1)
			}
			return
		}

		// Use TUI by default when no language/runtime flags are set and --no-tui is not specified
		useTUI := langFlag == "" && runtimeFlag == "" && !noTUI
		if useTUI {
			var ok bool
			languageType, framework, corsOriginsList, architecture, backendServices, routeOption, ok = runTUI(projectName)
			if !ok {
				fmt.Println("\n⚠️  Project initialization cancelled.")
				os.Exit(0)
			}
		}

		// Determine language from flags if not using TUI
		if !useTUI {
			if langFlag != "" {
				if !scaffolding.IsValidLanguage(langFlag) {
					HandleError(fmt.Errorf("invalid language '%s'. Supported: go, nodejs-express, nodejs-fastify, nodejs-apollo, nodejs-yoga, python-fastapi, go-graphql", langFlag), "language validation")
				}
				languageType = scaffolding.LanguageType(langFlag)
				config := scaffolding.GetLanguageConfig(languageType)
				framework = config.Framework
			} else if runtimeFlag != "" {
				if !scaffolding.IsValidLanguage(runtimeFlag) {
					HandleError(fmt.Errorf("invalid runtime '%s'. Supported: go, nodejs-express, nodejs-fastify, nodejs-apollo, nodejs-yoga, python-fastapi, go-graphql", runtimeFlag), "runtime validation")
				}
				languageType = scaffolding.LanguageType(runtimeFlag)
				config := scaffolding.GetLanguageConfig(languageType)
				framework = config.Framework
			}

			// Override framework if specified
			if frameworkFlag != "" {
				framework = frameworkFlag
			}
		}

		// Prepare project options
		opts := ProjectOptions{
			MiddlewareFlag:   middlewareFlag,
			ControllerType:   controllerType,
			SkipTests:        skipTests,
			SkipDocs:         skipDocs,
			LanguageExplicit: langFlag != "" || runtimeFlag != "" || useTUI,
			// Infrastructure options
			IncludeCI:      includeCI,
			IncludeDocker:  includeDocker,
			IncludeHealth:  includeHealth,
			IncludeCompose: includeCompose,
			// Python-specific options
			PkgManager:     pkgManager,
			AsyncEndpoints: asyncFlag,
		}

		// Initialize project (skip prompt-based config if TUI was used)
		var err error
		if useTUI {
			// Use TUI-collected configuration
			err = initializeProjectWithTUIConfig(projectName, languageType, framework, corsOriginsList, backendServices, opts)
		} else {
			// Use existing prompt-based flow
			languageType, framework, backendServices, err = initializeProjectWithOptions(projectName, languageType, framework, opts)
		}
		if err != nil {
			HandleError(err, "project initialization")
		}

		LogSuccess(fmt.Sprintf("BFF project '%s' initialized successfully", projectName))

		// Check tools and show personalized guidance
		showPostInitGuidance(projectName, string(languageType), framework, backendServices)

		showBackendConfigSummary(backendServices)
		showSetupInstructions(backendServices, projectName)

		if framework == "chi" || framework == "echo" {
			fmt.Println()
			fmt.Println("🔴 Redis Setup Required for Rate Limiting:")
			fmt.Println("   1. Install Redis: brew install redis (macOS) or apt install redis (Ubuntu)")
			fmt.Println("   2. Start Redis: redis-server")
			fmt.Println("   3. Set environment: export REDIS_URL=redis://localhost:6379")
		}

		fmt.Println()
		fmt.Println("🔐 Secure Authentication Setup:")
		fmt.Println("   1. Set encryption key: export ENCRYPTION_KEY=<base64-encoded-32-byte-key>")
		fmt.Println("   2. Set JWT secret: export JWT_SECRET=<base64-encoded-32-byte-key>")

		fmt.Println()
		fmt.Println("💡 To make bffgen available globally:")
		fmt.Println("   Or use: go install github.com/RichGod93/bffgen/cmd/bffgen")

		fmt.Println()
		fmt.Println("🔍 Run 'bffgen doctor' to check your project health")
	},
}

func init() {
	initCmd.Flags().StringP("lang", "l", "", "Programming language/runtime (go, nodejs-express, nodejs-fastify, nodejs-apollo, nodejs-yoga, python-fastapi, go-graphql)")
	initCmd.Flags().StringP("runtime", "r", "", "Programming language/runtime (go, nodejs-express, nodejs-fastify, nodejs-apollo, nodejs-yoga, python-fastapi, go-graphql) - alias for --lang")
	initCmd.Flags().StringP("framework", "f", "", "Framework (chi, echo, fiber for Go; express, fastify, apollo, yoga for Node.js; fastapi for Python; gqlgen for Go GraphQL)")
	initCmd.Flags().String("middleware", "", "Comma-separated list of middleware to include (validation,logger,requestId,all,none)")
	initCmd.Flags().String("controller-type", "both", "Controller type for Node.js projects (basic,aggregator,both)")
	initCmd.Flags().Bool("skip-tests", false, "Skip test file generation")
	initCmd.Flags().Bool("skip-docs", false, "Skip API documentation generation")
	initCmd.Flags().Bool("no-tui", false, "Disable interactive Terminal UI and use traditional prompts instead")
	initCmd.Flags().StringP("template", "t", "", "Use a project template (e.g., graphql-api, auth, ecommerce)")

	// Python-specific flags
	initCmd.Flags().String("pkg-manager", "pip", "Python package manager (pip or poetry)")
	initCmd.Flags().Bool("async", true, "Generate async FastAPI endpoints (default: true)")

	// Infrastructure scaffolding flags
	initCmd.Flags().Bool("include-ci", false, "Generate GitHub Actions CI/CD workflow")
	initCmd.Flags().Bool("include-docker", false, "Generate production Dockerfile and .dockerignore")
	initCmd.Flags().Bool("include-health", false, "Generate enhanced health checks with dependency checking")
	initCmd.Flags().Bool("include-compose", false, "Generate development docker-compose.yml")
	initCmd.Flags().Bool("include-all-infra", false, "Generate all infrastructure files (CI, Docker, health checks, docker-compose)")
}

// initializeFromTemplate creates a project from a template
func initializeFromTemplate(projectName, templateName string) error {
	fmt.Printf("\n🎨 Using template: %s\n", templateName)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Get template
	templatesDir, err := templates.GetDefaultTemplatesDir()
	if err != nil {
		return fmt.Errorf("failed to get templates directory: %w", err)
	}

	manager := templates.NewManager(templatesDir)
	template, err := manager.Get(templateName)
	if err != nil {
		return fmt.Errorf("template not found: %w", err)
	}

	// Map template language to scaffolding language type
	var langType scaffolding.LanguageType
	var framework string

	switch template.Language {
	case "nodejs-express":
		langType = scaffolding.LanguageNodeExpress
		framework = "express"
	case "nodejs-fastify":
		langType = scaffolding.LanguageNodeFastify
		framework = "fastify"
	case "go", "go-chi":
		langType = scaffolding.LanguageGo
		framework = "chi"
	case "go-echo":
		langType = scaffolding.LanguageGo
		framework = "echo"
	case "go-fiber":
		langType = scaffolding.LanguageGo
		framework = "fiber"
	case "python-fastapi":
		langType = scaffolding.LanguagePythonFastAPI
		framework = "fastapi"
	default:
		return fmt.Errorf("unsupported template language: %s", template.Language)
	}

	fmt.Printf("📦 Language: %s (%s)\n", langType, framework)
	fmt.Printf("🔧 Creating base BFF project...\n\n")

	// Create project directory structure
	if err := os.MkdirAll(projectName, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	if err := createProjectDirectories(projectName, langType); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Use template defaults for backend services and CORS
	backendServices := []types.BackendService{
		{Name: "api", BaseURL: "http://localhost:8000", Port: 8000},
	}
	corsOrigins := []string{"http://localhost:3000"}

	// Create CORS config
	corsConfig := generateCORSConfigWithLang(corsOrigins, framework, langType)

	// Copy auth package for Go projects
	if langType == scaffolding.LanguageGo {
		if err := copyAuthPackage(projectName); err != nil {
			fmt.Printf("⚠️  Warning: Could not copy auth package: %v\n", err)
		}
	}

	// Create dependency files
	opts := ProjectOptions{
		MiddlewareFlag:   "all",
		ControllerType:   "both",
		SkipTests:        false,
		SkipDocs:         false,
		LanguageExplicit: true,
		IncludeCI:        false,
		IncludeDocker:    false,
		IncludeHealth:    false,
		IncludeCompose:   false,
		PkgManager:       "pip",
		AsyncEndpoints:   true,
	}

	if err := createDependencyFilesWithOptions(projectName, langType, framework, opts); err != nil {
		return fmt.Errorf("failed to create dependency files: %w", err)
	}

	// Create main file
	if err := createMainFileWithOptions(projectName, langType, framework, corsConfig, backendServices, opts); err != nil {
		return fmt.Errorf("failed to create main file: %w", err)
	}

	fmt.Printf("\n🎨 Applying template customizations...\n")

	// Collect variable values
	vars := make(map[string]string)
	vars["PROJECT_NAME"] = projectName
	vars["PORT"] = "8080"

	// Use default values for variables
	for _, v := range template.Variables {
		if v.Default != "" {
			vars[v.Name] = template.GetVariableValue(v.Name, vars)
		}
	}

	// Overlay template files on top of base project
	scaffolder := templates.NewScaffolder(template, vars)

	// Prepare complex data for template execution (e.g. for GraphQL data sources)
	var templateServices []templates.BackendServiceData
	for _, svc := range backendServices {
		templateServices = append(templateServices, templates.BackendServiceData{
			Name:    svc.Name,
			BaseURL: svc.BaseURL,
			Port:    svc.Port,
		})
	}

	templateData := struct {
		ProjectName     string
		BackendServices []templates.BackendServiceData
		Variables       map[string]string
	}{
		ProjectName:     projectName,
		BackendServices: templateServices,
		Variables:       vars,
	}
	scaffolder.SetData(templateData)

	if err := scaffolder.OverlayFiles(projectName, "."); err != nil {
		return fmt.Errorf("failed to apply template: %w", err)
	}

	fmt.Println("\n✅ Project created successfully!")
	fmt.Printf("\n📋 Template: %s (v%s)\n", template.Name, template.Version)
	fmt.Printf("📁 Location: %s/\n", projectName)

	if len(template.Features) > 0 {
		fmt.Printf("\n✨ Features included:\n")
		for _, feature := range template.Features {
			fmt.Printf("   ✓ %s\n", feature)
		}
	}

	fmt.Printf("\n🚀 Next steps:\n")
	fmt.Printf("  1. cd %s\n", projectName)
	fmt.Printf("  2. Review configuration files\n")
	fmt.Printf("  3. Set up environment variables\n")
	fmt.Printf("  4. Install dependencies\n")
	fmt.Printf("  5. Start development\n\n")

	return nil
}
