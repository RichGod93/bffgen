package commands

import (
	"testing"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		shouldErr bool
	}{
		{"Valid HTTP", "http://localhost:3000", false},
		{"Valid HTTPS", "https://api.example.com", false},
		{"Valid with path", "http://localhost:3000/api", false},
		{"No scheme", "localhost:3000", true},
		{"Invalid scheme", "ftp://localhost:3000", true},
		{"No host", "http://", true},
		{"Empty URL", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURL(tt.url)
			if tt.shouldErr && err == nil {
				t.Errorf("Expected error for URL: %s", tt.url)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error for URL %s: %v", tt.url, err)
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		shouldErr bool
	}{
		{"Valid simple path", "/api/users", false},
		{"Valid with params", "/api/users/{id}", false},
		{"Valid with dashes", "/api/user-profiles", false},
		{"Valid with dots", "/api/v1.0/users", false},
		{"No leading slash", "api/users", true},
		{"With spaces", "/api/my users", true},
		{"Empty path", "", true},
		{"Special chars", "/api/users@admin", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePath(tt.path)
			if tt.shouldErr && err == nil {
				t.Errorf("Expected error for path: %s", tt.path)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error for path %s: %v", tt.path, err)
			}
		})
	}
}
