package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// PyPIPackageInfo represents PyPI package metadata
type PyPIPackageInfo struct {
	Info PyPIInfo `json:"info"`
}

// PyPIInfo contains package information
type PyPIInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// PyPIClient handles interactions with the PyPI registry
type PyPIClient struct {
	baseURL string
	client  *http.Client
}

// NewPyPIClient creates a new PyPI registry client
func NewPyPIClient() *PyPIClient {
	return &PyPIClient{
		baseURL: "https://pypi.org/pypi",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetLatestVersion fetches the latest version of a PyPI package
func (c *PyPIClient) GetLatestVersion(packageName string) (string, error) {
	url := fmt.Sprintf("%s/%s/json", c.baseURL, packageName)

	resp, err := c.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch package %s: %w", packageName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("pypi returned status %d for package %s", resp.StatusCode, packageName)
	}

	var pkg PyPIPackageInfo
	if err := json.NewDecoder(resp.Body).Decode(&pkg); err != nil {
		return "", fmt.Errorf("failed to decode package info for %s: %w", packageName, err)
	}

	return pkg.Info.Version, nil
}

// GetLatestVersionWithRange returns the version with Python version range
// For example: "1.2.3" -> ">=1.2.3,<2.0.0" (allows minor updates, blocks major)
func (c *PyPIClient) GetLatestVersionWithRange(packageName string) (string, error) {
	version, err := c.GetLatestVersion(packageName)
	if err != nil {
		return "", err
	}

	// Parse major version
	var major, minor, patch int
	_, err = fmt.Sscanf(version, "%d.%d.%d", &major, &minor, &patch)
	if err != nil {
		// If parsing fails, return as-is with >=
		return ">=" + version, nil
	}

	// Allow minor and patch updates, but not major version bumps
	return fmt.Sprintf(">=%s,<%d.0.0", version, major+1), nil
}

// FetchAllVersions fetches latest versions for multiple packages concurrently
func (c *PyPIClient) FetchAllVersions(packages []string) (map[string]string, error) {
	type result struct {
		pkg string
		ver string
		err error
	}

	results := make(chan result, len(packages))

	// Fetch concurrently
	for _, pkg := range packages {
		go func(p string) {
			v, e := c.GetLatestVersionWithRange(p)
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
