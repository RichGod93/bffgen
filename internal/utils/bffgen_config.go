package utils

import (
	"os"
	"path/filepath"

	"github.com/RichGod93/bffgen/internal/types"
	"gopkg.in/yaml.v3"
)

const (
	ConfigDir  = ".bffgen"
	ConfigFile = "bffgen.yaml"
)

// GetConfigPath returns the path to the bffgen configuration file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	
	configDir := filepath.Join(homeDir, ConfigDir)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}
	
	return filepath.Join(configDir, ConfigFile), nil
}

// LoadBFFGenConfig loads the bffgen configuration from file
func LoadBFFGenConfig() (*types.BFFGenConfig, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}
	
	// If config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return types.GetDefaultConfig(), nil
	}
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	
	var config types.BFFGenConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	
	return &config, nil
}

// SaveBFFGenConfig saves the bffgen configuration to file
func SaveBFFGenConfig(config *types.BFFGenConfig) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}
	
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	
	return os.WriteFile(configPath, data, 0644)
}

// UpdateRecentProject adds a project to the recent projects list
func UpdateRecentProject(projectName string) error {
	config, err := LoadBFFGenConfig()
	if err != nil {
		return err
	}
	
	// Remove if already exists
	for i, project := range config.History.RecentProjects {
		if project == projectName {
			config.History.RecentProjects = append(
				config.History.RecentProjects[:i],
				config.History.RecentProjects[i+1:]...,
			)
			break
		}
	}
	
	// Add to beginning
	config.History.RecentProjects = append([]string{projectName}, config.History.RecentProjects...)
	
	// Keep only last 10 projects
	if len(config.History.RecentProjects) > 10 {
		config.History.RecentProjects = config.History.RecentProjects[:10]
	}
	
	config.History.LastUsed = projectName
	
	return SaveBFFGenConfig(config)
}
