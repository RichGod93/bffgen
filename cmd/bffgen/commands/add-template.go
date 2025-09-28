package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/RichGod93/bffgen/internal/types"
	"github.com/RichGod93/bffgen/internal/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var addTemplateCmd = &cobra.Command{
	Use:   "add-template [template-name]",
	Short: "Add a predefined template (auth, ecommerce, content)",
	Long:  `Add a predefined template to your BFF configuration.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var templateName string
		if len(args) > 0 {
			templateName = args[0]
		} else {
			templateName = selectTemplate()
		}

		if err := addTemplate(templateName); err != nil {
			fmt.Fprintf(os.Stderr, "Error adding template: %v\n", err)
			os.Exit(1)
		}
	},
}

func selectTemplate() string {
	fmt.Println("🔧 Choose a template:")
	fmt.Println("  1) Auth (login, register, refresh token)")
	fmt.Println("  2) E-commerce (products, cart, checkout)")
	fmt.Println("  3) Content (posts, comments, likes)")
	fmt.Print("✔ Select template (1-3): ")

	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		return "auth"
	case "2":
		return "ecommerce"
	case "3":
		return "content"
	default:
		fmt.Println("❌ Invalid selection, defaulting to auth")
		return "auth"
	}
}

func addTemplate(templateName string) error {
	// Check if config file exists
	if _, err := os.Stat("bff.config.yaml"); os.IsNotExist(err) {
		fmt.Println("❌ bff.config.yaml not found in current directory")
		fmt.Println("💡 Run 'bffgen init <project-name>' first or navigate to a BFF project directory")
		return fmt.Errorf("config file not found")
	}

	fmt.Printf("🔧 Adding template: %s\n", templateName)
	fmt.Println()

	// Load template file
	templatePath := filepath.Join("internal", "templates", templateName+".yaml")
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		// Try relative to current working directory
		templatePath = filepath.Join("internal", "templates", templateName+".yaml")
		if _, err := os.Stat(templatePath); os.IsNotExist(err) {
			return fmt.Errorf("template file not found: %s", templatePath)
		}
	}

	templateData, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	var templateConfig types.BFFConfig
	if err := yaml.Unmarshal(templateData, &templateConfig); err != nil {
		return fmt.Errorf("failed to parse template file: %w", err)
	}

	// Load existing config
	config, err := utils.LoadConfig("bff.config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize services map if nil
	if config.Services == nil {
		config.Services = make(map[string]types.Service)
	}

	// Merge template services into existing config
	mergedCount := 0
	for serviceName, templateService := range templateConfig.Services {
		if existingService, exists := config.Services[serviceName]; exists {
			// Service exists, ask if user wants to merge
			fmt.Printf("⚠️  Service '%s' already exists. Merge endpoints? (y/N): ", serviceName)
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response == "y" || response == "yes" {
				// Merge endpoints
				existingService.Endpoints = append(existingService.Endpoints, templateService.Endpoints...)
				config.Services[serviceName] = existingService
				mergedCount += len(templateService.Endpoints)
				fmt.Printf("✅ Merged %d endpoints into existing service '%s'\n", len(templateService.Endpoints), serviceName)
			} else {
				fmt.Printf("⏭️  Skipped service '%s'\n", serviceName)
			}
		} else {
			// New service, add it
			config.Services[serviceName] = templateService
			mergedCount += len(templateService.Endpoints)
			fmt.Printf("✅ Added service '%s' with %d endpoints\n", serviceName, len(templateService.Endpoints))
		}
	}

	// Save updated config
	if err := utils.SaveConfig(config, "bff.config.yaml"); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Printf("✅ Template '%s' applied successfully! Added %d total endpoints.\n", templateName, mergedCount)
	fmt.Println("💡 Run 'bffgen generate' to update your Go code")

	return nil
}
