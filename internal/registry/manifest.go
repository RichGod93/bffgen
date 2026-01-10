package registry

import (
	"embed"
	"fmt"

	"gopkg.in/yaml.v3"
)

//go:embed versions.yaml
var manifestFS embed.FS

// Manifest represents the dependency version manifest
type Manifest struct {
	Metadata Metadata                     `yaml:"metadata"`
	Runtimes map[string]map[string]string `yaml:",inline"`
}

// Metadata contains manifest metadata
type Metadata struct {
	Version       string `yaml:"version"`
	LastUpdated   string `yaml:"last_updated"`
	BffgenVersion string `yaml:"bffgen_version"`
}

// LoadManifest loads the embedded versions.yaml file
func LoadManifest() (*Manifest, error) {
	data, err := manifestFS.ReadFile("versions.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read versions.yaml: %w", err)
	}

	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse versions.yaml: %w", err)
	}

	return &manifest, nil
}

// GetVersion returns the version for a specific runtime and package
// Returns empty string if not found
func (m *Manifest) GetVersion(runtime, packageName string) string {
	if runtimeDeps, ok := m.Runtimes[runtime]; ok {
		if version, ok := runtimeDeps[packageName]; ok {
			return version
		}
	}
	return ""
}

// GetAllVersions returns all dependency versions for a runtime
func (m *Manifest) GetAllVersions(runtime string) map[string]string {
	if versions, ok := m.Runtimes[runtime]; ok {
		return versions
	}
	return make(map[string]string)
}

// GetRuntimes returns list of all available runtimes in manifest
func (m *Manifest) GetRuntimes() []string {
	runtimes := make([]string, 0, len(m.Runtimes))
	for runtime := range m.Runtimes {
		// Skip metadata and common sections
		if runtime != "metadata" && runtime != "nodejs-common" && runtime != "python-common" {
			runtimes = append(runtimes, runtime)
		}
	}
	return runtimes
}
