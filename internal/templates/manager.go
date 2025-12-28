package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Manager manages project templates
type Manager struct {
	templatesDir string
	cache        map[string]*Template
}

// NewManager creates a new template manager
func NewManager(templatesDir string) *Manager {
	return &Manager{
		templatesDir: templatesDir,
		cache:        make(map[string]*Template),
	}
}

// GetDefaultTemplatesDir returns the default templates directory
func GetDefaultTemplatesDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".bffgen", "templates"), nil
}

// List returns all available templates
func (m *Manager) List() ([]*Template, error) {
	var templates []*Template

	// Check bundled templates (shipped with bffgen)
	if bundled, err := m.loadBundledTemplates(); err == nil {
		templates = append(templates, bundled...)
	}

	// Check built-in templates in user directory
	builtInDir := filepath.Join(m.templatesDir, "built-in")
	if builtIn, err := m.loadTemplatesFromDir(builtInDir); err == nil {
		templates = append(templates, builtIn...)
	}

	// Check community templates
	communityDir := filepath.Join(m.templatesDir, "community")
	if community, err := m.loadTemplatesFromDir(communityDir); err == nil {
		templates = append(templates, community...)
	}

	// Sort by name
	sort.Slice(templates, func(i, j int) bool {
		return templates[i].Name < templates[j].Name
	})

	return templates, nil
}

// loadBundledTemplates loads templates bundled with bffgen
func (m *Manager) loadBundledTemplates() ([]*Template, error) {
	// Try to find bundled templates relative to executable
	execPath, err := os.Executable()
	if err != nil {
		return nil, err
	}

	execDir := filepath.Dir(execPath)
	bundledDir := filepath.Join(execDir, "templates")

	// If not found relative to executable, try working directory
	if _, err := os.Stat(bundledDir); os.IsNotExist(err) {
		bundledDir = "templates"
	}

	return m.loadTemplatesFromDir(bundledDir)
}

// Get retrieves a template by name
func (m *Manager) Get(name string) (*Template, error) {
	// Check cache first
	if cached, ok := m.cache[name]; ok {
		return cached, nil
	}

	// Search in bundled templates
	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)
	bundledPath := filepath.Join(execDir, "templates", name)
	if template, err := LoadTemplate(bundledPath); err == nil {
		m.cache[name] = template
		return template, nil
	}

	// Try working directory
	workingBundled := filepath.Join("templates", name)
	if template, err := LoadTemplate(workingBundled); err == nil {
		m.cache[name] = template
		return template, nil
	}

	// Search in built-in templates
	builtInPath := filepath.Join(m.templatesDir, "built-in", name)
	if template, err := LoadTemplate(builtInPath); err == nil {
		m.cache[name] = template
		return template, nil
	}

	// Search in community templates
	communityPath := filepath.Join(m.templatesDir, "community", name)
	if template, err := LoadTemplate(communityPath); err == nil {
		m.cache[name] = template
		return template, nil
	}

	return nil, fmt.Errorf("template '%s' not found", name)
}

// Exists checks if a template exists
func (m *Manager) Exists(name string) bool {
	_, err := m.Get(name)
	return err == nil
}

// loadTemplatesFromDir loads all templates from a directory
func (m *Manager) loadTemplatesFromDir(dir string) ([]*Template, error) {
	var templates []*Template

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		templatePath := filepath.Join(dir, entry.Name())
		template, err := LoadTemplate(templatePath)
		if err != nil {
			// Skip invalid templates
			continue
		}

		templates = append(templates, template)
	}

	return templates, nil
}

// EnsureTemplatesDir creates the templates directory if it doesn't exist
func (m *Manager) EnsureTemplatesDir() error {
	dirs := []string{
		filepath.Join(m.templatesDir, "built-in"),
		filepath.Join(m.templatesDir, "community"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// GetRegistry returns the template registry, loading it if necessary
func (m *Manager) GetRegistry() (*Registry, error) {
	return LoadRegistry(m.templatesDir)
}

// UpdateRegistry updates the local registry cache from the remote source
func (m *Manager) UpdateRegistry() error {
	registry, err := m.GetRegistry()
	if err != nil {
		return err
	}
	return registry.Update(m.templatesDir)
}
