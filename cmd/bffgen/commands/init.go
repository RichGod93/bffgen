package commands

import (
	"fmt"
	"os"

	"github.com/RichGod93/bffgen/internal/scaffolding"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new BFF project",
	Long:  `Initialize a new BFF project with support for Go, Node.js (Express), and Node.js (Fastify).`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]

		// Get flags
		langFlag, _ := cmd.Flags().GetString("lang")
		runtimeFlag, _ := cmd.Flags().GetString("runtime")
		frameworkFlag, _ := cmd.Flags().GetString("framework")
		middlewareFlag, _ := cmd.Flags().GetString("middleware")
		controllerType, _ := cmd.Flags().GetString("controller-type")
		skipTests, _ := cmd.Flags().GetBool("skip-tests")
		skipDocs, _ := cmd.Flags().GetBool("skip-docs")

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

		// Determine language from flags
		if langFlag != "" {
			if !scaffolding.IsValidLanguage(langFlag) {
				fmt.Printf("❌ Invalid language: %s. Supported: go, nodejs-express, nodejs-fastify\n", langFlag)
				os.Exit(1)
			}
			languageType = scaffolding.LanguageType(langFlag)
			config := scaffolding.GetLanguageConfig(languageType)
			framework = config.Framework
		} else if runtimeFlag != "" {
			if !scaffolding.IsValidLanguage(runtimeFlag) {
				fmt.Printf("❌ Invalid runtime: %s. Supported: go, nodejs-express, nodejs-fastify\n", runtimeFlag)
				os.Exit(1)
			}
			languageType = scaffolding.LanguageType(runtimeFlag)
			config := scaffolding.GetLanguageConfig(languageType)
			framework = config.Framework
		}

		// Override framework if specified
		if frameworkFlag != "" {
			framework = frameworkFlag
		}

		// Prepare project options
		opts := ProjectOptions{
			MiddlewareFlag:   middlewareFlag,
			ControllerType:   controllerType,
			SkipTests:        skipTests,
			SkipDocs:         skipDocs,
			LanguageExplicit: langFlag != "" || runtimeFlag != "",
			// Infrastructure options
			IncludeCI:      includeCI,
			IncludeDocker:  includeDocker,
			IncludeHealth:  includeHealth,
			IncludeCompose: includeCompose,
		}

		languageType, framework, backendServices, err := initializeProjectWithOptions(projectName, languageType, framework, opts)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println()
		fmt.Printf("✅ BFF project '%s' initialized successfully!\n", projectName)
		fmt.Println()

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
	initCmd.Flags().StringP("lang", "l", "", "Programming language/runtime (go, nodejs-express, nodejs-fastify)")
	initCmd.Flags().StringP("runtime", "r", "", "Programming language/runtime (go, nodejs-express, nodejs-fastify) - alias for --lang")
	initCmd.Flags().StringP("framework", "f", "", "Framework (chi, echo, fiber for Go; express, fastify for Node.js)")
	initCmd.Flags().String("middleware", "", "Comma-separated list of middleware to include (validation,logger,requestId,all,none)")
	initCmd.Flags().String("controller-type", "both", "Controller type for Node.js projects (basic,aggregator,both)")
	initCmd.Flags().Bool("skip-tests", false, "Skip test file generation")
	initCmd.Flags().Bool("skip-docs", false, "Skip API documentation generation")

	// Infrastructure scaffolding flags
	initCmd.Flags().Bool("include-ci", false, "Generate GitHub Actions CI/CD workflow")
	initCmd.Flags().Bool("include-docker", false, "Generate production Dockerfile and .dockerignore")
	initCmd.Flags().Bool("include-health", false, "Generate enhanced health checks with dependency checking")
	initCmd.Flags().Bool("include-compose", false, "Generate development docker-compose.yml")
	initCmd.Flags().Bool("include-all-infra", false, "Generate all infrastructure files (CI, Docker, health checks, docker-compose)")
}
