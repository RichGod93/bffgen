package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RichGod93/bffgen/internal/scaffolding"
	"github.com/RichGod93/bffgen/internal/types"
)

// Integration tests that test actual file operations

func TestCreateGoModFile_Integration(t *testing.T) {
	tempDir := t.TempDir()
	projectName := filepath.Join(tempDir, "test-go-project")

	// Create project directory
	err := os.MkdirAll(projectName, 0755)
	if err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	// Create go.mod file
	err = createGoModFile(projectName, "chi")
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Verify file exists and has content
	goModPath := filepath.Join(projectName, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "module") {
		t.Error("go.mod should contain module directive")
	}
	if !strings.Contains(contentStr, "github.com/go-chi/chi/v5") {
		t.Error("go.mod should contain chi dependency")
	}
}

func TestCreatePackageJsonFile_Integration(t *testing.T) {
	tempDir := t.TempDir()
	projectName := filepath.Join(tempDir, "test-node-project")

	err := os.MkdirAll(projectName, 0755)
	if err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	// Create package.json file
	err = createPackageJsonFile(projectName, scaffolding.LanguageNodeExpress, "express")
	if err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Verify file exists
	packagePath := filepath.Join(projectName, "package.json")
	content, err := os.ReadFile(packagePath)
	if err != nil {
		t.Fatalf("Failed to read package.json: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, `"express"`) {
		t.Error("package.json should contain express dependency")
	}
	if !strings.Contains(contentStr, `"name"`) {
		t.Error("package.json should have name field")
	}
}

func TestCreateProjectDirectories_Integration_Go(t *testing.T) {
	tempDir := t.TempDir()
	projectName := filepath.Join(tempDir, "test-go-project")

	err := createProjectDirectories(projectName, scaffolding.LanguageGo)
	if err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	// Check expected directories exist
	expectedDirs := []string{
		"internal/routes",
		"internal/aggregators",
		"cmd/server",
	}

	for _, dir := range expectedDirs {
		fullPath := filepath.Join(projectName, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Expected directory %s to exist", dir)
		}
	}
}

func TestCreateProjectDirectories_Integration_NodeJS(t *testing.T) {
	tempDir := t.TempDir()
	projectName := filepath.Join(tempDir, "test-node-project")

	err := createProjectDirectories(projectName, scaffolding.LanguageNodeExpress)
	if err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	// Check expected directories exist
	expectedDirs := []string{
		"src",
		"src/routes",
		"src/middleware",
		"src/controllers",
		"tests",
	}

	for _, dir := range expectedDirs {
		fullPath := filepath.Join(projectName, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Expected directory %s to exist", dir)
		}
	}
}

func TestGenerateProxyHandlerFunction_OutputFormat(t *testing.T) {
	result := generateProxyHandlerFunction()

	// Verify it contains expected Go code patterns
	expectedPatterns := []string{
		"func createProxyHandler",
		"http.HandlerFunc",
		"httputil.NewSingleHostReverseProxy",
		"proxy.ServeHTTP",
		"url.Parse",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(result, pattern) {
			t.Errorf("Expected pattern %q not found in output", pattern)
		}
	}
}

func TestGenerateServerContent_OutputFormat(t *testing.T) {
	config := &types.BFFConfig{
		Services: map[string]types.Service{
			"users": {
				BaseURL: "http://localhost:4000/api",
				Endpoints: []types.Endpoint{
					{Path: "/users", Method: "GET", ExposeAs: "/api/users"},
				},
			},
		},
		Settings: types.Settings{Port: 8080},
	}

	result := generateServerContent(config)

	// Verify server code patterns
	expectedPatterns := []string{
		"package main",
		"func main",
		"http.ListenAndServe",
		":8080",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(result, pattern) {
			t.Errorf("Expected pattern %q not found in server content", pattern)
		}
	}
}

func TestCreateDependencyFiles_Go(t *testing.T) {
	tempDir := t.TempDir()
	projectName := filepath.Join(tempDir, "test-project")

	err := os.MkdirAll(projectName, 0755)
	if err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	err = createDependencyFiles(projectName, scaffolding.LanguageGo, "chi")
	if err != nil {
		t.Fatalf("createDependencyFiles failed: %v", err)
	}

	// Verify go.mod was created
	goModPath := filepath.Join(projectName, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Error("go.mod should have been created")
	}
}

func TestCreateDependencyFiles_NodeJS(t *testing.T) {
	tempDir := t.TempDir()
	projectName := filepath.Join(tempDir, "test-project")

	err := os.MkdirAll(projectName, 0755)
	if err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	err = createDependencyFiles(projectName, scaffolding.LanguageNodeExpress, "express")
	if err != nil {
		t.Fatalf("createDependencyFiles failed: %v", err)
	}

	// Verify package.json was created
	packagePath := filepath.Join(projectName, "package.json")
	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		t.Error("package.json should have been created")
	}
}

