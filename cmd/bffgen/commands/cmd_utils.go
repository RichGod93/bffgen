package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/RichGod93/bffgen/internal/types"
	"github.com/RichGod93/bffgen/internal/utils"
)

// ProjectNameValidator validates project names for safety and compatibility
type ProjectNameValidator struct {
	minLength int
	maxLength int
	pattern   *regexp.Regexp
}

// NewProjectNameValidator creates a new project name validator
func NewProjectNameValidator() *ProjectNameValidator {
	// Allow alphanumeric, hyphens, and underscores
	// Must start with letter or underscore
	pattern := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*$`)

	return &ProjectNameValidator{
		minLength: 2,
		maxLength: 50,
		pattern:   pattern,
	}
}

// Validate checks if a project name is valid
func (v *ProjectNameValidator) Validate(name string) error {
	// Check length
	if len(name) < v.minLength {
		return fmt.Errorf("project name must be at least %d characters", v.minLength)
	}
	if len(name) > v.maxLength {
		return fmt.Errorf("project name must not exceed %d characters", v.maxLength)
	}

	// Check pattern
	if !v.pattern.MatchString(name) {
		return fmt.Errorf("project name must start with letter or underscore and contain only alphanumeric characters, hyphens, and underscores")
	}

	// Check for reserved names
	reserved := []string{"go", "mod", "sum", "lock", "bffgen", "node_modules"}
	for _, reserved := range reserved {
		if strings.EqualFold(name, reserved) {
			return fmt.Errorf("'%s' is a reserved project name", name)
		}
	}

	return nil
}

// ServiceNameValidator validates service names
type ServiceNameValidator struct {
	pattern *regexp.Regexp
}

// NewServiceNameValidator creates a new service name validator
func NewServiceNameValidator() *ServiceNameValidator {
	pattern := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)

	return &ServiceNameValidator{
		pattern: pattern,
	}
}

// Validate checks if a service name is valid
func (v *ServiceNameValidator) Validate(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("service name cannot be empty")
	}

	if len(name) > 100 {
		return fmt.Errorf("service name must not exceed 100 characters")
	}

	if !v.pattern.MatchString(name) {
		return fmt.Errorf("service name must start with letter and contain only alphanumeric characters, hyphens, and underscores")
	}

	return nil
}

// URLValidator validates URLs for basic correctness
type URLValidator struct {
	pattern *regexp.Regexp
}

// NewURLValidator creates a new URL validator
func NewURLValidator() *URLValidator {
	// Simple URL pattern for basic validation
	pattern := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)

	return &URLValidator{
		pattern: pattern,
	}
}

// Validate checks if a URL is valid
func (v *URLValidator) Validate(url string) error {
	if len(url) == 0 {
		return fmt.Errorf("URL cannot be empty")
	}

	if !v.pattern.MatchString(url) {
		return fmt.Errorf("URL must start with http:// or https:// and be properly formatted")
	}

	if len(url) > 2000 {
		return fmt.Errorf("URL is too long (max 2000 characters)")
	}

	return nil
}

// ConfigLoader encapsulates config loading logic with validation
type ConfigLoader struct {
	validator *ProjectNameValidator
}

// NewConfigLoader creates a new config loader
func NewConfigLoader() *ConfigLoader {
	return &ConfigLoader{
		validator: NewProjectNameValidator(),
	}
}

// LoadBFFConfig loads BFF config from file with error handling
func (c *ConfigLoader) LoadBFFConfig(configPath string) (*types.BFFConfig, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	config, err := utils.LoadConfig(configPath)
	if err != nil {
		return nil, ErrorContext(err, "loading BFF config")
	}

	return config, nil
}

// LoadBFFGenConfig loads global BFF generator config with defaults
func (c *ConfigLoader) LoadBFFGenConfig() (*types.BFFGenConfig, error) {
	config, err := utils.LoadBFFGenConfig()
	if err != nil {
		LogWarning(fmt.Sprintf("Could not load global config: %v", err))
		// Return default config instead of failing
		return types.GetDefaultConfig(), nil
	}

	return config, nil
}

// ValidateProject checks if a directory is a valid BFF project
func (c *ConfigLoader) ValidateProject(projectDir string) error {
	// Check for config files
	bffConfigPath := filepath.Join(projectDir, "bff.config.yaml")
	bffGenConfigPath := filepath.Join(projectDir, "bffgen.config.json")

	bffConfigExists := isFile(bffConfigPath)
	bffGenConfigExists := isFile(bffGenConfigPath)

	if !bffConfigExists && !bffGenConfigExists {
		return fmt.Errorf("project directory must contain bff.config.yaml or bffgen.config.json")
	}

	return nil
}

// isFile checks if a path is a regular file
func isFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}

// DirectoryManager handles directory creation with proper permissions
type DirectoryManager struct {
	perm os.FileMode
}

// NewDirectoryManager creates a new directory manager
func NewDirectoryManager() *DirectoryManager {
	return &DirectoryManager{
		perm: utils.ProjectDirPerm,
	}
}

// CreateDirectory creates a directory with proper permissions
func (d *DirectoryManager) CreateDirectory(path string) error {
	if err := os.MkdirAll(path, d.perm); err != nil {
		return ErrorContext(err, fmt.Sprintf("creating directory %s", path))
	}

	return nil
}

// CreateDirectories creates multiple directories
func (d *DirectoryManager) CreateDirectories(paths ...string) error {
	for _, path := range paths {
		if err := d.CreateDirectory(path); err != nil {
			return err
		}
	}

	return nil
}

// SafeCreate creates a directory only if it doesn't exist
func (d *DirectoryManager) SafeCreate(path string) error {
	if _, err := os.Stat(path); err == nil {
		// Directory exists
		return nil
	} else if !os.IsNotExist(err) {
		// Error checking directory
		return ErrorContext(err, "checking directory")
	}

	// Directory doesn't exist, create it
	return d.CreateDirectory(path)
}

// RuntimeDetector detects project runtime with validation
type RuntimeDetector struct {
	validRuntimes map[string]bool
}

// NewRuntimeDetector creates a new runtime detector
func NewRuntimeDetector() *RuntimeDetector {
	return &RuntimeDetector{
		validRuntimes: map[string]bool{
			"go":             true,
			"nodejs":         true,
			"nodejs-express": true,
			"nodejs-fastify": true,
			"node":           true,
			"node-express":   true,
			"node-fastify":   true,
			"express":        true,
			"fastify":        true,
		},
	}
}

// DetectRuntime detects the runtime of a project
func (r *RuntimeDetector) DetectRuntime(projectDir string) (string, error) {
	// Check for config files
	if _, err := os.Stat(filepath.Join(projectDir, "bffgen.config.json")); err == nil {
		return "nodejs", nil
	}

	if _, err := os.Stat(filepath.Join(projectDir, "bff.config.yaml")); err == nil {
		return "go", nil
	}

	// Check for package.json
	if _, err := os.Stat(filepath.Join(projectDir, "package.json")); err == nil {
		return "nodejs", nil
	}

	// Check for go.mod
	if _, err := os.Stat(filepath.Join(projectDir, "go.mod")); err == nil {
		return "go", nil
	}

	return "", fmt.Errorf("could not detect project runtime")
}

// IsValidRuntime checks if a runtime string is valid
func (r *RuntimeDetector) IsValidRuntime(runtime string) bool {
	return r.validRuntimes[strings.ToLower(runtime)]
}

// NormalizeRuntime normalizes runtime strings to standard format
func (r *RuntimeDetector) NormalizeRuntime(runtime string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(runtime))

	// Map common variations
	switch normalized {
	case "go", "golang":
		return "go", nil
	case "nodejs", "node", "js", "javascript":
		return "nodejs", nil
	case "express", "node-express", "nodejs-express":
		return "nodejs-express", nil
	case "fastify", "node-fastify", "nodejs-fastify":
		return "nodejs-fastify", nil
	default:
		return "", fmt.Errorf("invalid runtime: %s", runtime)
	}
}
