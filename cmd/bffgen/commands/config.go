package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/RichGod93/bffgen/internal/utils"
	"github.com/spf13/cobra"
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

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configResetCmd)
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
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
