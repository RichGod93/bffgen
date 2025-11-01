package validation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/RichGod93/bffgen/internal/types"
)

func TestSchemaValidator_ValidateConfig(t *testing.T) {
	validator, err := NewSchemaValidator()
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Test valid configuration
	validConfig := types.GetDefaultBFFGenV1Config()
	err = validator.ValidateConfig(validConfig)
	if err != nil {
		t.Errorf("Valid config should pass validation: %v", err)
	}

	// Test invalid configuration (missing required fields)
	invalidConfig := &types.BFFGenV1Config{
		Version: "1.0",
		// Missing required project field
	}
	err = validator.ValidateConfig(invalidConfig)
	if err == nil {
		t.Error("Invalid config should fail validation")
	}
}

func TestSchemaValidator_ValidateYAMLFile(t *testing.T) {
	validator, err := NewSchemaValidator()
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Create a temporary YAML file
	tempDir := t.TempDir()
	yamlFile := filepath.Join(tempDir, "test.yaml")

	validYAML := `
version: "1.0"
project:
  name: "test-bff"
  framework: "chi"
`

	err = os.WriteFile(yamlFile, []byte(validYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to write test YAML file: %v", err)
	}

	// Test valid YAML file
	config, err := validator.ValidateYAMLFile(yamlFile)
	if err != nil {
		t.Errorf("Valid YAML file should pass validation: %v", err)
	}
	if config == nil {
		t.Error("Config should not be nil")
		return
	}
	if config.Project.Name != "test-bff" {
		t.Errorf("Expected project name 'test-bff', got '%s'", config.Project.Name)
	}

	// Test invalid YAML file
	invalidYAML := `
version: "1.0"
# Missing required project field
`

	invalidFile := filepath.Join(tempDir, "invalid.yaml")
	err = os.WriteFile(invalidFile, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid YAML file: %v", err)
	}

	_, err = validator.ValidateYAMLFile(invalidFile)
	if err == nil {
		t.Error("Invalid YAML file should fail validation")
	}
}

func TestSchemaValidator_ValidateAndSetDefaults(t *testing.T) {
	validator, err := NewSchemaValidator()
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Test with minimal config
	minimalConfig := &types.BFFGenV1Config{
		Version: "1.0",
		Project: types.ProjectConfig{
			Name:      "minimal-bff",
			Framework: "chi",
		},
	}

	mergedConfig, err := validator.ValidateAndSetDefaults(minimalConfig)
	if err != nil {
		t.Errorf("Should merge defaults successfully: %v", err)
	}

	// Check that defaults were applied
	if mergedConfig.Server == nil {
		t.Error("Server config should be set from defaults")
	}
	if mergedConfig.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", mergedConfig.Server.Port)
	}
	if mergedConfig.Auth == nil {
		t.Error("Auth config should be set from defaults")
	}
	if mergedConfig.Auth.Mode != "jwt" {
		t.Errorf("Expected default auth mode 'jwt', got '%s'", mergedConfig.Auth.Mode)
	}
}

func TestSchemaValidator_GetSchemaVersion(t *testing.T) {
	validator, err := NewSchemaValidator()
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	version := validator.GetSchemaVersion()
	if version != "1.0" {
		t.Errorf("Expected schema version '1.0', got '%s'", version)
	}
}

func TestValidateSchemaFile(t *testing.T) {
	// Test with valid schema file
	schemaPath := "schemas/bffgen-v1.json"

	// Try to find the schema file in the current directory or parent directories
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		// Try parent directory
		schemaPath = filepath.Join("..", "schemas", "bffgen-v1.json")
		if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
			// Try project root
			schemaPath = filepath.Join("..", "..", "schemas", "bffgen-v1.json")
		}
	}

	err := ValidateSchemaFile(schemaPath)
	if err != nil {
		t.Errorf("Valid schema file should pass validation: %v", err)
	}

	// Test with non-existent file
	err = ValidateSchemaFile("non-existent.json")
	if err == nil {
		t.Error("Non-existent file should fail validation")
	}

	// Test with invalid JSON
	tempDir := t.TempDir()
	invalidSchemaFile := filepath.Join(tempDir, "invalid.json")
	invalidJSON := `{ invalid json }`

	err = os.WriteFile(invalidSchemaFile, []byte(invalidJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid JSON file: %v", err)
	}

	err = ValidateSchemaFile(invalidSchemaFile)
	if err == nil {
		t.Error("Invalid JSON schema should fail validation")
	}
}

func TestSchemaValidator_ValidateJSONFile(t *testing.T) {
	validator, err := NewSchemaValidator()
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// Create a temporary JSON file
	tempDir := t.TempDir()
	jsonFile := filepath.Join(tempDir, "test.json")

	validJSON := `{
		"version": "1.0",
		"project": {
			"name": "test-bff",
			"framework": "chi"
		}
	}`

	err = os.WriteFile(jsonFile, []byte(validJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write test JSON file: %v", err)
	}

	// Test valid JSON file
	config, err := validator.ValidateJSONFile(jsonFile)
	if err != nil {
		t.Errorf("Valid JSON file should pass validation: %v", err)
	}
	if config == nil {
		t.Error("Config should not be nil")
		return
	}
	if config.Project.Name != "test-bff" {
		t.Errorf("Expected project name 'test-bff', got '%s'", config.Project.Name)
	}
}

func TestSchemaValidator_mergeConfigs(t *testing.T) {
	validator, err := NewSchemaValidator()
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	defaults := types.GetDefaultBFFGenV1Config()
	config := &types.BFFGenV1Config{
		Version: "1.0",
		Project: types.ProjectConfig{
			Name:      "custom-bff",
			Framework: "fiber", // Override default
		},
	}

	merged := validator.mergeConfigs(defaults, config)

	// Check that custom values override defaults
	if merged.Project.Name != "custom-bff" {
		t.Errorf("Expected custom name 'custom-bff', got '%s'", merged.Project.Name)
	}
	if merged.Project.Framework != "fiber" {
		t.Errorf("Expected custom framework 'fiber', got '%s'", merged.Project.Framework)
	}

	// Check that defaults are preserved where not overridden
	if merged.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", merged.Server.Port)
	}
}
