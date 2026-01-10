package registry

import (
	"testing"
)

func TestLoadManifest(t *testing.T) {
	manifest, err := LoadManifest()
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	if manifest == nil {
		t.Fatal("Manifest is nil")
	}

	// Check metadata
	if manifest.Metadata.Version == "" {
		t.Error("Manifest version is empty")
	}
}

func TestGetVersion(t *testing.T) {
	manifest, err := LoadManifest()
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	tests := []struct {
		runtime     string
		packageName string
		wantEmpty   bool
	}{
		{"nodejs-express", "express", false},
		{"nodejs-express", "helmet", false},
		{"nodejs-fastify", "fastify", false},
		{"python-fastapi", "fastapi", false},
		{"nodejs-express", "nonexistent-package", true},
		{"nonexistent-runtime", "express", true},
	}

	for _, tt := range tests {
		t.Run(tt.runtime+"/"+tt.packageName, func(t *testing.T) {
			version := manifest.GetVersion(tt.runtime, tt.packageName)
			if tt.wantEmpty && version != "" {
				t.Errorf("Expected empty version, got %s", version)
			}
			if !tt.wantEmpty && version == "" {
				t.Errorf("Expected non-empty version for %s/%s", tt.runtime, tt.packageName)
			}
		})
	}
}

func TestGetAllVersions(t *testing.T) {
	manifest, err := LoadManifest()
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	// Test Express runtime
	expressVersions := manifest.GetAllVersions("nodejs-express")
	if len(expressVersions) == 0 {
		t.Error("No versions found for nodejs-express")
	}

	// Should have at least express, helmet, cors
	requiredPackages := []string{"express", "helmet", "cors"}
	for _, pkg := range requiredPackages {
		if _, ok := expressVersions[pkg]; !ok {
			t.Errorf("Missing required package %s in nodejs-express", pkg)
		}
	}

	// Test non-existent runtime
	emptyVersions := manifest.GetAllVersions("nonexistent")
	if len(emptyVersions) != 0 {
		t.Error("Expected empty map for nonexistent runtime")
	}
}

func TestGetRuntimes(t *testing.T) {
	manifest, err := LoadManifest()
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	runtimes := manifest.GetRuntimes()
	if len(runtimes) == 0 {
		t.Error("No runtimes found in manifest")
	}

	// Should have at least these runtimes
	expectedRuntimes := []string{"nodejs-express", "nodejs-fastify", "python-fastapi"}
	for _, expected := range expectedRuntimes {
		found := false
		for _, runtime := range runtimes {
			if runtime == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected runtime %s not found in manifest", expected)
		}
	}

	// Should NOT include metadata or common sections
	excludedSections := []string{"metadata", "nodejs-common", "python-common"}
	for _, excluded := range excludedSections {
		for _, runtime := range runtimes {
			if runtime == excluded {
				t.Errorf("Runtime list should not include %s", excluded)
			}
		}
	}
}
