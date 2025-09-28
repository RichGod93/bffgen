package commands

// not tracked
import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/RichGod93/bffgen/internal/types"
	"github.com/RichGod93/bffgen/internal/utils"
	"github.com/spf13/cobra"
)

var postmanCmd = &cobra.Command{
	Use:   "postman",
	Short: "Generate Postman collection for BFF endpoints",
	Long:  `Generate a Postman collection JSON file based on your BFF configuration for easy API testing.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := generatePostmanCollection(); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating Postman collection: %v\n", err)
			os.Exit(1)
		}
	},
}

func generatePostmanCollection() error {
	fmt.Println("📮 Generating Postman collection from bff.config.yaml")
	fmt.Println()
	fmt.Println("🔍 Step 1: Checking for BFF configuration...")

	// Check if config file exists
	if _, err := os.Stat("bff.config.yaml"); os.IsNotExist(err) {
		fmt.Println("❌ bff.config.yaml not found in current directory")
		fmt.Println()
		fmt.Println("🚀 To get started with BFF generation:")
		fmt.Println("   1. Run: bffgen init <your-project-name>")
		fmt.Println("   2. Add services: bffgen add-template <template-name>")
		fmt.Println("   3. Generate BFF: bffgen generate")
		fmt.Println("   4. Generate Postman: bffgen postman")
		fmt.Println()
		fmt.Println("💡 Or navigate to an existing BFF project directory")
		return fmt.Errorf("config file not found")
	}

	fmt.Println("✅ Found bff.config.yaml")
	fmt.Println("🔍 Step 2: Loading and validating configuration...")

	// Load configuration
	config, err := utils.LoadConfig("bff.config.yaml")
	if err != nil {
		fmt.Println("❌ Failed to load bff.config.yaml")
		fmt.Println()
		fmt.Println("🔧 Common issues and solutions:")
		fmt.Println("   • Invalid YAML syntax - check indentation and quotes")
		fmt.Println("   • Missing required fields - ensure 'services' and 'settings' are defined")
		fmt.Println("   • File permissions - ensure the file is readable")
		fmt.Println()
		fmt.Println("📖 Example valid configuration:")
		fmt.Println("   services:")
		fmt.Println("     my-service:")
		fmt.Println("       baseUrl: \"http://localhost:3000\"")
		fmt.Println("       endpoints:")
		fmt.Println("         - name: \"Get Data\"")
		fmt.Println("           path: \"/api/data\"")
		fmt.Println("           method: \"GET\"")
		fmt.Println("           exposeAs: \"/data\"")
		fmt.Println("   settings:")
		fmt.Println("     port: 8080")
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("✅ Configuration loaded successfully")

	// Validate configuration
	if len(config.Services) == 0 {
		fmt.Println("⚠️  No services configured in bff.config.yaml")
		fmt.Println()
		fmt.Println("🔧 To add services to your BFF:")
		fmt.Println("   • Use templates: bffgen add-template auth")
		fmt.Println("   • Add custom routes: bffgen add-route")
		fmt.Println("   • Available templates: auth, content, ecommerce")
		fmt.Println()
		fmt.Println("📖 Example service configuration:")
		fmt.Println("   services:")
		fmt.Println("     user-service:")
		fmt.Println("       baseUrl: \"http://localhost:3001\"")
		fmt.Println("       endpoints:")
		fmt.Println("         - name: \"Get User\"")
		fmt.Println("           path: \"/api/users/{id}\"")
		fmt.Println("           method: \"GET\"")
		fmt.Println("           exposeAs: \"/users/{id}\"")
		return nil
	}

	// Validate each service
	fmt.Println("🔍 Step 3: Validating service configurations...")
	for serviceName, service := range config.Services {
		if service.BaseURL == "" {
			fmt.Printf("❌ Service '%s' missing baseUrl\n", serviceName)
			fmt.Println("💡 Add baseUrl to your service configuration")
			return fmt.Errorf("service %s missing baseUrl", serviceName)
		}

		if len(service.Endpoints) == 0 {
			fmt.Printf("⚠️  Service '%s' has no endpoints\n", serviceName)
			fmt.Println("💡 Add endpoints to your service configuration")
		}

		// Validate endpoints
		for i, endpoint := range service.Endpoints {
			if endpoint.Name == "" {
				fmt.Printf("❌ Service '%s' endpoint %d missing name\n", serviceName, i+1)
				return fmt.Errorf("service %s endpoint %d missing name", serviceName, i+1)
			}
			if endpoint.Path == "" {
				fmt.Printf("❌ Service '%s' endpoint '%s' missing path\n", serviceName, endpoint.Name)
				return fmt.Errorf("service %s endpoint %s missing path", serviceName, endpoint.Name)
			}
			if endpoint.Method == "" {
				fmt.Printf("❌ Service '%s' endpoint '%s' missing method\n", serviceName, endpoint.Name)
				return fmt.Errorf("service %s endpoint %s missing method", serviceName, endpoint.Name)
			}
			if endpoint.ExposeAs == "" {
				fmt.Printf("❌ Service '%s' endpoint '%s' missing exposeAs\n", serviceName, endpoint.Name)
				return fmt.Errorf("service %s endpoint %s missing exposeAs", serviceName, endpoint.Name)
			}
		}
	}

	fmt.Println("✅ All service configurations are valid")
	fmt.Println("🔍 Step 4: Generating Postman collection...")

	// Generate Postman collection
	if err := generateCollectionJSON(config); err != nil {
		fmt.Println("❌ Failed to generate Postman collection")
		fmt.Println()
		fmt.Println("🔧 Common issues and solutions:")
		fmt.Println("   • File write permissions - ensure directory is writable")
		fmt.Println("   • Disk space - ensure sufficient space for file creation")
		fmt.Println("   • Invalid characters in service/endpoint names")
		fmt.Println()
		fmt.Println("💡 Try running the command in a different directory or check permissions")
		return fmt.Errorf("failed to generate collection: %w", err)
	}

	fmt.Println("✅ Postman collection generated successfully!")
	fmt.Println("📁 Created file: bff-postman-collection.json")
	fmt.Println()
	fmt.Println("📋 Collection Summary:")

	// Count endpoints
	totalEndpoints := 0
	for serviceName, service := range config.Services {
		endpointCount := len(service.Endpoints)
		totalEndpoints += endpointCount
		fmt.Printf("   • %s service: %d endpoints\n", serviceName, endpointCount)
	}

	fmt.Printf("   • Total: %d endpoints across %d services\n", totalEndpoints, len(config.Services))
	fmt.Printf("   • BFF server port: %d\n", config.Settings.Port)
	fmt.Println()
	fmt.Println("🚀 Next Steps:")
	fmt.Println("   1. Import 'bff-postman-collection.json' into Postman")
	fmt.Println("   2. Start your BFF server: go run main.go")
	fmt.Println("   3. Test your endpoints using the collection")
	fmt.Println()
	fmt.Println("💡 Pro Tips:")
	fmt.Println("   • Use the 'baseUrl' variable to switch between environments")
	fmt.Println("   • The collection includes a health check endpoint")
	fmt.Println("   • All endpoints are pre-configured with proper headers")

	return nil
}

func generateCollectionJSON(config *types.BFFConfig) error {
	// Set default port if not specified
	if config.Settings.Port == 0 {
		config.Settings.Port = 8080
	}

	// Validate port range
	if config.Settings.Port < 1 || config.Settings.Port > 65535 {
		return fmt.Errorf("invalid port number %d - must be between 1 and 65535", config.Settings.Port)
	}

	// Create collection data structure
	collectionData := PostmanCollection{
		Info: CollectionInfo{
			Name:        "BFF API Collection",
			Description: "Generated collection for BFF endpoints",
			Schema:      "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
			PostmanID:   generatePostmanID(),
			Updated:     time.Now().Format("2006-01-02T15:04:05.000Z"),
		},
		Item:     []CollectionItem{},
		Variable: []CollectionVariable{},
	}

	// Add base URL variable
	collectionData.Variable = append(collectionData.Variable, CollectionVariable{
		Key:   "baseUrl",
		Value: fmt.Sprintf("http://localhost:%d", config.Settings.Port),
		Type:  "string",
	})

	// Generate items for each service
	for serviceName, service := range config.Services {
		// Validate service name for Postman compatibility
		if strings.ContainsAny(serviceName, " \t\n\r") {
			return fmt.Errorf("service name '%s' contains invalid characters - avoid spaces and special characters", serviceName)
		}

		serviceItem := CollectionItem{
			Name:        serviceName,
			Description: fmt.Sprintf("Endpoints for %s service", serviceName),
			Item:        []RequestItem{},
		}

		// Add endpoints for this service
		for _, endpoint := range service.Endpoints {
			// Validate endpoint name for Postman compatibility (allow spaces but not control characters)
			if strings.ContainsAny(endpoint.Name, "\t\n\r") {
				return fmt.Errorf("endpoint name '%s' in service '%s' contains invalid characters - avoid tabs, newlines, and carriage returns", endpoint.Name, serviceName)
			}

			// Validate HTTP method
			validMethods := map[string]bool{
				"GET": true, "POST": true, "PUT": true, "DELETE": true,
				"PATCH": true, "HEAD": true, "OPTIONS": true,
			}
			if !validMethods[strings.ToUpper(endpoint.Method)] {
				return fmt.Errorf("invalid HTTP method '%s' for endpoint '%s' in service '%s' - use GET, POST, PUT, DELETE, PATCH, HEAD, or OPTIONS", endpoint.Method, endpoint.Name, serviceName)
			}

			requestItem := RequestItem{
				Name: endpoint.Name,
				Request: Request{
					Method: strings.ToUpper(endpoint.Method),
					Header: []Header{
						{Key: "Content-Type", Value: "application/json", Type: "text"},
						{Key: "Accept", Value: "application/json", Type: "text"},
					},
					URL: URL{
						Raw:  fmt.Sprintf("{{baseUrl}}%s", endpoint.ExposeAs),
						Host: []string{"{{baseUrl}}"},
						Path: parsePath(endpoint.ExposeAs),
					},
					Description: fmt.Sprintf("Proxy to %s%s", service.BaseURL, endpoint.Path),
				},
				Response: []interface{}{},
			}

			serviceItem.Item = append(serviceItem.Item, requestItem)
		}

		collectionData.Item = append(collectionData.Item, serviceItem)
	}

	// Add health check endpoint
	healthItem := CollectionItem{
		Name:        "Health Check",
		Description: "BFF server health check",
		Item: []RequestItem{
			{
				Name: "Health Check",
				Request: Request{
					Method: "GET",
					Header: []Header{},
					URL: URL{
						Raw:  "{{baseUrl}}/health",
						Host: []string{"{{baseUrl}}"},
						Path: []string{"health"},
					},
					Description: "Check if BFF server is running",
				},
				Response: []interface{}{},
			},
		},
	}

	collectionData.Item = append(collectionData.Item, healthItem)

	// Marshal to JSON with error handling
	jsonData, err := json.MarshalIndent(collectionData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal collection to JSON: %w", err)
	}

	// Check if file already exists
	if _, err := os.Stat("bff-postman-collection.json"); err == nil {
		fmt.Println("⚠️  bff-postman-collection.json already exists - overwriting")
	}

	// Create output file with error handling
	file, err := os.Create("bff-postman-collection.json")
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Write data with error handling
	bytesWritten, err := file.Write(jsonData)
	if err != nil {
		return fmt.Errorf("failed to write collection data: %w", err)
	}

	// Verify file size
	if bytesWritten != len(jsonData) {
		return fmt.Errorf("incomplete write - expected %d bytes, wrote %d bytes", len(jsonData), bytesWritten)
	}

	return nil
}

// Helper functions
func generatePostmanID() string {
	return "bff-collection-" + fmt.Sprintf("%d", time.Now().Unix())
}

func parsePath(path string) []string {
	// Remove leading slash and split by /
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	if path == "" {
		return []string{}
	}
	// Split path by / but keep path segments as single elements
	parts := strings.Split(path, "/")
	result := []string{}
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// Collection data structures
type PostmanCollection struct {
	Info     CollectionInfo       `json:"info"`
	Item     []CollectionItem     `json:"item"`
	Variable []CollectionVariable `json:"variable"`
}

type CollectionInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Schema      string `json:"schema"`
	PostmanID   string `json:"_postman_id"`
	Updated     string `json:"updated"`
}

type CollectionItem struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Item        []RequestItem `json:"item"`
}

type RequestItem struct {
	Name     string        `json:"name"`
	Request  Request       `json:"request"`
	Response []interface{} `json:"response"`
}

type Request struct {
	Method      string   `json:"method"`
	Header      []Header `json:"header"`
	URL         URL      `json:"url"`
	Description string   `json:"description"`
}

type Header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

type URL struct {
	Raw  string   `json:"raw"`
	Host []string `json:"host"`
	Path []string `json:"path"`
}

type CollectionVariable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type"`
}
