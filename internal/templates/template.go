package templates

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Template represents a project template
type Template struct {
	Name        string             `yaml:"name"`
	Version     string             `yaml:"version"`
	Description string             `yaml:"description"`
	Author      string             `yaml:"author"`
	Category    string             `yaml:"category"`
	Language    string             `yaml:"language"`
	Features    []string           `yaml:"features"`
	Variables   []TemplateVariable `yaml:"variables"`
	Files       []string           `yaml:"files"`
	PostInstall []string           `yaml:"post_install"`

	// Internal fields
	Path string `yaml:"-"` // Path to template directory
}

// TemplateVariable represents a configurable variable in a template
type TemplateVariable struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Default     string `yaml:"default"`
	Required    bool   `yaml:"required"`
}

// LoadTemplate loads a template from a directory
func LoadTemplate(templatePath string) (*Template, error) {
	manifestPath := filepath.Join(templatePath, "template.yaml")

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template manifest: %w", err)
	}

	var template Template
	if err := yaml.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("failed to parse template manifest: %w", err)
	}

	template.Path = templatePath

	// Validate template
	if err := template.Validate(); err != nil {
		return nil, fmt.Errorf("invalid template: %w", err)
	}

	return &template, nil
}

// Validate checks if the template is valid
func (t *Template) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("template name is required")
	}
	if t.Version == "" {
		return fmt.Errorf("template version is required")
	}
	if t.Language == "" {
		return fmt.Errorf("template language is required")
	}

	// Validate required variables
	for _, v := range t.Variables {
		if v.Name == "" {
			return fmt.Errorf("variable name is required")
		}
	}

	return nil
}

// GetRequiredVariables returns variables that are required
func (t *Template) GetRequiredVariables() []TemplateVariable {
	var required []TemplateVariable
	for _, v := range t.Variables {
		if v.Required {
			required = append(required, v)
		}
	}
	return required
}

// GetVariableValue gets a variable value with fallback to default
func (t *Template) GetVariableValue(name string, provided map[string]string) string {
	if val, ok := provided[name]; ok {
		return val
	}

	for _, v := range t.Variables {
		if v.Name == name {
			return v.Default
		}
	}

	return ""
}
