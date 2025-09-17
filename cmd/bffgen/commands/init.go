package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new BFF project",
	Long:  `Initialize a new BFF project with chi router and config file.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		if err := initializeProject(projectName); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing project: %v\n", err)
			os.Exit(1)
		}
	fmt.Printf("✅ BFF project '%s' initialized successfully!\n", projectName)
	fmt.Printf("📁 Navigate to the project: cd %s\n", projectName)
	fmt.Printf("🚀 Start development server: bffgen dev\n")
	
	// Add global installation instructions
	fmt.Println()
	fmt.Println("💡 To make bffgen available globally:")
	fmt.Println("   macOS/Linux: sudo cp ../bffgen /usr/local/bin/")
	fmt.Println("   Windows: Add the bffgen directory to your PATH")
	fmt.Println("   Or use: go install github.com/richgodusen/bffgen/cmd/bffgen")
	},
}

func initializeProject(projectName string) error {
	// Create project directory
	if err := os.MkdirAll(projectName, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create subdirectories
	dirs := []string{
		filepath.Join(projectName, "internal", "routes"),
		filepath.Join(projectName, "internal", "aggregators"),
		filepath.Join(projectName, "internal", "templates"),
		filepath.Join(projectName, "cmd", "server"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Interactive prompts
	reader := bufio.NewReader(os.Stdin)

	// Framework selection
	fmt.Print("✔ Which framework? (chi/echo/fiber) [chi]: ")
	framework, _ := reader.ReadString('\n')
	framework = strings.TrimSpace(strings.ToLower(framework))
	if framework == "" {
		framework = "chi"
	}

	// Route configuration
	fmt.Println("✔ Configure routes now or later?")
	fmt.Println("   1) Define manually")
	fmt.Println("   2) Use a template")
	fmt.Println("   3) Skip for now")
	fmt.Print("✔ Select option (1-3) [3]: ")
	routeOption, _ := reader.ReadString('\n')
	routeOption = strings.TrimSpace(routeOption)
	if routeOption == "" {
		routeOption = "3"
	}

	// Copy template files only if user selected template option
	if routeOption == "2" {
		templateFiles := []string{"auth.yaml", "ecommerce.yaml", "content.yaml"}
		for _, templateFile := range templateFiles {
			srcPath := filepath.Join("internal", "templates", templateFile)
			dstPath := filepath.Join(projectName, "internal", "templates", templateFile)
			
			if _, err := os.Stat(srcPath); err == nil {
				if err := copyFile(srcPath, dstPath); err != nil {
					return fmt.Errorf("failed to copy template %s: %w", templateFile, err)
				}
			}
		}
	}

	// Create go.mod
	goModContent := fmt.Sprintf(`module %s

go 1.21

require (
	github.com/go-chi/chi/v5 v5.2.3
	github.com/go-chi/cors v1.2.2
	gopkg.in/yaml.v3 v3.0.1
)`, projectName)

	if err := os.WriteFile(filepath.Join(projectName, "go.mod"), []byte(goModContent), 0644); err != nil {
		return fmt.Errorf("failed to create go.mod: %w", err)
	}

	// Run go mod tidy to download dependencies
	if err := runCommand("go", "mod", "tidy"); err != nil {
		fmt.Printf("⚠️  Warning: Failed to run go mod tidy: %v\n", err)
		fmt.Println("   You may need to run 'go mod tidy' manually in the project directory")
	}

	// Create main.go based on framework
	var mainGoContent string
	switch framework {
	case "chi":
		mainGoContent = `package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "BFF server is running!")
	})

	// TODO: Add your aggregated routes here
	// Run 'bffgen add-route' or 'bffgen add-template' to add routes
	// Then run 'bffgen generate' to generate the code

	fmt.Println("🚀 BFF server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}`
	case "echo":
		mainGoContent = `package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "BFF server is running!")
	})

	// TODO: Add your aggregated routes here
	// Run 'bffgen add-route' or 'bffgen add-template' to add routes
	// Then run 'bffgen generate' to generate the code

	fmt.Println("🚀 BFF server starting on :8080")
	log.Fatal(e.Start(":8080"))
}`
	case "fiber":
		mainGoContent = `package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	app := fiber.New()

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New())

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("BFF server is running!")
	})

	// TODO: Add your aggregated routes here
	// Run 'bffgen add-route' or 'bffgen add-template' to add routes
	// Then run 'bffgen generate' to generate the code

	fmt.Println("🚀 BFF server starting on :8080")
	log.Fatal(app.Listen(":8080"))
}`
	default:
		return fmt.Errorf("unsupported framework: %s", framework)
	}

	if err := os.WriteFile(filepath.Join(projectName, "main.go"), []byte(mainGoContent), 0644); err != nil {
		return fmt.Errorf("failed to create main.go: %w", err)
	}

	// Create bff.config.yaml
	configContent := `# BFF Configuration
# Define your backend services and endpoints here

services:
  # Example service configuration
  # users:
  #   baseUrl: "http://localhost:4000/api"
  #   endpoints:
  #     - name: "getUser"
  #       path: "/users/:id"
  #       method: "GET"
  #       exposeAs: "/api/users/:id"
  #     - name: "createUser"
  #       path: "/users"
  #       method: "POST"
  #       exposeAs: "/api/users"

# Global settings
settings:
  port: 8080
  timeout: 30s
  retries: 3
`

	if err := os.WriteFile(filepath.Join(projectName, "bff.config.yaml"), []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create bff.config.yaml: %w", err)
	}

	// Create README.md
	readmeContent := fmt.Sprintf(`# %s

A Backend-for-Frontend (BFF) service generated by bffgen.

## Getting Started

1. Install dependencies:
   `+"```"+`bash
   go mod tidy
   `+"```"+`

2. Configure your backend services in bff.config.yaml

3. Run the development server:
   `+"```"+`bash
   go run main.go
   `+"```"+`

The server will start on http://localhost:8080

## Configuration

Edit bff.config.yaml to define your backend services and endpoints.

## Health Check

Visit http://localhost:8080/health to verify the server is running.

## Global Installation

To make bffgen available globally:
- macOS/Linux: sudo cp ../bffgen /usr/local/bin/
- Windows: Add the bffgen directory to your PATH
- Or use: go install github.com/richgodusen/bffgen/cmd/bffgen
`, projectName)

	if err := os.WriteFile(filepath.Join(projectName, "README.md"), []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}

	// Handle route configuration based on user choice
	switch routeOption {
	case "1":
		fmt.Println()
		fmt.Println("💡 To add routes manually, run:")
		fmt.Printf("   cd %s\n", projectName)
		fmt.Println("   bffgen add-route")
	case "2":
		fmt.Println()
		fmt.Println("💡 To add a template, run:")
		fmt.Printf("   cd %s\n", projectName)
		fmt.Println("   bffgen add-template")
	}

	return nil
}

// copyFile copies a file from src to dst
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

// runCommand runs a command in the project directory
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = "."
	return cmd.Run()
}
