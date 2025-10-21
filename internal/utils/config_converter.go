package utils

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/RichGod93/bffgen/internal/types"
	"gopkg.in/yaml.v3"
)

// ConvertYAMLToJSON converts bff.config.yaml to bffgen.config.json
func ConvertYAMLToJSON(outputPath string) error {
	// Read YAML file
	yamlPath := "bff.config.yaml"
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		return fmt.Errorf("bff.config.yaml not found")
	}

	config, err := LoadConfig(yamlPath)
	if err != nil {
		return fmt.Errorf("failed to load YAML config: %w", err)
	}

	// Convert to Node.js config format
	nodeConfig := convertGoConfigToNodeJS(config)

	// Determine output path
	if outputPath == "" {
		outputPath = "bffgen.config.json"
	}

	// Write JSON file
	data, err := json.MarshalIndent(nodeConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	fmt.Printf("📄 Converted bff.config.yaml → %s\n", outputPath)
	return nil
}

// ConvertJSONToYAML converts bffgen.config.json to bff.config.yaml
func ConvertJSONToYAML(outputPath string) error {
	// Read JSON file
	jsonPath := "bffgen.config.json"
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		return fmt.Errorf("bffgen.config.json not found")
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to read JSON config: %w", err)
	}

	var nodeConfig map[string]interface{}
	if err := json.Unmarshal(data, &nodeConfig); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Convert to Go config format
	goConfig := convertNodeJSConfigToGo(nodeConfig)

	// Determine output path
	if outputPath == "" {
		outputPath = "bff.config.yaml"
	}

	// Write YAML file
	yamlData, err := yaml.Marshal(goConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	if err := os.WriteFile(outputPath, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write YAML file: %w", err)
	}

	fmt.Printf("📄 Converted bffgen.config.json → %s\n", outputPath)
	return nil
}

// convertGoConfigToNodeJS converts Go config structure to Node.js format
func convertGoConfigToNodeJS(config *types.BFFConfig) map[string]interface{} {
	result := map[string]interface{}{
		"project": map[string]interface{}{
			"name":      "converted-project",
			"version":   "1.0.0",
			"framework": "express", // Default to express
		},
		"server": map[string]interface{}{
			"port":     config.Settings.Port,
			"host":     "0.0.0.0",
			"nodeEnv":  "development",
			"logLevel": "info",
		},
		"cors": map[string]interface{}{
			"origins":        []string{"http://localhost:3000"},
			"credentials":    true,
			"methods":        []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			"allowedHeaders": []string{"Content-Type", "Authorization"},
			"maxAge":         3600,
		},
		"rateLimiting": map[string]interface{}{
			"enabled":         true,
			"windowMs":        900000,
			"maxRequests":     100,
			"standardHeaders": true,
			"legacyHeaders":   false,
		},
		"backends": []interface{}{},
	}

	// Convert services to backends
	backends := []interface{}{}
	for serviceName, service := range config.Services {
		backend := map[string]interface{}{
			"name":    serviceName,
			"baseUrl": service.BaseURL,
			"timeout": 30000,
			"retries": 3,
			"healthCheck": map[string]interface{}{
				"enabled":  true,
				"path":     "/health",
				"interval": 60000,
			},
			"endpoints": []interface{}{},
		}

		// Convert endpoints
		endpoints := []interface{}{}
		for _, endpoint := range service.Endpoints {
			ep := map[string]interface{}{
				"name":         endpoint.Name,
				"path":         endpoint.Path,
				"method":       endpoint.Method,
				"exposeAs":     endpoint.ExposeAs,
				"requiresAuth": false,
				"timeout":      30000,
			}
			endpoints = append(endpoints, ep)
		}

		backend["endpoints"] = endpoints
		backends = append(backends, backend)
	}

	result["backends"] = backends

	return result
}

// convertNodeJSConfigToGo converts Node.js config structure to Go format
func convertNodeJSConfigToGo(nodeConfig map[string]interface{}) *types.BFFConfig {
	config := &types.BFFConfig{
		Services: make(map[string]types.Service),
		Settings: types.Settings{
			Port:    8080,
			Timeout: "30s",
			Retries: 3,
		},
	}

	// Get server settings
	if server, ok := nodeConfig["server"].(map[string]interface{}); ok {
		if port, ok := server["port"].(float64); ok {
			config.Settings.Port = int(port)
		}
	}

	// Convert backends to services
	if backends, ok := nodeConfig["backends"].([]interface{}); ok {
		for _, b := range backends {
			backend, ok := b.(map[string]interface{})
			if !ok {
				continue
			}

			serviceName, _ := backend["name"].(string)
			baseURL, _ := backend["baseUrl"].(string)

			if serviceName == "" || baseURL == "" {
				continue
			}

			service := types.Service{
				BaseURL:   baseURL,
				Endpoints: []types.Endpoint{},
			}

			// Convert endpoints
			if endpoints, ok := backend["endpoints"].([]interface{}); ok {
				for _, e := range endpoints {
					endpoint, ok := e.(map[string]interface{})
					if !ok {
						continue
					}

					ep := types.Endpoint{
						Name:     getStringOrDefault(endpoint, "name", ""),
						Path:     getStringOrDefault(endpoint, "path", ""),
						Method:   getStringOrDefault(endpoint, "method", "GET"),
						ExposeAs: getStringOrDefault(endpoint, "exposeAs", ""),
					}

					service.Endpoints = append(service.Endpoints, ep)
				}
			}

			config.Services[serviceName] = service
		}
	}

	return config
}

// getStringOrDefault safely gets a string value with a default
func getStringOrDefault(m map[string]interface{}, key, defaultValue string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return defaultValue
}
