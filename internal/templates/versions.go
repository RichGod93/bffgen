package templates

import (
	"fmt"

	"github.com/RichGod93/bffgen/internal/registry"
)

// NewTemplateDataWithVersions creates TemplateData populated with dependency versions
// from the manifest for the given runtime
func NewTemplateDataWithVersions(runtime string) (*TemplateData, error) {
	manifest, err := registry.LoadManifest()
	if err != nil {
		return nil, fmt.Errorf("failed to load manifest: %w", err)
	}

	versions := manifest.GetAllVersions(runtime)
	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions found for runtime: %s", runtime)
	}

	return &TemplateData{
		Runtime:  runtime,
		Versions: versions,
	}, nil
}

// GetVersion is a template helper function to retrieve a package version
// Safe to use in templates - returns empty string if package not found
func (td *TemplateData) GetVersion(packageName string) string {
	if td.Versions == nil {
		return ""
	}
	if version, ok := td.Versions[packageName]; ok {
		return version
	}
	return ""
}
