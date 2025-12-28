package templates_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/RichGod93/bffgen/internal/templates"
)

func TestManagerList(t *testing.T) {
	// Create temp directory for templates
	tempDir := t.TempDir()

	manager := templates.NewManager(tempDir)

	// Should ensure directory creation works
	if err := manager.EnsureTemplatesDir(); err != nil {
		t.Fatalf("Failed to ensure templates dir: %v", err)
	}

	// List should not error even with empty directory
	list, err := manager.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	// Should return empty list or bundled templates (both valid)
	if list == nil {
		t.Fatal("List() should return empty slice, not nil")
	}
}

func TestManagerGet(t *testing.T) {
	tempDir := t.TempDir()

	manager := templates.NewManager(tempDir)

	// Getting non-existent template should error
	_, err := manager.Get("nonexistent-template")
	if err == nil {
		t.Fatal("Expected error for non-existent template, got nil")
	}
}

func TestManagerExists(t *testing.T) {
	tempDir := t.TempDir()
	manager := templates.NewManager(tempDir)

	// Non-existent template should return false
	if manager.Exists("nonexistent-template") {
		t.Fatal("Exists() returned true for non-existent template")
	}
}

func TestLoadTemplate(t *testing.T) {
	// Create a test template
	tempDir := t.TempDir()
	templateDir := filepath.Join(tempDir, "test-template")

	if err := os.MkdirAll(filepath.Join(templateDir, "src"), 0755); err != nil {
		t.Fatalf("Failed to create test template directory: %v", err)
	}

	// Create template.yaml
	templateYAML := `name: test-template
version: 1.0.0
description: Test template
language: nodejs-express
features:
  - Feature 1
  - Feature 2
`
	yamlPath := filepath.Join(templateDir, "template.yaml")
	if err := os.WriteFile(yamlPath, []byte(templateYAML), 0644); err != nil {
		t.Fatalf("Failed to write template.yaml: %v", err)
	}

	// Load template
	tmpl, err := templates.LoadTemplate(templateDir)
	if err != nil {
		t.Fatalf("LoadTemplate() failed: %v", err)
	}

	// Verify template fields
	if tmpl.Name != "test-template" {
		t.Errorf("Expected name 'test-template', got '%s'", tmpl.Name)
	}
	if tmpl.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", tmpl.Version)
	}
	if tmpl.Language != "nodejs-express" {
		t.Errorf("Expected language 'nodejs-express', got '%s'", tmpl.Language)
	}
	if len(tmpl.Features) != 2 {
		t.Errorf("Expected 2 features, got %d", len(tmpl.Features))
	}
}

func TestTemplateValidation(t *testing.T) {
	tests := []struct {
		name        string
		template    templates.Template
		shouldError bool
	}{
		{
			name: "valid template",
			template: templates.Template{
				Name:     "test",
				Version:  "1.0.0",
				Language: "nodejs-express",
			},
			shouldError: false,
		},
		{
			name: "missing name",
			template: templates.Template{
				Version:  "1.0.0",
				Language: "nodejs-express",
			},
			shouldError: true,
		},
		{
			name: "missing version",
			template: templates.Template{
				Name:     "test",
				Language: "nodejs-express",
			},
			shouldError: true,
		},
		{
			name: "missing language",
			template: templates.Template{
				Name:    "test",
				Version: "1.0.0",
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.template.Validate()
			if tt.shouldError && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestGetRequiredVariables(t *testing.T) {
	tmpl := &templates.Template{
		Variables: []templates.TemplateVariable{
			{Name: "VAR1", Required: true},
			{Name: "VAR2", Required: false},
			{Name: "VAR3", Required: true},
		},
	}

	required := tmpl.GetRequiredVariables()
	if len(required) != 2 {
		t.Errorf("Expected 2 required variables, got %d", len(required))
	}
}

func TestGetVariableValue(t *testing.T) {
	tmpl := &templates.Template{
		Variables: []templates.TemplateVariable{
			{Name: "VAR1", Default: "default1"},
			{Name: "VAR2", Default: "default2"},
		},
	}

	provided := map[string]string{
		"VAR1": "custom1",
	}

	// Should return custom value when provided
	val := tmpl.GetVariableValue("VAR1", provided)
	if val != "custom1" {
		t.Errorf("Expected 'custom1', got '%s'", val)
	}

	// Should return default when not provided
	val = tmpl.GetVariableValue("VAR2", provided)
	if val != "default2" {
		t.Errorf("Expected 'default2', got '%s'", val)
	}

	// Should return empty string for unknown variable
	val = tmpl.GetVariableValue("VAR3", provided)
	if val != "" {
		t.Errorf("Expected empty string, got '%s'", val)
	}
}

func TestInstallerNormalizeGitHubURL(t *testing.T) {
	// Skipping this test as it would require making private methods public
	// The functionality is tested through integration tests
	t.Skip("URL normalization tested via InstallFromGitHub integration tests")
}

func TestRegistryLoad(t *testing.T) {
	tempDir := t.TempDir()

	// Loading from empty directory should work
	registry, err := templates.LoadRegistry(tempDir)
	if err != nil {
		t.Fatalf("LoadRegistry() failed: %v", err)
	}

	if registry == nil {
		t.Fatal("LoadRegistry() returned nil")
	}

	if len(registry.Templates) != 0 {
		t.Errorf("Expected empty templates, got %d", len(registry.Templates))
	}
}

func TestRegistryFind(t *testing.T) {
	registry := &templates.Registry{
		Templates: []templates.RegistryEntry{
			{Name: "template1", Version: "1.0.0"},
			{Name: "template2", Version: "2.0.0"},
		},
	}

	// Should find existing template
	entry := registry.Find("template1")
	if entry == nil {
		t.Fatal("Find() returned nil for existing template")
	}
	if entry.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", entry.Version)
	}

	// Should return nil for non-existent template
	entry = registry.Find("nonexistent")
	if entry != nil {
		t.Error("Find() should return nil for non-existent template")
	}
}
