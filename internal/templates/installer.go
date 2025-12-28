package templates

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Installer handles template installation from remote sources
type Installer struct {
	templatesDir string
}

// NewInstaller creates a new template installer
func NewInstaller(templatesDir string) *Installer {
	return &Installer{
		templatesDir: templatesDir,
	}
}

// InstallFromGitHub installs a template from a GitHub repository
// Supports formats:
//   - github.com/user/repo
//   - https://github.com/user/repo
//   - user/repo
func (i *Installer) InstallFromGitHub(repoURL string) (*Template, error) {
	// Normalize GitHub URL
	gitURL := i.normalizeGitHubURL(repoURL)

	// Extract template name from URL
	templateName := i.extractTemplateName(gitURL)

	// Check if template already exists
	templatePath := filepath.Join(i.templatesDir, "community", templateName)
	if _, err := os.Stat(templatePath); err == nil {
		return nil, fmt.Errorf("template '%s' already exists. Use 'bffgen template update %s' to update", templateName, templateName)
	}

	// Create community directory if it doesn't exist
	communityDir := filepath.Join(i.templatesDir, "community")
	if err := os.MkdirAll(communityDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create community directory: %w", err)
	}

	fmt.Printf("📦 Installing template from %s...\n", gitURL)

	// Clone repository
	if err := i.cloneRepository(gitURL, templatePath); err != nil {
		// Cleanup on failure
		os.RemoveAll(templatePath)
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	// Load and validate template
	template, err := LoadTemplate(templatePath)
	if err != nil {
		// Cleanup invalid template
		os.RemoveAll(templatePath)
		return nil, fmt.Errorf("invalid template: %w", err)
	}

	// Validate template structure
	if err := i.validateTemplate(template); err != nil {
		os.RemoveAll(templatePath)
		return nil, fmt.Errorf("template validation failed: %w", err)
	}

	fmt.Printf("✅ Template '%s' installed successfully!\n", template.Name)
	return template, nil
}

// InstallFromRegistry installs a template from the official registry
func (i *Installer) InstallFromRegistry(name string, registry *Registry) (*Template, error) {
	entry := registry.Find(name)
	if entry == nil {
		return nil, fmt.Errorf("template '%s' not found in registry", name)
	}

	fmt.Printf("📦 Installing '%s' from registry\n", name)
	fmt.Printf("   Author: %s\n", entry.Author)
	fmt.Printf("   Version: %s\n", entry.Version)
	fmt.Printf("   Description: %s\n\n", entry.Description)

	return i.InstallFromGitHub(entry.URL)
}

// Remove removes an installed template
func (i *Installer) Remove(name string) error {
	templatePath := filepath.Join(i.templatesDir, "community", name)

	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("template '%s' is not installed", name)
	}

	fmt.Printf("🗑️  Removing template '%s'...\n", name)

	if err := os.RemoveAll(templatePath); err != nil {
		return fmt.Errorf("failed to remove template: %w", err)
	}

	fmt.Printf("✅ Template '%s' removed successfully\n", name)
	return nil
}

// Update updates an installed template
func (i *Installer) Update(name string) error {
	templatePath := filepath.Join(i.templatesDir, "community", name)

	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("template '%s' is not installed", name)
	}

	fmt.Printf("🔄 Updating template '%s'...\n", name)

	// Check if it's a git repository
	gitDir := filepath.Join(templatePath, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		// Pull latest changes
		cmd := exec.Command("git", "pull")
		cmd.Dir = templatePath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to pull updates: %w", err)
		}

		fmt.Printf("✅ Template '%s' updated successfully\n", name)
		return nil
	}

	return fmt.Errorf("template '%s' is not a git repository and cannot be updated", name)
}

// normalizeGitHubURL converts various GitHub URL formats to a standard git URL
func (i *Installer) normalizeGitHubURL(repoURL string) string {
	// Remove trailing slashes
	repoURL = strings.TrimSuffix(repoURL, "/")

	// If it's already a proper git URL, return it
	if strings.HasPrefix(repoURL, "https://") || strings.HasPrefix(repoURL, "git@") {
		return repoURL
	}

	// If it starts with github.com, add https://
	if strings.HasPrefix(repoURL, "github.com/") {
		return "https://" + repoURL
	}

	// If it's just user/repo format, add github.com
	if strings.Count(repoURL, "/") == 1 && !strings.Contains(repoURL, "://") {
		return "https://github.com/" + repoURL
	}

	return repoURL
}

// extractTemplateName extracts the template name from a git URL
func (i *Installer) extractTemplateName(gitURL string) string {
	// Remove .git suffix if present
	gitURL = strings.TrimSuffix(gitURL, ".git")

	// Extract last path component
	parts := strings.Split(gitURL, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return "unknown-template"
}

// cloneRepository clones a git repository to the specified path
func (i *Installer) cloneRepository(gitURL, destPath string) error {
	// Check if git is installed
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git is not installed. Please install git and try again")
	}

	// Clone with progress
	cmd := exec.Command("git", "clone", "--depth", "1", gitURL, destPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	// Remove .git directory to save space
	gitDir := filepath.Join(destPath, ".git")
	os.RemoveAll(gitDir)

	return nil
}

// validateTemplate validates the template structure and required files
func (i *Installer) validateTemplate(template *Template) error {
	// Check for required template.yaml
	templateYAML := filepath.Join(template.Path, "template.yaml")
	if _, err := os.Stat(templateYAML); os.IsNotExist(err) {
		return fmt.Errorf("missing required file: template.yaml")
	}

	// Check for src directory
	srcDir := filepath.Join(template.Path, "src")
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("missing required directory: src/")
	}

	// Validate template name
	if template.Name == "" {
		return fmt.Errorf("template.yaml must specify a 'name' field")
	}

	// Validate language
	validLanguages := []string{"nodejs-express", "nodejs-fastify", "go", "python-fastapi", "go-graphql", "nodejs-apollo", "nodejs-yoga"}
	langValid := false
	for _, lang := range validLanguages {
		if template.Language == lang {
			langValid = true
			break
		}
	}
	if !langValid {
		return fmt.Errorf("unsupported language '%s'. Valid options: %s", template.Language, strings.Join(validLanguages, ", "))
	}

	return nil
}

// VerifyIntegrity verifies the integrity of an installed template
func (i *Installer) VerifyIntegrity(template *Template) error {
	// Check if all required files exist
	requiredFiles := []string{"template.yaml", "src"}

	for _, file := range requiredFiles {
		path := filepath.Join(template.Path, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("missing required file/directory: %s", file)
		}
	}

	// Check if README exists (recommended but not required)
	readmePath := filepath.Join(template.Path, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		fmt.Printf("⚠️  Warning: No README.md found in template '%s'\n", template.Name)
	}

	return nil
}
