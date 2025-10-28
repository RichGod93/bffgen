package commands

import (
	"os"
	"path/filepath"
	"testing"
)

// TestProjectNameValidation tests various project name validations
func TestProjectNameValidation(t *testing.T) {
	validator := NewProjectNameValidator()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// Valid names
		{"simple", "myproject", false},
		{"with underscores", "my_project", false},
		{"with hyphens", "my-project", false},
		{"with numbers", "myproject123", false},
		{"underscore start", "_project", false},
		{"mixed case", "MyProject", false},

		// Invalid names
		{"too short", "a", true},
		{"empty", "", true},
		{"too long", "a" + string(make([]byte, 100)), true},
		{"starts with number", "1project", true},
		{"starts with hyphen", "-project", true},
		{"special chars", "my@project", true},
		{"spaces", "my project", true},
		{"reserved go", "go", true},
		{"reserved mod", "mod", true},
		{"reserved bffgen", "bffgen", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestServiceNameValidation tests service name validation
func TestServiceNameValidation(t *testing.T) {
	validator := NewServiceNameValidator()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "auth-service", false},
		{"simple", "auth", false},
		{"with underscores", "auth_service", false},
		{"with numbers", "service123", false},

		{"empty", "", true},
		{"starts with number", "1service", true},
		{"special chars", "service@", true},
		{"too long", "a" + string(make([]byte, 200)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestURLValidation tests URL validation
func TestURLValidation(t *testing.T) {
	validator := NewURLValidator()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"http", "http://localhost:8080", false},
		{"https", "https://api.example.com", false},
		{"with path", "https://example.com/api/v1", false},

		{"empty", "", true},
		{"no protocol", "example.com", true},
		{"invalid protocol", "ftp://example.com", true},
		{"spaces", "http://example .com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDirectoryManager tests directory creation
func TestDirectoryManager(t *testing.T) {
	testDir := t.TempDir()
	defer os.RemoveAll(testDir)

	manager := NewDirectoryManager()

	t.Run("CreateDirectory", func(t *testing.T) {
		path := filepath.Join(testDir, "test", "nested", "dir")
		err := manager.CreateDirectory(path)
		if err != nil {
			t.Fatalf("CreateDirectory() error = %v", err)
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Error("Directory was not created")
		}

		// Verify permissions
		info, _ := os.Stat(path)
		if info.Mode().Perm() != manager.perm {
			t.Errorf("Expected permissions %o, got %o", manager.perm, info.Mode().Perm())
		}
	})

	t.Run("CreateDirectories", func(t *testing.T) {
		paths := []string{
			filepath.Join(testDir, "dir1"),
			filepath.Join(testDir, "dir2"),
			filepath.Join(testDir, "dir3"),
		}

		err := manager.CreateDirectories(paths...)
		if err != nil {
			t.Fatalf("CreateDirectories() error = %v", err)
		}

		for _, path := range paths {
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("Directory %s was not created", path)
			}
		}
	})

	t.Run("SafeCreate existing", func(t *testing.T) {
		path := filepath.Join(testDir, "existing")
		os.MkdirAll(path, 0755)

		// Should not error when directory exists
		err := manager.SafeCreate(path)
		if err != nil {
			t.Errorf("SafeCreate() error = %v", err)
		}
	})

	t.Run("SafeCreate new", func(t *testing.T) {
		path := filepath.Join(testDir, "newdir")

		err := manager.SafeCreate(path)
		if err != nil {
			t.Fatalf("SafeCreate() error = %v", err)
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Error("Directory was not created")
		}
	})
}

// TestRuntimeDetectorValidation tests runtime string validation
func TestRuntimeDetectorValidation(t *testing.T) {
	detector := NewRuntimeDetector()

	tests := []struct {
		name    string
		runtime string
		valid   bool
	}{
		{"go", "go", true},
		{"nodejs", "nodejs", true},
		{"nodejs-express", "nodejs-express", true},
		{"nodejs-fastify", "nodejs-fastify", true},
		{"invalid", "rust", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detector.IsValidRuntime(tt.runtime); got != tt.valid {
				t.Errorf("IsValidRuntime() = %v, want %v", got, tt.valid)
			}
		})
	}
}

// TestRuntimeNormalization tests runtime string normalization
func TestRuntimeNormalization(t *testing.T) {
	detector := NewRuntimeDetector()

	tests := []struct {
		name       string
		input      string
		wantOutput string
		wantErr    bool
	}{
		{"go", "go", "go", false},
		{"golang", "golang", "go", false},
		{"nodejs", "nodejs", "nodejs", false},
		{"node", "node", "nodejs", false},
		{"express", "express", "nodejs-express", false},
		{"node-express", "node-express", "nodejs-express", false},
		{"fastify", "fastify", "nodejs-fastify", false},
		{"with spaces", "  go  ", "go", false},

		{"invalid", "invalid", "", true},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := detector.NormalizeRuntime(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeRuntime() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.wantOutput {
				t.Errorf("NormalizeRuntime() = %v, want %v", got, tt.wantOutput)
			}
		})
	}
}

// TestRuntimeDetection tests project runtime detection
func TestRuntimeDetection(t *testing.T) {
	testDir := t.TempDir()
	defer os.RemoveAll(testDir)

	detector := NewRuntimeDetector()

	t.Run("detect go from go.mod", func(t *testing.T) {
		os.Create(filepath.Join(testDir, "go.mod"))
		runtime, err := detector.DetectRuntime(testDir)
		if err != nil {
			t.Fatalf("DetectRuntime() error = %v", err)
		}
		if runtime != "go" {
			t.Errorf("DetectRuntime() = %v, want go", runtime)
		}
	})

	t.Run("detect nodejs from package.json", func(t *testing.T) {
		dir := filepath.Join(testDir, "node-test")
		os.Mkdir(dir, 0755)
		os.Create(filepath.Join(dir, "package.json"))

		runtime, err := detector.DetectRuntime(dir)
		if err != nil {
			t.Fatalf("DetectRuntime() error = %v", err)
		}
		if runtime != "nodejs" {
			t.Errorf("DetectRuntime() = %v, want nodejs", runtime)
		}
	})

	t.Run("no config found", func(t *testing.T) {
		dir := filepath.Join(testDir, "empty")
		os.Mkdir(dir, 0755)

		_, err := detector.DetectRuntime(dir)
		if err == nil {
			t.Error("DetectRuntime() expected error for empty project")
		}
	})
}
