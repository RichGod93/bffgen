package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/RichGod93/bffgen/v3/internal/aggregators"
	"github.com/RichGod93/bffgen/v3/internal/utils"
	"github.com/spf13/cobra"
)

var addAggregatorCmd = &cobra.Command{
	Use:   "add-aggregator",
	Short: "Add a data aggregator to your BFF",
	Long: `Add a data aggregator that combines data from multiple backend services.
	
Examples:
  bffgen add-aggregator                    # Interactive selection
  bffgen add-aggregator user-dashboard     # Quick add user dashboard
  bffgen add-aggregator ecommerce-catalog  # Quick add ecommerce catalog`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var aggregatorName string
		if len(args) > 0 {
			aggregatorName = args[0]
		} else {
			aggregatorName = selectAggregator()
		}

		if err := addAggregator(aggregatorName); err != nil {
			fmt.Fprintf(os.Stderr, "Error adding aggregator: %v\n", err)
			os.Exit(1)
		}
	},
}

func selectAggregator() string {
	fmt.Println("🔧 Choose an aggregator:")
	fmt.Println("  1) User Dashboard (combines user, orders, preferences)")
	fmt.Println("  2) E-commerce Catalog (combines products, inventory, cart)")
	fmt.Println("  3) Custom aggregator")
	fmt.Print("✔ Select aggregator (1-3): ")

	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		return "user-dashboard"
	case "2":
		return "ecommerce-catalog"
	case "3":
		return "custom"
	default:
		fmt.Println("❌ Invalid selection, defaulting to user-dashboard")
		return "user-dashboard"
	}
}

func addAggregator(aggregatorName string) error {
	fmt.Printf("🔧 Adding aggregator: %s\n", aggregatorName)
	fmt.Println()

	// Check if config file exists
	if _, err := os.Stat("bff.config.yaml"); os.IsNotExist(err) {
		fmt.Println("❌ bff.config.yaml not found in current directory")
		fmt.Println("💡 Run 'bffgen init <project-name>' first or navigate to a BFF project directory")
		return fmt.Errorf("config file not found")
	}

	// Load existing config to verify we're in a BFF project
	_, err := utils.LoadConfig("bff.config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create aggregator configuration
	var aggregatorConfig aggregators.AggregatorConfig
	switch aggregatorName {
	case "user-dashboard":
		aggregatorConfig = aggregators.AggregatorConfig{
			Name:        "user-dashboard",
			Path:        "/api/user-dashboard/:id",
			Description: "Aggregates user, orders, and preferences data",
			Services:    []string{"users", "orders", "preferences"},
			Cache: aggregators.CacheConfig{
				Enabled: true,
				TTL:     5 * time.Minute,
				MaxSize: 1000,
			},
			Timeout: 30 * time.Second,
			Retries: 3,
			Headers: map[string]string{
				"Accept": "application/json",
			},
			Enabled: true,
		}
	case "ecommerce-catalog":
		aggregatorConfig = aggregators.AggregatorConfig{
			Name:        "ecommerce-catalog",
			Path:        "/api/catalog/:category",
			Description: "Aggregates products, inventory, and cart data",
			Services:    []string{"products", "inventory", "cart"},
			Cache: aggregators.CacheConfig{
				Enabled: true,
				TTL:     10 * time.Minute,
				MaxSize: 500,
			},
			Timeout: 20 * time.Second,
			Retries: 2,
			Headers: map[string]string{
				"Accept": "application/json",
			},
			Enabled: true,
		}
	case "custom":
		aggregatorConfig = createCustomAggregator()
	default:
		return fmt.Errorf("unknown aggregator: %s", aggregatorName)
	}

	// Save aggregator configuration
	aggregatorConfigs := &aggregators.AggregatorConfigs{
		Aggregators: []aggregators.AggregatorConfig{aggregatorConfig},
		Global: aggregators.GlobalConfig{
			DefaultTimeout: 30 * time.Second,
			DefaultRetries: 3,
			CacheEnabled:   true,
			CacheTTL:       5 * time.Minute,
		},
	}

	if err := aggregators.SaveAggregatorConfigs("aggregators.json", aggregatorConfigs); err != nil {
		return fmt.Errorf("failed to save aggregator config: %w", err)
	}

	fmt.Println()
	fmt.Printf("✅ Aggregator '%s' added successfully!\n", aggregatorName)
	fmt.Printf("📁 Configuration saved to: aggregators.json\n")
	fmt.Println("💡 Run 'bffgen generate' to update your Go code with aggregator routes")

	return nil
}

func createCustomAggregator() aggregators.AggregatorConfig {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("✔ Aggregator name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("✔ Path (e.g., /api/custom/:id): ")
	path, _ := reader.ReadString('\n')
	path = strings.TrimSpace(path)

	fmt.Print("✔ Description: ")
	description, _ := reader.ReadString('\n')
	description = strings.TrimSpace(description)

	fmt.Print("✔ Services (comma-separated): ")
	servicesInput, _ := reader.ReadString('\n')
	servicesInput = strings.TrimSpace(servicesInput)
	services := strings.Split(servicesInput, ",")
	for i, service := range services {
		services[i] = strings.TrimSpace(service)
	}

	return aggregators.AggregatorConfig{
		Name:        name,
		Path:        path,
		Description: description,
		Services:    services,
		Cache: aggregators.CacheConfig{
			Enabled: true,
			TTL:     5 * time.Minute,
			MaxSize: 1000,
		},
		Timeout: 30 * time.Second,
		Retries: 3,
		Headers: map[string]string{
			"Accept": "application/json",
		},
		Enabled: true,
	}
}
