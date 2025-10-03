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
		
		languageType, framework, backendServices, err := initializeProject(projectName, languageType, framework)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ BFF project '%s' initialized successfully!\n", projectName)

		showBackendConfigSummary(backendServices)
		showSetupInstructions(backendServices, projectName)
		fmt.Printf("📁 Navigate to the project: cd %s\n", projectName)
		fmt.Printf("🚀 Start development server: bffgen dev\n")

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
}
