package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/RichGod93/bffgen/internal/scaffolding"
	"github.com/spf13/cobra"
)

var addInfraCmd = &cobra.Command{
	Use:   "add-infra",
	Short: "Add infrastructure to existing project",
	Long:  `Add infrastructure files (CI/CD, Docker, health checks) to an existing BFF project.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get flags
		includeCI, _ := cmd.Flags().GetBool("ci")
		includeDocker, _ := cmd.Flags().GetBool("docker")
		includeHealth, _ := cmd.Flags().GetBool("health")
		includeCompose, _ := cmd.Flags().GetBool("compose")
		includeAll, _ := cmd.Flags().GetBool("all")

		if includeAll {
			includeCI = true
			includeDocker = true
			includeHealth = true
			includeCompose = true
		}

		if !includeCI && !includeDocker && !includeHealth && !includeCompose {
			fmt.Println("❌ No infrastructure options specified")
			fmt.Println("💡 Use flags: --ci, --docker, --health, --compose, or --all")
			os.Exit(1)
		}

		if err := addInfrastructure(includeCI, includeDocker, includeHealth, includeCompose); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to add infrastructure: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ Infrastructure added successfully!")
	},
}

func init() {
	addInfraCmd.Flags().Bool("ci", false, "Add CI/CD pipeline (GitHub Actions)")
	addInfraCmd.Flags().Bool("docker", false, "Add Dockerfile and .dockerignore")
	addInfraCmd.Flags().Bool("health", false, "Add health check endpoints")
	addInfraCmd.Flags().Bool("compose", false, "Add docker-compose.yml")
	addInfraCmd.Flags().Bool("all", false, "Add all infrastructure")

	rootCmd.AddCommand(addInfraCmd)
}

func addInfrastructure(includeCI, includeDocker, includeHealth, includeCompose bool) error {
	// Detect project type
	projectType := detectProjectType()

	if projectType == "unknown" {
		fmt.Println("❌ No BFF project found in current directory")
		fmt.Println("💡 Run 'bffgen init <project-name>' first or navigate to a BFF project directory")
		return fmt.Errorf("no project configuration found")
	}

	fmt.Printf("🔧 Adding infrastructure to %s project\n", projectType)
	fmt.Println()

	filesAdded := 0

	// Add CI/CD pipeline
	if includeCI {
		if err := addCIInfrastructure(projectType); err != nil {
			fmt.Printf("⚠️  Warning: Failed to add CI/CD: %v\n", err)
		} else {
			fmt.Println("✅ Added CI/CD pipeline")
			filesAdded++
		}
	}

	// Add Docker
	if includeDocker {
		if err := addDockerInfrastructure(projectType); err != nil {
			fmt.Printf("⚠️  Warning: Failed to add Docker: %v\n", err)
		} else {
			fmt.Println("✅ Added Dockerfile and .dockerignore")
			filesAdded++
		}
	}

	// Add health checks
	if includeHealth {
		if err := addHealthCheckInfrastructure(projectType); err != nil {
			fmt.Printf("⚠️  Warning: Failed to add health checks: %v\n", err)
		} else {
			fmt.Println("✅ Added health check endpoints")
			filesAdded++
		}
	}

	// Add docker-compose
	if includeCompose {
		if err := addComposeInfrastructure(projectType); err != nil {
			fmt.Printf("⚠️  Warning: Failed to add docker-compose: %v\n", err)
		} else {
			fmt.Println("✅ Added docker-compose.yml")
			filesAdded++
		}
	}

	fmt.Println()
	fmt.Printf("📁 Added %d infrastructure files\n", filesAdded)

	return nil
}

func addCIInfrastructure(projectType string) error {
	// Check if .github/workflows already exists
	if _, err := os.Stat(".github/workflows/ci.yml"); err == nil {
		fmt.Println("⚠️  CI pipeline already exists, skipping")
		return nil
	}

	// Determine language type
	langType := getLangTypeFromProjectType(projectType)

	// Use existing infrastructure generator
	return generateCIWorkflow(".", langType, true)
}

func addDockerInfrastructure(projectType string) error {
	// Check if Dockerfile already exists
	if _, err := os.Stat("Dockerfile"); err == nil {
		fmt.Println("⚠️  Dockerfile already exists, skipping")
		return nil
	}

	langType := getLangTypeFromProjectType(projectType)
	framework := detectFrameworkFromProject(projectType)

	return generateDockerfile(".", langType, framework, 8080)
}

func addHealthCheckInfrastructure(projectType string) error {
	if _, err := os.Stat("internal/health/health.go"); err == nil {
		fmt.Println("⚠️  Health checks already exist, skipping")
		return nil
	}
	if _, err := os.Stat("src/utils/health.js"); err == nil {
		fmt.Println("⚠️  Health checks already exist, skipping")
		return nil
	}

	langType := getLangTypeFromProjectType(projectType)
	framework := detectFrameworkFromProject(projectType)

	return generateHealthChecks(".", langType, framework, []BackendService{})
}

func addComposeInfrastructure(projectType string) error {
	// Check if docker-compose.yml already exists
	if _, err := os.Stat("docker-compose.yml"); err == nil {
		fmt.Println("⚠️  docker-compose.yml already exists, skipping")
		return nil
	}

	langType := getLangTypeFromProjectType(projectType)

	return generateDockerCompose(".", langType, []BackendService{}, 8080)
}

func getLangTypeFromProjectType(projectType string) scaffolding.LanguageType {
	switch projectType {
	case "nodejs":
		// Try to detect framework
		if data, err := os.ReadFile("package.json"); err == nil {
			content := string(data)
			if strings.Contains(content, "fastify") {
				return scaffolding.LanguageNodeFastify
			}
		}
		return scaffolding.LanguageNodeExpress
	case "go":
		return scaffolding.LanguageGo
	default:
		return scaffolding.LanguageGo
	}
}

func detectFrameworkFromProject(projectType string) string {
	if projectType == "nodejs" {
		// Try to detect from package.json dependencies
		if data, err := os.ReadFile("package.json"); err == nil {
			content := string(data)
			if strings.Contains(content, "fastify") {
				return "fastify"
			}
			return "express"
		}
	}
	// For Go, try to detect from go.mod
	if data, err := os.ReadFile("go.mod"); err == nil {
		content := string(data)
		if strings.Contains(content, "echo") {
			return "echo"
		} else if strings.Contains(content, "fiber") {
			return "fiber"
		}
	}
	return "chi" // Default for Go
}
