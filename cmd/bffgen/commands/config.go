package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/RichGod93/bffgen/internal/types"
	"github.com/RichGod93/bffgen/internal/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage bffgen configuration",
	Long:  `Manage global bffgen configuration settings and view recent projects.`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current bffgen configuration settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := utils.LoadBFFGenConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("🔧 Current bffgen Configuration:")
		fmt.Println()

		fmt.Println("📋 Defaults:")
		fmt.Printf("   Framework: %s\n", config.Defaults.Framework)
		fmt.Printf("   CORS Origins: %s\n", strings.Join(config.Defaults.CORSOrigins, ", "))
		fmt.Printf("   JWT Secret: %s\n", maskSecret(config.Defaults.JWTSecret))
		fmt.Printf("   Redis URL: %s\n", config.Defaults.RedisURL)
		fmt.Printf("   Port: %d\n", config.Defaults.Port)
		fmt.Printf("   Route Option: %s\n", getRouteOptionName(config.Defaults.RouteOption))
		fmt.Println()

		if config.User.Name != "" || config.User.Email != "" || config.User.GitHub != "" {
			fmt.Println("👤 User Info:")
			if config.User.Name != "" {
				fmt.Printf("   Name: %s\n", config.User.Name)
			}
			if config.User.Email != "" {
				fmt.Printf("   Email: %s\n", config.User.Email)
			}
			if config.User.GitHub != "" {
				fmt.Printf("   GitHub: %s\n", config.User.GitHub)
			}
			if config.User.Company != "" {
				fmt.Printf("   Company: %s\n", config.User.Company)
			}
			fmt.Println()
		}

		if len(config.History.RecentProjects) > 0 {
			fmt.Println("📁 Recent Projects:")
			for i, project := range config.History.RecentProjects {
				marker := "  "
				if project == config.History.LastUsed {
					marker = "→ "
				}
				fmt.Printf("   %s%d. %s\n", marker, i+1, project)
			}
		}
	},
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset configuration to defaults",
	Long:  `Reset all configuration settings to their default values.`,
	Run: func(cmd *cobra.Command, args []string) {
		configPath, err := utils.GetConfigPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting config path: %v\n", err)
			os.Exit(1)
		}

		if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error removing config file: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ Configuration reset to defaults")
		fmt.Println("📁 Config file removed:", configPath)
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set configuration value",
	Long:  `Set a specific configuration value. Available keys: framework, cors_origins, jwt_secret, redis_url, port, route_option`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		config, err := utils.LoadBFFGenConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		switch key {
		case "framework":
			if value != "chi" && value != "echo" && value != "fiber" {
				fmt.Fprintf(os.Stderr, "Invalid framework: %s. Must be chi, echo, or fiber\n", value)
				os.Exit(1)
			}
			config.Defaults.Framework = value
		case "cors_origins":
			config.Defaults.CORSOrigins = strings.Split(value, ",")
		case "jwt_secret":
			config.Defaults.JWTSecret = value
		case "redis_url":
			config.Defaults.RedisURL = value
		case "port":
			var port int
			if _, err := fmt.Sscanf(value, "%d", &port); err != nil {
				fmt.Fprintf(os.Stderr, "Invalid port: %s\n", value)
				os.Exit(1)
			}
			config.Defaults.Port = port
		case "route_option":
			if value != "1" && value != "2" && value != "3" {
				fmt.Fprintf(os.Stderr, "Invalid route option: %s. Must be 1, 2, or 3\n", value)
				os.Exit(1)
			}
			config.Defaults.RouteOption = value
		default:
			fmt.Fprintf(os.Stderr, "Unknown key: %s\n", key)
			fmt.Println("Available keys: framework, cors_origins, jwt_secret, redis_url, port, route_option")
			os.Exit(1)
		}

		if err := utils.SaveBFFGenConfig(config); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Set %s = %s\n", key, value)
	},
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate project configuration",
	Long:  `Validate the current project's bff.config.yaml or bffgen.config.json against the schema.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := validateProjectConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Validation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Configuration is valid!")
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configResetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configValidateCmd)
	rootCmd.AddCommand(configCmd)
}

// validateProjectConfig validates the current project's configuration
func validateProjectConfig() error {
	// Detect project type
	projectType := detectProjectType()

	if projectType == "unknown" {
		fmt.Println("❌ No BFF project found in current directory")
		fmt.Println("💡 Run 'bffgen init <project-name>' first or navigate to a BFF project directory")
		return fmt.Errorf("no project configuration found")
	}

	fmt.Printf("🔍 Validating %s project configuration...\n", projectType)
	fmt.Println()

	if projectType == "nodejs" {
		return validateNodeJSConfig()
	}

	return validateGoConfig()
}

// validateGoConfig validates Go project configuration (bff.config.yaml)
func validateGoConfig() error {
	configPath := "bff.config.yaml"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("bff.config.yaml not found")
	}

	// Read and parse config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// First, validate YAML syntax
	var rawConfig map[string]interface{}
	if err := yaml.Unmarshal(data, &rawConfig); err != nil {
		return fmt.Errorf("invalid YAML syntax: %w", err)
	}

	fmt.Println("✅ YAML syntax is valid")

	// Load with bffgen config loader
	config, err := utils.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate structure
	errors := validateConfigStructure(config)
	if len(errors) > 0 {
		fmt.Println("❌ Configuration has errors:")
		for _, err := range errors {
			fmt.Printf("   - %s\n", err)
		}
		return fmt.Errorf("found %d validation errors", len(errors))
	}

	fmt.Println("✅ Configuration structure is valid")
	fmt.Printf("   - %d services configured\n", len(config.Services))

	totalEndpoints := 0
	for _, service := range config.Services {
		totalEndpoints += len(service.Endpoints)
	}
	fmt.Printf("   - %d total endpoints\n", totalEndpoints)

	return nil
}

// validateNodeJSConfig validates Node.js project configuration (bffgen.config.json)
func validateNodeJSConfig() error {
	configPath := "bffgen.config.json"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("bffgen.config.json not found")
	}

	// Read and parse config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// Validate JSON syntax
	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("invalid JSON syntax: %w", err)
	}

	fmt.Println("✅ JSON syntax is valid")

	// Validate structure
	errors := validateNodeJSConfigStructure(config)
	if len(errors) > 0 {
		fmt.Println("❌ Configuration has errors:")
		for _, err := range errors {
			fmt.Printf("   - %s\n", err)
		}
		return fmt.Errorf("found %d validation errors", len(errors))
	}

	fmt.Println("✅ Configuration structure is valid")

	// Count backends and endpoints
	backends, _ := config["backends"].([]interface{})
	fmt.Printf("   - %d backends configured\n", len(backends))

	totalEndpoints := 0
	for _, b := range backends {
		backend, ok := b.(map[string]interface{})
		if !ok {
			continue
		}
		endpoints, _ := backend["endpoints"].([]interface{})
		totalEndpoints += len(endpoints)
	}
	fmt.Printf("   - %d total endpoints\n", totalEndpoints)

	return nil
}

// validateConfigStructure validates the internal structure of Go config
func validateConfigStructure(config *types.BFFConfig) []string {
	var errors []string

	// Check if services exist
	if len(config.Services) == 0 {
		errors = append(errors, "No services configured")
	}

	// Validate each service
	seenEndpoints := make(map[string]bool)
	for serviceName, service := range config.Services {
		// Validate service name
		if serviceName == "" {
			errors = append(errors, "Service has empty name")
			continue
		}

		// Validate base URL
		if err := validateURL(service.BaseURL); err != nil {
			errors = append(errors, fmt.Sprintf("Service '%s': invalid baseURL: %v", serviceName, err))
		}

		// Validate endpoints
		if len(service.Endpoints) == 0 {
			errors = append(errors, fmt.Sprintf("Service '%s': no endpoints configured", serviceName))
		}

		for _, endpoint := range service.Endpoints {
			// Validate path
			if err := validatePath(endpoint.Path); err != nil {
				errors = append(errors, fmt.Sprintf("Service '%s', endpoint '%s': invalid path: %v", serviceName, endpoint.Name, err))
			}

			// Validate exposeAs
			if err := validatePath(endpoint.ExposeAs); err != nil {
				errors = append(errors, fmt.Sprintf("Service '%s', endpoint '%s': invalid exposeAs: %v", serviceName, endpoint.Name, err))
			}

			// Check for duplicate endpoints
			endpointKey := fmt.Sprintf("%s:%s", endpoint.Method, endpoint.ExposeAs)
			if seenEndpoints[endpointKey] {
				errors = append(errors, fmt.Sprintf("Duplicate endpoint: %s %s", endpoint.Method, endpoint.ExposeAs))
			}
			seenEndpoints[endpointKey] = true

			// Validate method
			validMethods := map[string]bool{
				"GET": true, "POST": true, "PUT": true, "DELETE": true, "PATCH": true, "HEAD": true, "OPTIONS": true,
			}
			if !validMethods[strings.ToUpper(endpoint.Method)] {
				errors = append(errors, fmt.Sprintf("Service '%s', endpoint '%s': invalid HTTP method: %s", serviceName, endpoint.Name, endpoint.Method))
			}
		}
	}

	return errors
}

// validateNodeJSConfigStructure validates the internal structure of Node.js config
func validateNodeJSConfigStructure(config map[string]interface{}) []string {
	var errors []string

	// Check project info
	project, ok := config["project"].(map[string]interface{})
	if !ok {
		errors = append(errors, "Missing 'project' section")
	} else {
		name, _ := project["name"].(string)
		if name == "" {
			errors = append(errors, "Project name is empty")
		}
	}

	// Check backends
	backends, ok := config["backends"].([]interface{})
	if !ok {
		errors = append(errors, "Missing or invalid 'backends' section")
		return errors
	}

	if len(backends) == 0 {
		errors = append(errors, "No backends configured")
	}

	// Validate each backend
	seenEndpoints := make(map[string]bool)
	for i, b := range backends {
		backend, ok := b.(map[string]interface{})
		if !ok {
			errors = append(errors, fmt.Sprintf("Backend %d: invalid structure", i))
			continue
		}

		// Validate backend name
		name, _ := backend["name"].(string)
		if name == "" {
			errors = append(errors, fmt.Sprintf("Backend %d: empty name", i))
			continue
		}

		// Validate baseUrl
		baseURL, _ := backend["baseUrl"].(string)
		if err := validateURL(baseURL); err != nil {
			errors = append(errors, fmt.Sprintf("Backend '%s': invalid baseUrl: %v", name, err))
		}

		// Validate endpoints
		endpoints, ok := backend["endpoints"].([]interface{})
		if !ok || len(endpoints) == 0 {
			errors = append(errors, fmt.Sprintf("Backend '%s': no endpoints configured", name))
			continue
		}

		for j, e := range endpoints {
			endpoint, ok := e.(map[string]interface{})
			if !ok {
				errors = append(errors, fmt.Sprintf("Backend '%s', endpoint %d: invalid structure", name, j))
				continue
			}

			// Validate path
			path, _ := endpoint["path"].(string)
			if err := validatePath(path); err != nil {
				errors = append(errors, fmt.Sprintf("Backend '%s', endpoint %d: invalid path: %v", name, j, err))
			}

			// Validate exposeAs
			exposeAs, _ := endpoint["exposeAs"].(string)
			if err := validatePath(exposeAs); err != nil {
				errors = append(errors, fmt.Sprintf("Backend '%s', endpoint %d: invalid exposeAs: %v", name, j, err))
			}

			// Check for duplicate endpoints
			method, _ := endpoint["method"].(string)
			endpointKey := fmt.Sprintf("%s:%s", method, exposeAs)
			if seenEndpoints[endpointKey] {
				errors = append(errors, fmt.Sprintf("Duplicate endpoint: %s %s", method, exposeAs))
			}
			seenEndpoints[endpointKey] = true

			// Validate method
			validMethods := map[string]bool{
				"GET": true, "POST": true, "PUT": true, "DELETE": true, "PATCH": true, "HEAD": true, "OPTIONS": true,
			}
			if !validMethods[strings.ToUpper(method)] {
				errors = append(errors, fmt.Sprintf("Backend '%s', endpoint %d: invalid HTTP method: %s", name, j, method))
			}
		}
	}

	return errors
}

// Helper functions
func maskSecret(secret string) string {
	if len(secret) <= 8 {
		return "***"
	}
	return secret[:4] + "***" + secret[len(secret)-4:]
}

func getRouteOptionName(option string) string {
	switch option {
	case "1":
		return "Define manually"
	case "2":
		return "Use a template"
	case "3":
		return "Skip for now"
	default:
		return "Unknown"
	}
}
