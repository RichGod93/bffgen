package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// GlobalConfig holds global CLI configuration
type GlobalConfig struct {
	ConfigPath      string
	Verbose         bool
	NoColor         bool
	RuntimeOverride string // Explicit runtime override (go, nodejs-express, nodejs-fastify)
}

var globalConfig GlobalConfig

// InitGlobalConfig initializes global configuration
func InitGlobalConfig() error {
	// Set default config path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	defaultConfigPath := filepath.Join(homeDir, ".bffgen", "config.yaml")
	globalConfig.ConfigPath = defaultConfigPath

	// Initialize viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(filepath.Join(homeDir, ".bffgen"))
	viper.AddConfigPath(".")

	// Set default values
	viper.SetDefault("verbose", false)
	viper.SetDefault("no_color", false)
	viper.SetDefault("config_path", defaultConfigPath)

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is OK, we'll use defaults
	}

	return nil
}

// GetGlobalConfig returns the global configuration
func GetGlobalConfig() GlobalConfig {
	return globalConfig
}

// SetGlobalConfig updates the global configuration
func SetGlobalConfig(config GlobalConfig) {
	globalConfig = config
}

// LogVerbose logs a message if verbose mode is enabled
func LogVerbose(format string, args ...interface{}) {
	if globalConfig.Verbose {
		fmt.Printf(format+"\n", args...)
	}
}

// GetConfigPath returns the configuration file path
func GetConfigPath() string {
	if globalConfig.ConfigPath != "" {
		return globalConfig.ConfigPath
	}

	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".bffgen", "config.yaml")
}

// EnsureConfigDir creates the config directory if it doesn't exist
func EnsureConfigDir() error {
	configPath := GetConfigPath()
	configDir := filepath.Dir(configPath)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return nil
}
