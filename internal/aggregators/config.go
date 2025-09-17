package aggregators

import (
	"encoding/json"
	"os"
	"time"
)

// AggregatorConfig represents the configuration for an aggregator
type AggregatorConfig struct {
	Name        string            `yaml:"name" json:"name"`
	Path        string            `yaml:"path" json:"path"`
	Description string            `yaml:"description" json:"description"`
	Services    []string          `yaml:"services" json:"services"`
	Cache       CacheConfig       `yaml:"cache" json:"cache"`
	Timeout     time.Duration     `yaml:"timeout" json:"timeout"`
	Retries     int               `yaml:"retries" json:"retries"`
	Headers     map[string]string `yaml:"headers" json:"headers"`
	Enabled     bool              `yaml:"enabled" json:"enabled"`
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	Enabled bool          `yaml:"enabled" json:"enabled"`
	TTL     time.Duration `yaml:"ttl" json:"ttl"`
	MaxSize int           `yaml:"maxSize" json:"maxSize"`
}

// AggregatorConfigs represents a collection of aggregator configurations
type AggregatorConfigs struct {
	Aggregators []AggregatorConfig `yaml:"aggregators" json:"aggregators"`
	Global      GlobalConfig       `yaml:"global" json:"global"`
}

// GlobalConfig represents global aggregator settings
type GlobalConfig struct {
	DefaultTimeout time.Duration `yaml:"defaultTimeout" json:"defaultTimeout"`
	DefaultRetries int           `yaml:"defaultRetries" json:"defaultRetries"`
	CacheEnabled   bool          `yaml:"cacheEnabled" json:"cacheEnabled"`
	CacheTTL       time.Duration `yaml:"cacheTTL" json:"cacheTTL"`
}

// LoadAggregatorConfigs loads aggregator configurations from a file
func LoadAggregatorConfigs(filename string) (*AggregatorConfigs, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var configs AggregatorConfigs
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, err
	}

	return &configs, nil
}

// SaveAggregatorConfigs saves aggregator configurations to a file
func SaveAggregatorConfigs(filename string, configs *AggregatorConfigs) error {
	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// DefaultAggregatorConfigs returns default aggregator configurations
func DefaultAggregatorConfigs() *AggregatorConfigs {
	return &AggregatorConfigs{
		Aggregators: []AggregatorConfig{
			{
				Name:        "user-dashboard",
				Path:        "/api/user-dashboard/:id",
				Description: "Aggregates user, orders, and preferences data",
				Services:    []string{"users", "orders", "preferences"},
				Cache: CacheConfig{
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
			},
			{
				Name:        "ecommerce-catalog",
				Path:        "/api/catalog/:category",
				Description: "Aggregates products, inventory, and cart data",
				Services:    []string{"products", "inventory", "cart"},
				Cache: CacheConfig{
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
			},
		},
		Global: GlobalConfig{
			DefaultTimeout: 30 * time.Second,
			DefaultRetries: 3,
			CacheEnabled:   true,
			CacheTTL:       5 * time.Minute,
		},
	}
}
