package utils

import (
	"os"

	"github.com/richgodusen/bffgen/internal/types"
	"gopkg.in/yaml.v3"
)

// LoadConfig loads BFF configuration from a YAML file
func LoadConfig(configPath string) (*types.BFFConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config types.BFFConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfig saves BFF configuration to a YAML file
func SaveConfig(configPath string, config *types.BFFConfig) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
