package templates

import (
	"fmt"
	"text/template"

	"github.com/RichGod93/bffgen/internal/registry"
)

// LoadVersions returns the dependency versions from the embedded manifest for the given runtime
func LoadVersions(runtime string) (map[string]string, error) {
	manifest, err := registry.LoadManifest()
	if err != nil {
		return nil, fmt.Errorf("failed to load manifest: %w", err)
	}

	versions := manifest.GetAllVersions(runtime)
	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions found for runtime: %s", runtime)
	}

	return versions, nil
}

// NewTemplateDataWithVersions creates TemplateData populated with dependency versions
// from the manifest for the given runtime
func NewTemplateDataWithVersions(runtime string) (*TemplateData, error) {
	versions, err := LoadVersions(runtime)
	if err != nil {
		return nil, err
	}

	return &TemplateData{
		Runtime:  runtime,
		Versions: versions,
	}, nil
}

// VersionFuncMap returns a template FuncMap whose GetVersion fails on packages missing from versions,
// so a template can never silently render an empty dependency version
func VersionFuncMap(versions map[string]string) template.FuncMap {
	return template.FuncMap{
		"GetVersion": func(packageName string) (string, error) {
			return requireVersion(versions, packageName)
		},
	}
}

// RequireVersion is the template-facing variant of GetVersion: it returns an error
// instead of an empty string when the package is not in the manifest
func (td *TemplateData) RequireVersion(packageName string) (string, error) {
	return requireVersion(td.Versions, packageName)
}

func requireVersion(versions map[string]string, packageName string) (string, error) {
	if version, ok := versions[packageName]; ok && version != "" {
		return version, nil
	}
	return "", fmt.Errorf("no version for package %q in versions manifest", packageName)
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
