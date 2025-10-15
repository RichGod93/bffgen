package utils

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/RichGod93/bffgen/internal/types"
	"gopkg.in/yaml.v3"
)

func TestConvertYAMLToJSON(t *testing.T) {
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldDir)

	// Create a sample YAML config
	yamlConfig := `services:
  users:
    baseUrl: http://localhost:4000/api
    endpoints:
      - name: getUsers
        path: /users
        method: GET
        exposeAs: /api/users
settings:
  port: 8080
  timeout: 30s
  retries: 3
`

	os.WriteFile("bff.config.yaml", []byte(yamlConfig), 0644)

	// Convert
	err := ConvertYAMLToJSON("")
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	// Verify JSON file exists
	if _, err := os.Stat("bffgen.config.json"); os.IsNotExist(err) {
		t.Fatal("bffgen.config.json should have been created")
	}

	// Read and validate JSON
	data, _ := os.ReadFile("bffgen.config.json")
	var nodeConfig map[string]interface{}
	if err := json.Unmarshal(data, &nodeConfig); err != nil {
		t.Fatalf("Failed to parse generated JSON: %v", err)
	}

	// Check structure
	if _, ok := nodeConfig["backends"]; !ok {
		t.Error("JSON should have backends field")
	}

	if _, ok := nodeConfig["project"]; !ok {
		t.Error("JSON should have project field")
	}
}

func TestConvertJSONToYAML(t *testing.T) {
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldDir)

	// Create a sample JSON config
	jsonConfig := `{
  "project": {
    "name": "test-project",
    "framework": "express"
  },
  "backends": [
    {
      "name": "users",
      "baseUrl": "http://localhost:4000/api",
      "endpoints": [
        {
          "name": "getUsers",
          "path": "/users",
          "method": "GET",
          "exposeAs": "/api/users"
        }
      ]
    }
  ]
}`

	os.WriteFile("bffgen.config.json", []byte(jsonConfig), 0644)

	// Convert
	err := ConvertJSONToYAML("")
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	// Verify YAML file exists
	if _, err := os.Stat("bff.config.yaml"); os.IsNotExist(err) {
		t.Fatal("bff.config.yaml should have been created")
	}

	// Read and validate YAML
	data, _ := os.ReadFile("bff.config.yaml")
	var goConfig types.BFFConfig
	if err := yaml.Unmarshal(data, &goConfig); err != nil {
		t.Fatalf("Failed to parse generated YAML: %v", err)
	}

	// Check structure
	if len(goConfig.Services) == 0 {
		t.Error("YAML should have services")
	}

	if goConfig.Settings.Port == 0 {
		t.Error("YAML should have port setting")
	}
}

func TestConvertGoConfigToNodeJS(t *testing.T) {
	config := &types.BFFConfig{
		Services: map[string]types.Service{
			"users": {
				BaseURL: "http://localhost:4000/api",
				Endpoints: []types.Endpoint{
					{
						Name:     "getUsers",
						Path:     "/users",
						Method:   "GET",
						ExposeAs: "/api/users",
					},
				},
			},
		},
		Settings: types.Settings{
			Port:    8080,
			Timeout: "30s",
		},
	}

	result := convertGoConfigToNodeJS(config)

	if result["project"] == nil {
		t.Error("Result should have project field")
	}

	if result["backends"] == nil {
		t.Error("Result should have backends field")
	}

	backends := result["backends"].([]interface{})
	if len(backends) != 1 {
		t.Errorf("Expected 1 backend, got %d", len(backends))
	}
}

func TestConvertNodeJSConfigToGo(t *testing.T) {
	nodeConfig := map[string]interface{}{
		"project": map[string]interface{}{
			"name": "test",
		},
		"server": map[string]interface{}{
			"port": float64(8080),
		},
		"backends": []interface{}{
			map[string]interface{}{
				"name":    "users",
				"baseUrl": "http://localhost:4000/api",
				"endpoints": []interface{}{
					map[string]interface{}{
						"name":     "getUsers",
						"path":     "/users",
						"method":   "GET",
						"exposeAs": "/api/users",
					},
				},
			},
		},
	}

	result := convertNodeJSConfigToGo(nodeConfig)

	if len(result.Services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(result.Services))
	}

	if result.Settings.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", result.Settings.Port)
	}

	if _, ok := result.Services["users"]; !ok {
		t.Error("Should have users service")
	}
}

func TestGetStringOrDefault(t *testing.T) {
	m := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}

	t.Run("ExistingKey", func(t *testing.T) {
		result := getStringOrDefault(m, "key1", "default")
		if result != "value1" {
			t.Errorf("Expected 'value1', got '%s'", result)
		}
	})

	t.Run("NonExistingKey", func(t *testing.T) {
		result := getStringOrDefault(m, "nonexistent", "default")
		if result != "default" {
			t.Errorf("Expected 'default', got '%s'", result)
		}
	})

	t.Run("NonStringValue", func(t *testing.T) {
		result := getStringOrDefault(m, "key2", "default")
		if result != "default" {
			t.Errorf("Expected 'default' for non-string value, got '%s'", result)
		}
	})
}
