package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// NPMPackageInfo represents NPM package metadata
type NPMPackageInfo struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	DistTags    map[string]string `json:"dist-tags"`
}

// NPMClient handles interactions with the NPM registry
type NPMClient struct {
	baseURL string
	client  *http.Client
}

// NewNPMClient creates a new NPM registry client
func NewNPMClient() *NPMClient {
	return &NPMClient{
		baseURL: "https://registry.npmjs.org",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetLatestVersion fetches the latest version of an NPM package
func (c *NPMClient) GetLatestVersion(packageName string) (string, error) {
	url := fmt.Sprintf("%s/%s/latest", c.baseURL, packageName)

	resp, err := c.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch package %s: %w", packageName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("npm registry returned status %d for package %s", resp.StatusCode, packageName)
	}

	var pkg NPMPackageInfo
	if err := json.NewDecoder(resp.Body).Decode(&pkg); err != nil {
		return "", fmt.Errorf("failed to decode package info for %s: %w", packageName, err)
	}

	return pkg.Version, nil
}

// GetLatestVersionWithPrefix returns the version with a caret prefix (^)
func (c *NPMClient) GetLatestVersionWithPrefix(packageName string) (string, error) {
	version, err := c.GetLatestVersion(packageName)
	if err != nil {
		return "", err
	}
	return "^" + version, nil
}

// FetchAllVersions fetches latest versions for multiple packages concurrently
func (c *NPMClient) FetchAllVersions(packages []string) (map[string]string, error) {
	type result struct {
		pkg string
		ver string
		err error
	}

	results := make(chan result, len(packages))

	// Fetch concurrently
	for _, pkg := range packages {
		go func(p string) {
			v, e := c.GetLatestVersionWithPrefix(p)
			results <- result{pkg: p, ver: v, err: e}
		}(pkg)
	}

	// Collect results
	versions := make(map[string]string)
	var errors []error

	for i := 0; i < len(packages); i++ {
		r := <-results
		if r.err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", r.pkg, r.err))
		} else {
			versions[r.pkg] = r.ver
		}
	}

	if len(errors) > 0 {
		return versions, fmt.Errorf("failed to fetch %d packages: %v", len(errors), errors[0])
	}

	return versions, nil
}
