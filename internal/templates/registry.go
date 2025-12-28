package templates

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Registry represents the template registry
type Registry struct {
	Templates   []RegistryEntry `json:"templates"`
	LastUpdated time.Time       `json:"last_updated"`
}

// RegistryEntry represents a single template in the registry
type RegistryEntry struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	URL         string   `json:"url"` // Git URL
	Author      string   `json:"author"`
	Tags        []string `json:"tags"`
	Language    string   `json:"language"`
	Version     string   `json:"version"`
}

// DefaultRegistryURL is the URL to the official registry
const DefaultRegistryURL = "https://raw.githubusercontent.com/RichGod93/bffgen/main/templates/registry.json"

// LoadRegistry loads the registry from local cache or initializes a new one
func LoadRegistry(templatesDir string) (*Registry, error) {
	registryPath := filepath.Join(templatesDir, "registry.json")

	// Check if registry exists
	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		return &Registry{
			Templates:   []RegistryEntry{},
			LastUpdated: time.Now(),
		}, nil
	}

	data, err := os.ReadFile(registryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read registry: %w", err)
	}

	var registry Registry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("failed to parse registry: %w", err)
	}

	return &registry, nil
}

// UpdateRegistry fetches the latest registry from the remote URL
func (r *Registry) Update(templatesDir string) error {
	resp, err := http.Get(DefaultRegistryURL)
	if err != nil {
		return fmt.Errorf("failed to fetch registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch registry: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var newRegistry Registry
	if err := json.Unmarshal(body, &newRegistry); err != nil {
		// Try to unmarshal as list of templates directly (common pattern)
		var templates []RegistryEntry
		if err := json.Unmarshal(body, &templates); err == nil {
			newRegistry = Registry{
				Templates: templates,
			}
		} else {
			return fmt.Errorf("failed to parse remote registry: %w", err)
		}
	}

	r.Templates = newRegistry.Templates
	r.LastUpdated = time.Now()

	// Save to local cache
	return r.Save(templatesDir)
}

// Save saves the registry to the local cache
func (r *Registry) Save(templatesDir string) error {
	registryPath := filepath.Join(templatesDir, "registry.json")

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	return os.WriteFile(registryPath, data, 0644)
}

// Find returns a registry entry by name
func (r *Registry) Find(name string) *RegistryEntry {
	for _, entry := range r.Templates {
		if entry.Name == name {
			return &entry
		}
	}
	return nil
}
