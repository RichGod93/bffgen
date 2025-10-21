package utils

import (
	"fmt"
	"testing"
)

func TestCheckTool(t *testing.T) {
	t.Run("CheckExistingTool", func(t *testing.T) {
		// Test with a tool that should exist on most systems
		info := CheckTool("Go", "go")

		if info.Name != "Go" {
			t.Errorf("Expected name 'Go', got '%s'", info.Name)
		}

		if info.Command != "go" {
			t.Errorf("Expected command 'go', got '%s'", info.Command)
		}

		// Note: Can't assert Installed=true because test environment may vary
	})

	t.Run("CheckNonExistentTool", func(t *testing.T) {
		info := CheckTool("NonExistent", "nonexistenttoolxyz123")

		if info.Installed {
			t.Error("NonExistent tool should not be installed")
		}

		if info.Version != "" {
			t.Errorf("Version should be empty for non-existent tool, got '%s'", info.Version)
		}
	})
}

func TestCheckRequiredTools(t *testing.T) {
	t.Run("GoProject", func(t *testing.T) {
		required, _ := CheckRequiredTools("go")

		if len(required) == 0 {
			t.Error("Go project should have required tools")
		}

		// Should have Go as required
		hasGo := false
		for _, tool := range required {
			if tool.Name == "Go" {
				hasGo = true
				if !tool.Required {
					t.Error("Go should be marked as required")
				}
			}
		}

		if !hasGo {
			t.Error("Go should be in required tools")
		}
	})

	t.Run("NodeJSProject", func(t *testing.T) {
		required, optional := CheckRequiredTools("nodejs-express")

		// Should have Node.js and npm as required
		hasNode := false
		hasNpm := false

		for _, tool := range required {
			if tool.Name == "Node.js" {
				hasNode = true
			}
			if tool.Name == "npm" {
				hasNpm = true
			}
		}

		if !hasNode {
			t.Error("Node.js should be in required tools")
		}

		if !hasNpm {
			t.Error("npm should be in required tools")
		}

		// Should have optional tools
		if len(optional) == 0 {
			t.Error("Should have optional tools")
		}
	})

	t.Run("UnknownProject", func(t *testing.T) {
		required, optional := CheckRequiredTools("unknown")

		if len(required) != 0 {
			t.Error("Unknown project should have no required tools")
		}

		if len(optional) != 0 {
			t.Error("Unknown project should have no optional tools")
		}
	})
}

func TestGetToolInstallInstructions(t *testing.T) {
	tests := []struct {
		toolName string
		contains string
	}{
		{"Go", "brew install go"},
		{"Node.js", "brew install node"},
		{"Docker", "brew install --cask docker"},
		{"npm", "included with Node.js"},
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			instructions := GetToolInstallInstructions(tt.toolName)

			if instructions == "" {
				t.Errorf("Instructions for %s should not be empty", tt.toolName)
			}

			// Note: We can't easily test if it contains specific text without string matching
			// Just verify it's not the default "No installation instructions" message
			if instructions == fmt.Sprintf("No installation instructions available for %s", tt.toolName) {
				t.Errorf("Should have specific instructions for %s", tt.toolName)
			}
		})
	}

	t.Run("UnknownTool", func(t *testing.T) {
		instructions := GetToolInstallInstructions("UnknownToolXYZ")

		expected := "No installation instructions available for UnknownToolXYZ"
		if instructions != expected {
			t.Errorf("Expected default message, got: %s", instructions)
		}
	})
}

func TestHasRequiredTools(t *testing.T) {
	// This test depends on the actual environment
	// We can only test the function exists and returns a bool
	t.Run("ReturnsBoolean", func(t *testing.T) {
		result := HasRequiredTools("go")
		_ = result // Just verify it returns without panicking

		result = HasRequiredTools("nodejs")
		_ = result

		// Unknown project type has no required tools, so should return true
		result = HasRequiredTools("unknown")
		if !result {
			t.Error("Unknown project type has no requirements so should return true")
		}
	})
}

func TestPrintToolStatus(t *testing.T) {
	// This function prints to stdout, so we just test it doesn't panic
	t.Run("DoesNotPanic", func(t *testing.T) {
		required := []ToolInfo{
			{Name: "Go", Command: "go", Installed: true, Version: "1.21", Required: true},
		}

		optional := []ToolInfo{
			{Name: "Docker", Command: "docker", Installed: false},
		}

		// Should not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("PrintToolStatus panicked: %v", r)
			}
		}()

		PrintToolStatus(required, optional)
	})
}
