package utils

import (
	"encoding/json"
	"os"

	"github.com/RichGod93/bffgen/internal/types"
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
func SaveConfig(config *types.BFFConfig, configPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// ConfigToJSON converts BFF configuration to JSON format
func ConfigToJSON(config *types.BFFConfig) ([]byte, error) {
	return json.MarshalIndent(config, "", "  ")
}
