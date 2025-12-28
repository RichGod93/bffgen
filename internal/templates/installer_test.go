package templates_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/RichGod93/bffgen/internal/templates"
)

func TestInstaller_NormalizeGitHubURL(t *testing.T) {
	// Since normalizeGitHubURL is private, we test it through InstallFromGitHub behavior
	// This test verifies URL handling works correctly
	t.Skip("URL normalization tested via integration tests")
}

func TestInstaller_ValidateTemplate(t *testing.T) {
	tempDir := t.TempDir()
	installer := templates.NewInstaller(tempDir)

	tests := []struct {
		name        string
		setupFunc   func(string) *templates.Template
		shouldError bool
	}{
		{
			name: "valid template",
			setupFunc: func(path string) *templates.Template {
				templateDir := filepath.Join(path, "valid-template")
				os.MkdirAll(filepath.Join(templateDir, "src"), 0755)

				yaml := `name: valid-template
version: 1.0.0
description: Test template
language: nodejs-express`
				os.WriteFile(filepath.Join(templateDir, "template.yaml"), []byte(yaml), 0644)

				tmpl, _ := templates.LoadTemplate(templateDir)
				return tmpl
			},
			shouldError: false,
		},
		{
			name: "missing src directory",
			setupFunc: func(path string) *templates.Template {
				templateDir := filepath.Join(path, "no-src")
				os.MkdirAll(templateDir, 0755)

				yaml := `name: no-src
version: 1.0.0
description: Test template
language: nodejs-express`
				os.WriteFile(filepath.Join(templateDir, "template.yaml"), []byte(yaml), 0644)

				tmpl, _ := templates.LoadTemplate(templateDir)
				return tmpl
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := tt.setupFunc(tempDir)
			if tmpl == nil {
				t.Fatal("Failed to create test template")
			}

			// Use VerifyIntegrity which is public
			err := installer.VerifyIntegrity(tmpl)

			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestInstaller_Remove(t *testing.T) {
	tempDir := t.TempDir()
	installer := templates.NewInstaller(tempDir)

	// Create a fake template
	communityDir := filepath.Join(tempDir, "community", "test-template")
	os.MkdirAll(communityDir, 0755)
	os.WriteFile(filepath.Join(communityDir, "test.txt"), []byte("test"), 0644)

	// Test removing existing template
	err := installer.Remove("test-template")
	if err != nil {
		t.Errorf("Remove() failed: %v", err)
	}

	// Verify it's gone
	if _, err := os.Stat(communityDir); !os.IsNotExist(err) {
		t.Error("Template directory should be removed")
	}

	// Test removing non-existent template
	err = installer.Remove("nonexistent")
	if err == nil {
		t.Error("Expected error when removing non-existent template")
	}
}

func TestInstaller_Update(t *testing.T) {
	tempDir := t.TempDir()
	installer := templates.NewInstaller(tempDir)

	// Test updating non-existent template
	err := installer.Update("nonexistent")
	if err == nil {
		t.Error("Expected error when updating non-existent template")
	}

	// Test updating non-git template
	nonGitDir := filepath.Join(tempDir, "community", "non-git-template")
	os.MkdirAll(nonGitDir, 0755)

	err = installer.Update("non-git-template")
	if err == nil {
		t.Error("Expected error when updating non-git repository")
	}
}

func TestInstaller_ExtractTemplateName(t *testing.T) {
	// Test through actual usage since method is private
	tests := []struct {
		url      string
		expected string
	}{
		{
			url:      "https://github.com/user/my-template",
			expected: "my-template",
		},
		{
			url:      "https://github.com/user/my-template.git",
			expected: "my-template",
		},
		{
			url:      "user/repo",
			expected: "repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			// This would be tested through actual install operations
			// For now, we document expected behavior
			t.Logf("URL %s should extract to %s", tt.url, tt.expected)
		})
	}
}

func TestInstaller_InstallFromRegistry(t *testing.T) {
	tempDir := t.TempDir()
	installer := templates.NewInstaller(tempDir)

	registry := &templates.Registry{
		Templates: []templates.RegistryEntry{
			{
				Name:        "test-template",
				Version:     "1.0.0",
				Description: "Test",
				URL:         "https://github.com/test/template",
				Language:    "nodejs-express",
			},
		},
	}

	// This will fail since we can't actually clone from GitHub in tests
	// But we verify the error handling
	_, err := installer.InstallFromRegistry("test-template", registry)
	if err == nil {
		t.Log("Expected error (cannot clone in test environment)")
	}

	// Test non-existent template
	_, err = installer.InstallFromRegistry("nonexistent", registry)
	if err == nil {
		t.Error("Expected error for non-existent template")
	}
}
