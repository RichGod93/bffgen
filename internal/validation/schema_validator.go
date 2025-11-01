package validation

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/RichGod93/bffgen/internal/types"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

// SchemaValidator validates bffgen configuration against JSON Schema
type SchemaValidator struct {
	schema *gojsonschema.Schema
}

// NewSchemaValidator creates a new schema validator
func NewSchemaValidator() (*SchemaValidator, error) {
	// Load the JSON schema
	schemaPath := filepath.Join("schemas", "bffgen-v1.json")

	// Try to find the schema file in the current directory or parent directories
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		// Try parent directory
		schemaPath = filepath.Join("..", "schemas", "bffgen-v1.json")
		if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
			// Try project root
			schemaPath = filepath.Join("..", "..", "schemas", "bffgen-v1.json")
		}
	}

	schemaData, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	// Parse the schema
	schemaLoader := gojsonschema.NewBytesLoader(schemaData)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schema: %w", err)
	}

	return &SchemaValidator{
		schema: schema,
	}, nil
}

// ValidateConfig validates a bffgen configuration against the schema
func (sv *SchemaValidator) ValidateConfig(config *types.BFFGenV1Config) error {
	// Convert config to JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	// Create document loader
	documentLoader := gojsonschema.NewBytesLoader(configJSON)

	// Validate
	result, err := sv.schema.Validate(documentLoader)
	if err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	if !result.Valid() {
		var errors []string
		for _, desc := range result.Errors() {
			errors = append(errors, fmt.Sprintf("%s: %s", desc.Field(), desc.Description()))
		}
		return fmt.Errorf("validation errors: %v", errors)
	}

	return nil
}

// ValidateYAMLFile validates a YAML configuration file against the schema
func (sv *SchemaValidator) ValidateYAMLFile(filePath string) (*types.BFFGenV1Config, error) {
	// Read YAML file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse YAML
	var config types.BFFGenV1Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate against schema
	if err := sv.ValidateConfig(&config); err != nil {
		return nil, fmt.Errorf("schema validation failed: %w", err)
	}

	return &config, nil
}

// ValidateJSONFile validates a JSON configuration file against the schema
func (sv *SchemaValidator) ValidateJSONFile(filePath string) (*types.BFFGenV1Config, error) {
	// Read JSON file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse JSON
	var config types.BFFGenV1Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Validate against schema
	if err := sv.ValidateConfig(&config); err != nil {
		return nil, fmt.Errorf("schema validation failed: %w", err)
	}

	return &config, nil
}

// ValidateFromURL validates a configuration from a URL against the schema
func (sv *SchemaValidator) ValidateFromURL(url string) (*types.BFFGenV1Config, error) {
	// Fetch configuration from URL
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch config from URL: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// Read response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Try to parse as YAML first, then JSON
	var config types.BFFGenV1Config

	// Try YAML
	if err := yaml.Unmarshal(data, &config); err != nil {
		// Try JSON
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse config as YAML or JSON: %w", err)
		}
	}

	// Validate against schema
	if err := sv.ValidateConfig(&config); err != nil {
		return nil, fmt.Errorf("schema validation failed: %w", err)
	}

	return &config, nil
}

// ValidateAndSetDefaults validates a configuration and sets default values
func (sv *SchemaValidator) ValidateAndSetDefaults(config *types.BFFGenV1Config) (*types.BFFGenV1Config, error) {
	// Get default configuration
	defaults := types.GetDefaultBFFGenV1Config()

	// Merge with defaults (config takes precedence)
	merged := sv.mergeConfigs(defaults, config)

	// Validate the merged configuration
	if err := sv.ValidateConfig(merged); err != nil {
		return nil, fmt.Errorf("validation failed after merging defaults: %w", err)
	}

	return merged, nil
}

// mergeConfigs merges two configurations, with the second taking precedence
func (sv *SchemaValidator) mergeConfigs(defaults, config *types.BFFGenV1Config) *types.BFFGenV1Config {
	// Convert to JSON for deep merging
	defaultsJSON, _ := json.Marshal(defaults)
	configJSON, _ := json.Marshal(config)

	// Parse back to get merged result
	var merged types.BFFGenV1Config
	_ = json.Unmarshal(defaultsJSON, &merged)
	_ = json.Unmarshal(configJSON, &merged)

	return &merged
}

// GetSchemaVersion returns the schema version
func (sv *SchemaValidator) GetSchemaVersion() string {
	return "1.0"
}

// ValidateSchemaFile validates that the schema file itself is valid
func ValidateSchemaFile(schemaPath string) error {
	// Read schema file
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	// Parse as JSON to validate syntax
	var schema map[string]interface{}
	if err := json.Unmarshal(data, &schema); err != nil {
		return fmt.Errorf("invalid JSON schema syntax: %w", err)
	}

	// Validate against JSON Schema meta-schema
	metaSchemaLoader := gojsonschema.NewStringLoader(`{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"$schema": {"type": "string"},
			"$id": {"type": "string"},
			"title": {"type": "string"},
			"description": {"type": "string"},
			"type": {"type": "string"},
			"properties": {"type": "object"},
			"required": {"type": "array"}
		},
		"required": ["$schema", "type"]
	}`)

	schemaLoader := gojsonschema.NewBytesLoader(data)
	metaSchema, err := gojsonschema.NewSchema(metaSchemaLoader)
	if err != nil {
		return fmt.Errorf("failed to create meta-schema: %w", err)
	}

	result, err := metaSchema.Validate(schemaLoader)
	if err != nil {
		return fmt.Errorf("meta-schema validation failed: %w", err)
	}

	if !result.Valid() {
		var errors []string
		for _, desc := range result.Errors() {
			errors = append(errors, fmt.Sprintf("%s: %s", desc.Field(), desc.Description()))
		}
		return fmt.Errorf("schema validation errors: %v", errors)
	}

	return nil
}
