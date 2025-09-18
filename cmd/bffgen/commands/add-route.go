package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/RichGod93/bffgen/internal/types"
	"github.com/RichGod93/bffgen/internal/utils"
	"github.com/spf13/cobra"
)

var addRouteCmd = &cobra.Command{
	Use:   "add-route",
	Short: "Interactively add a backend endpoint to your BFF",
	Long:  `Interactively add a backend endpoint to your BFF configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := addRoute(); err != nil {
			fmt.Fprintf(os.Stderr, "Error adding route: %v\n", err)
			os.Exit(1)
		}
	},
}

func addRoute() error {
	// Check if config file exists
	if _, err := os.Stat("bff.config.yaml"); os.IsNotExist(err) {
		fmt.Println("❌ bff.config.yaml not found in current directory")
		fmt.Println("💡 Run 'bffgen init <project-name>' first or navigate to a BFF project directory")
		return fmt.Errorf("config file not found")
	}

	fmt.Println("🔧 Adding a new route to your BFF")
	fmt.Println()

	// Load existing config
	config, err := utils.LoadConfig("bff.config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize services map if nil
	if config.Services == nil {
		config.Services = make(map[string]types.Service)
	}

	reader := bufio.NewReader(os.Stdin)

	// Get service name
	fmt.Print("✔ Service name: ")
	serviceName, _ := reader.ReadString('\n')
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	// Check if service exists, if not ask for base URL
	var baseURL string
	if service, exists := config.Services[serviceName]; exists {
		baseURL = service.BaseURL
		fmt.Printf("✔ Using existing service '%s' with base URL: %s\n", serviceName, baseURL)
	} else {
		fmt.Print("✔ Base URL: ")
		baseURL, _ = reader.ReadString('\n')
		baseURL = strings.TrimSpace(baseURL)
		if baseURL == "" {
			return fmt.Errorf("base URL cannot be empty")
		}
	}

	// Get endpoint details
	fmt.Print("✔ Endpoint name: ")
	endpointName, _ := reader.ReadString('\n')
	endpointName = strings.TrimSpace(endpointName)
	if endpointName == "" {
		return fmt.Errorf("endpoint name cannot be empty")
	}

	fmt.Print("✔ Path: ")
	path, _ := reader.ReadString('\n')
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	fmt.Print("✔ Method (GET/POST/PUT/DELETE/PATCH): ")
	method, _ := reader.ReadString('\n')
	method = strings.TrimSpace(strings.ToUpper(method))
	if method == "" {
		method = "GET"
	}

	fmt.Print("✔ Expose as: ")
	exposeAs, _ := reader.ReadString('\n')
	exposeAs = strings.TrimSpace(exposeAs)
	if exposeAs == "" {
		exposeAs = path
	}

	// Create new endpoint
	newEndpoint := types.Endpoint{
		Name:     endpointName,
		Path:     path,
		Method:   method,
		ExposeAs: exposeAs,
	}

	// Add endpoint to service
	if service, exists := config.Services[serviceName]; exists {
		service.Endpoints = append(service.Endpoints, newEndpoint)
		config.Services[serviceName] = service
	} else {
		config.Services[serviceName] = types.Service{
			BaseURL:   baseURL,
			Endpoints: []types.Endpoint{newEndpoint},
		}
	}

	// Save updated config
	if err := utils.SaveConfig("bff.config.yaml", config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Printf("✅ Added route: %s %s → %s%s\n", method, exposeAs, baseURL, path)
	fmt.Println("💡 Run 'bffgen generate' to update your Go code")

	return nil
}
