package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RichGod93/bffgen/v3/internal/scaffolding"
)

func TestGenerateGoModContent(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
		framework   string
		expectChi   bool
		expectEcho  bool
		expectFiber bool
	}{
		{"Chi", "test-project", "chi", true, false, false},
		{"Echo", "test-project", "echo", false, true, false},
		{"Fiber", "test-project", "fiber", false, false, true},
		{"Default", "test-project", "unknown", false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateGoModContent(tt.projectName, tt.framework)

			// Check module name
			if !strings.Contains(result, tt.projectName) {
				t.Errorf("Expected module name %q, got:\n%s", tt.projectName, result)
			}

			// Check framework-specific dependencies
			if tt.expectChi && !strings.Contains(result, "github.com/go-chi/chi/v5") {
				t.Error("Expected chi dependencies")
			}
			if tt.expectEcho && !strings.Contains(result, "github.com/labstack/echo/v4") {
				t.Error("Expected echo dependencies")
			}
			if tt.expectFiber && !strings.Contains(result, "github.com/gofiber/fiber/v2") {
				t.Error("Expected fiber dependencies")
			}

			// Check common dependencies
			if !strings.Contains(result, "go 1.25.0") {
				t.Error("Expected go 1.25.0")
			}
			if !strings.Contains(result, "gopkg.in/yaml.v3") {
				t.Error("Expected yaml.v3 dependency")
			}
		})
	}
}

// Note: TestGenerateCORSConfig already exists in init_test.go
// Adding only framework-specific tests below

func TestGenerateCORSConfig_EmptyOrigins(t *testing.T) {
	result := generateCORSConfig([]string{}, "chi")

	// Should still produce valid CORS config even with no origins
	if result == "" {
		t.Error("Expected non-empty CORS config")
	}
}

func TestGenerateCORSConfig_ChiSpecific(t *testing.T) {
	origins := []string{"http://localhost:3000"}
	result := generateCORSConfig(origins, "chi")

	// Chi-specific checks
	if !strings.Contains(result, "cors.Handler") {
		t.Error("Expected cors.Handler for Chi")
	}
	if !strings.Contains(result, "AllowedOrigins") {
		t.Error("Expected AllowedOrigins for Chi")
	}
	if !strings.Contains(result, "AllowedMethods") {
		t.Error("Expected AllowedMethods")
	}
	if !strings.Contains(result, "GET") || !strings.Contains(result, "POST") {
		t.Error("Expected HTTP methods in AllowedMethods")
	}
}

func TestGenerateCORSConfig_EchoSpecific(t *testing.T) {
	origins := []string{"http://localhost:3000", "https://example.com"}
	result := generateCORSConfig(origins, "echo")

	// Echo-specific checks
	if !strings.Contains(result, "AllowOrigins") {
		t.Error("Expected AllowOrigins for Echo")
	}
	if !strings.Contains(result, "AllowMethods") {
		t.Error("Expected AllowMethods for Echo")
	}
	if !strings.Contains(result, "http://localhost:3000") {
		t.Error("Expected localhost origin")
	}
	if !strings.Contains(result, "https://example.com") {
		t.Error("Expected example.com origin")
	}
}

func TestGenerateCORSConfig_FiberSpecific(t *testing.T) {
	origins := []string{"http://localhost:3000"}
	result := generateCORSConfig(origins, "fiber")

	// Fiber-specific checks
	if !strings.Contains(result, "AllowOrigins") {
		t.Error("Expected AllowOrigins for Fiber")
	}
	if !strings.Contains(result, "AllowMethods") {
		t.Error("Expected AllowMethods for Fiber")
	}
	if !strings.Contains(result, "GET") || !strings.Contains(result, "POST") {
		t.Error("Expected HTTP methods")
	}
}

func TestGenerateCORSConfigWithLang(t *testing.T) {
	origins := []string{"http://localhost:3000"}

	// Test with different language types
	tests := []struct {
		name      string
		langType  scaffolding.LanguageType
		framework string
		check     string
	}{
		{"Go Chi", scaffolding.LanguageGo, "chi", "cors.Handler"},
		{"Go Echo", scaffolding.LanguageGo, "echo", "AllowOrigins"},
		// Node.js frameworks use generateCORSConfig, not generateCORSConfigWithLang in the actual code
		// So we only test Go frameworks here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateCORSConfigWithLang(origins, tt.framework, tt.langType)
			if result == "" {
				t.Errorf("Expected non-empty CORS config for %s", tt.name)
			}
			if !strings.Contains(result, tt.check) {
				t.Errorf("Expected %s pattern in result for %s\nGot:\n%s", tt.check, tt.name, result)
			}
		})
	}
}

func TestCreateProjectDirectories_Go(t *testing.T) {
	// This test checks directory creation for Go projects
	// We'll verify the expected directories are included
	// Note: This requires actual file system, so we test the logic

	tempDir := t.TempDir()
	projectName := filepath.Join(tempDir, "test-go-project")

	err := createProjectDirectories(projectName, scaffolding.LanguageGo)
	if err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	// Check that Go-specific directories were created
	expectedDirs := []string{
		"internal/routes",
		"internal/aggregators",
		"internal/templates",
		"cmd/server",
	}

	for _, dir := range expectedDirs {
		fullPath := filepath.Join(projectName, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Expected directory %s to exist", dir)
		}
	}
}

func TestCreateProjectDirectories_NodeJS(t *testing.T) {
	tempDir := t.TempDir()
	projectName := filepath.Join(tempDir, "test-nodejs-project")

	err := createProjectDirectories(projectName, scaffolding.LanguageNodeExpress)
	if err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	// Check that Node.js-specific directories were created
	expectedDirs := []string{
		"src",
		"src/routes",
		"src/middleware",
		"src/controllers",
		"src/services",
		"tests",
		"tests/unit",
	}

	for _, dir := range expectedDirs {
		fullPath := filepath.Join(projectName, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Expected directory %s to exist", dir)
		}
	}
}
