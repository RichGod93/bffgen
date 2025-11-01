package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSanitizeFunctionName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"get_users", "get_users"},
		{"get-users", "get_users"},
		{"get users", "get_users"},
		{"getUsers", "get_users"},
		{"GetUsers", "get_users"},
		{"get/users", "get_users"},
		{"get.users", "get_users"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// sanitizeFunctionName requires name, path, method - passing name with empty path/method
			result := sanitizeFunctionName(tt.input, "", "")
			if result != tt.expected {
				t.Errorf("sanitizeFunctionName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user_service", "UserService"},
		{"product_api", "ProductApi"},
		{"simple", "Simple"},
		{"", ""},
		{"user", "User"},
		{"my_long_service_name", "MyLongServiceName"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toCamelCase(tt.input)
			if result != tt.expected {
				t.Errorf("toCamelCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGeneratePythonWithConfig(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := filepath.Join(tmpDir, "test-python-project")

	// Create project structure
	err := os.MkdirAll(projectName, 0755)
	if err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	err = createPythonDirectories(projectName)
	if err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	// Create main.py
	opts := ProjectOptions{AsyncEndpoints: true}
	err = createFastAPIMainFile(projectName, opts)
	if err != nil {
		t.Fatalf("Failed to create main.py: %v", err)
	}

	// Create config file
	config := map[string]interface{}{
		"project": map[string]interface{}{
			"name":      "test-project",
			"framework": "fastapi",
			"async":     true,
		},
		"backends": []map[string]interface{}{
			{
				"name":    "api",
				"baseUrl": "http://localhost:5000",
				"endpoints": []map[string]interface{}{
					{
						"name":         "health",
						"method":       "GET",
						"path":         "/api/health",
						"upstreamPath": "/health",
					},
				},
			},
		},
	}

	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	configPath := filepath.Join(projectName, "bffgen.config.py.json")
	err = os.WriteFile(configPath, configData, 0644)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Change to project directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(projectName)

	// Generate code
	err = generatePython()
	if err != nil {
		t.Fatalf("generatePython() failed: %v", err)
	}

	// Verify router was created
	routerPath := filepath.Join(projectName, "routers", "api_router.py")
	if _, err := os.Stat(routerPath); os.IsNotExist(err) {
		t.Error("Router was not generated")
	}

	// Verify service was created
	servicePath := filepath.Join(projectName, "services", "api_service.py")
	if _, err := os.Stat(servicePath); os.IsNotExist(err) {
		t.Error("Service was not generated")
	}

	// Verify main.py was updated
	mainPath := filepath.Join(projectName, "main.py")
	content, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatalf("Failed to read main.py: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "from routers import api_router") {
		t.Error("main.py should import api_router")
	}
	if !contains(contentStr, "app.include_router(api_router.router)") {
		t.Error("main.py should register api_router")
	}
}

// Note: contains helper function is defined in init_python_test.go and shared across test files