func TestGenerateGoModContent_AllFrameworks(t *testing.T) {
	frameworks := []string{"chi", "echo", "fiber", "unknown"}

	for _, fw := range frameworks {
		t.Run(fw, func(t *testing.T) {
			result := generateGoModContent("test-project", fw)

			// All should have basic structure
			if !strings.Contains(result, "module test-project") {
				t.Error("Should contain module directive")
			}
			if !strings.Contains(result, "go 1.21") {
				t.Error("Should contain go version")
			}

			// Framework-specific checks
			switch fw {
			case "chi":
				if !strings.Contains(result, "github.com/go-chi/chi/v5") {
					t.Error("Chi should have chi dependency")
				}
			case "echo":
				if !strings.Contains(result, "github.com/labstack/echo/v4") {
					t.Error("Echo should have echo dependency")
				}
			case "fiber":
				if !strings.Contains(result, "github.com/gofiber/fiber/v2") {
					t.Error("Fiber should have fiber dependency")
				}
			}
		})
	}
}

func TestGeneratePackageJsonContent_AllFrameworks(t *testing.T) {
	frameworks := []struct {
		name     string
		langType scaffolding.LanguageType
	}{
		{"express", scaffolding.LanguageNodeExpress},
		{"fastify", scaffolding.LanguageNodeFastify},
	}

	for _, fw := range frameworks {
		t.Run(fw.name, func(t *testing.T) {
			result := generatePackageJsonContent("test-project", fw.langType, fw.name)

			// Basic structure
			if !strings.Contains(result, `"name"`) {
				t.Error("Should have name field")
			}
			if !strings.Contains(result, `"version"`) {
				t.Error("Should have version field")
			}
			if !strings.Contains(result, `"scripts"`) {
				t.Error("Should have scripts")
			}

			// Framework-specific
			switch fw.name {
			case "express":
				if !strings.Contains(result, `"express"`) {
					t.Error("Express should have express dependency")
				}
			case "fastify":
				if !strings.Contains(result, `"fastify"`) {
					t.Error("Fastify should have fastify dependency")
				}
			}
		})
	}
}

func TestDetectProjectType_Integration(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func(string) error
		expectedType string
	}{
		{
			name: "Go project with go.mod",
			setupFunc: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\ngo 1.21"), 0644)
			},
			expectedType: "go",
		},
		{
			name: "Go project with bff.config.yaml",
			setupFunc: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "bff.config.yaml"), []byte("name: test\nlanguage: go"), 0644)
			},
			expectedType: "go",
		},
		{
			name: "Node.js project with package.json",
			setupFunc: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name": "test"}`), 0644)
			},
			expectedType: "nodejs",
		},
		{
			name: "Node.js project with bffgen.config.json",
			setupFunc: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "bffgen.config.json"), []byte(`{"name": "test", "language": "nodejs"}`), 0644)
			},
			expectedType: "nodejs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			oldWd, _ := os.Getwd()
			os.Chdir(tempDir)
			defer os.Chdir(oldWd)

			if tt.setupFunc != nil {
				err := tt.setupFunc(tempDir)
				if err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			result := detectProjectType()
			if result != tt.expectedType {
				t.Errorf("Expected project type %q, got %q", tt.expectedType, result)
			}
		})
	}
}

func TestGenerateCORSConfig_AllFrameworks_Integration(t *testing.T) {
	origins := []string{"http://localhost:3000", "https://example.com"}
	frameworks := []string{"chi", "echo", "fiber"}

	for _, fw := range frameworks {
		t.Run(fw, func(t *testing.T) {
			result := generateCORSConfig(origins, fw)

			if result == "" {
				t.Error("Expected non-empty CORS config")
			}

			// All should contain the origins
			for _, origin := range origins {
				if !strings.Contains(result, origin) {
					t.Errorf("Expected origin %s in config", origin)
				}
			}

			// All should have standard HTTP methods
			methods := []string{"GET", "POST", "PUT", "DELETE"}
			for _, method := range methods {
				if !strings.Contains(result, method) {
					t.Errorf("Expected method %s in config", method)
				}
			}
		})
	}
}
