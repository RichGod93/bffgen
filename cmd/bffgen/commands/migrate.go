package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/RichGod93/bffgen/internal/types"
	"github.com/RichGod93/bffgen/internal/utils"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate [source] [target]",
	Short: "Migrate BFF configuration between versions",
	Long: `Migrate BFF configuration files between different versions or formats.
This command helps upgrade your bff.config.yaml files when the format changes.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		source := args[0]
		target := args[1]
		
		if err := migrateConfig(source, target); err != nil {
			fmt.Fprintf(os.Stderr, "Error migrating config: %v\n", err)
			os.Exit(1)
		}
	},
}

func migrateConfig(source, target string) error {
	LogVerbose("Starting migration from %s to %s", source, target)
	
	// Check if source file exists
	if _, err := os.Stat(source); os.IsNotExist(err) {
		return fmt.Errorf("source file %s does not exist", source)
	}
	
	// Load source configuration
	config, err := utils.LoadConfig(source)
	if err != nil {
		return fmt.Errorf("failed to load source config: %w", err)
	}
	
	LogVerbose("Loaded configuration with %d services", len(config.Services))
	
	// Perform migration based on target format
	switch strings.ToLower(filepath.Ext(target)) {
	case ".yaml", ".yml":
		if err := migrateToYAML(config, target); err != nil {
			return fmt.Errorf("failed to migrate to YAML: %w", err)
		}
	case ".json":
		if err := migrateToJSON(config, target); err != nil {
			return fmt.Errorf("failed to migrate to JSON: %w", err)
		}
	default:
		return fmt.Errorf("unsupported target format: %s", filepath.Ext(target))
	}
	
	fmt.Printf("✅ Successfully migrated configuration from %s to %s\n", source, target)
	return nil
}

func migrateToYAML(config *types.BFFConfig, target string) error {
	// Ensure target directory exists
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}
	
	// Write YAML configuration
	if err := utils.SaveConfig(config, target); err != nil {
		return fmt.Errorf("failed to save YAML config: %w", err)
	}
	
	LogVerbose("Saved YAML configuration to %s", target)
	return nil
}

func migrateToJSON(config *types.BFFConfig, target string) error {
	// Ensure target directory exists
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}
	
	// Convert to JSON format
	jsonData, err := utils.ConfigToJSON(config)
	if err != nil {
		return fmt.Errorf("failed to convert to JSON: %w", err)
	}
	
	// Write JSON file
	if err := os.WriteFile(target, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}
	
	LogVerbose("Saved JSON configuration to %s", target)
	return nil
}
