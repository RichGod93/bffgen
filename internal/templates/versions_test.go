package templates

import (
	"testing"
)

func TestNewTemplateDataWithVersions(t *testing.T) {
	data, err := NewTemplateDataWithVersions("nodejs-express")
	if err != nil {
		t.Fatalf("Failed to create template data: %v", err)
	}

	if data == nil {
		t.Fatal("Template data is nil")
	}

	if data.Versions == nil {
		t.Fatal("Versions map is nil")
	}

	if len(data.Versions) == 0 {
		t.Fatal("Versions map is empty")
	}

	// Check for expected packages
	requiredPackages := []string{"express", "helmet", "cors", "dotenv"}
	for _, pkg := range requiredPackages {
		if version := data.GetVersion(pkg); version == "" {
			t.Errorf("Missing version for required package: %s", pkg)
		} else {
			t.Logf("✓ %s: %s", pkg, version)
		}
	}
}

func TestGetVersion(t *testing.T) {
	data := &TemplateData{
		Versions: map[string]string{
			"express": "^4.19.2",
			"helmet":  "^8.0.0",
		},
	}

	tests := []struct {
		pkg  string
		want string
	}{
		{"express", "^4.19.2"},
		{"helmet", "^8.0.0"},
		{"nonexistent", ""},
	}

	for _, tt := range tests {
		t.Run(tt.pkg, func(t *testing.T) {
			got := data.GetVersion(tt.pkg)
			if got != tt.want {
				t.Errorf("GetVersion(%q) = %q, want %q", tt.pkg, got, tt.want)
			}
		})
	}
}
