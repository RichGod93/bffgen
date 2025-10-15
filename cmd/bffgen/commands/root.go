package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "bffgen",
	Short: "A CLI tool for generating Backend-for-Frontend (BFF) services",
	Long: `bffgen is a Go-based CLI tool that helps developers quickly scaffold 
Backend-for-Frontend (BFF) services. It enables teams to aggregate backend 
endpoints and expose them in a frontend-friendly way, with minimal setup.

Global Installation:
  macOS/Linux: sudo cp bffgen /usr/local/bin/
  Windows: Add the bffgen directory to your PATH
  Or use: go install github.com/RichGod93/bffgen/cmd/bffgen`,
	Version: "dev", // This will be set during build
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize global configuration
		if err := InitGlobalConfig(); err != nil {
			return fmt.Errorf("failed to initialize global config: %w", err)
		}

		// Update global config with flag values
		globalConfig.Verbose, _ = cmd.Flags().GetBool("verbose")
		globalConfig.NoColor, _ = cmd.Flags().GetBool("no-color")

		if configPath, _ := cmd.Flags().GetString("config-path"); configPath != "" {
			globalConfig.ConfigPath = configPath
		}

		// Set runtime override if specified
		if runtime, _ := cmd.Flags().GetString("runtime"); runtime != "" {
			globalConfig.RuntimeOverride = runtime
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute(version, buildTime, commit string) error {
	// Set version information
	rootCmd.Version = version

	// Add version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("bffgen version %s\n", version)
			fmt.Printf("Build time: %s\n", buildTime)
			fmt.Printf("Commit: %s\n", commit)
		},
	})

	// Add all subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(generateDocsCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(doctorCmd)

	// Legacy commands (for backward compatibility)
	rootCmd.AddCommand(addRouteCmd)
	rootCmd.AddCommand(addTemplateCmd)
	rootCmd.AddCommand(addAggregatorCmd)
	rootCmd.AddCommand(devCmd)
	rootCmd.AddCommand(postmanCmd)

	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().String("config-path", "", "Path to configuration file (default: ~/.bffgen/config.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().String("runtime", "", "Override runtime detection (go, nodejs-express, nodejs-fastify)")

	// Bind flags to viper
	viper.BindPFlag("config_path", rootCmd.PersistentFlags().Lookup("config-path"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("no_color", rootCmd.PersistentFlags().Lookup("no-color"))
	viper.BindPFlag("runtime", rootCmd.PersistentFlags().Lookup("runtime"))
}
