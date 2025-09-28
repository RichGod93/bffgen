package scaffolding

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Generator represents a code generator with regeneration-safe capabilities
type Generator struct {
	Markers     CodeMarker
	DryRun      bool
	CheckMode   bool
	BackupDir   string
	Verbose     bool
}

// NewGenerator creates a new generator with default settings
func NewGenerator() *Generator {
	return &Generator{
		Markers:   DefaultMarkers(),
		DryRun:    false,
		CheckMode: false,
		Verbose:   false,
	}
}

// SetMarkers sets custom markers for the generator
func (g *Generator) SetMarkers(marker CodeMarker) {
	g.Markers = marker
}

// SetDryRun enables or disables dry-run mode
func (g *Generator) SetDryRun(dryRun bool) {
	g.DryRun = dryRun
}

// SetCheckMode enables or disables check mode (dry-run with diffing)
func (g *Generator) SetCheckMode(checkMode bool) {
	g.CheckMode = checkMode
}

// SetBackupDir sets the backup directory for file operations
func (g *Generator) SetBackupDir(backupDir string) {
	g.BackupDir = backupDir
}

// SetVerbose enables or disables verbose output
func (g *Generator) SetVerbose(verbose bool) {
	g.Verbose = verbose
}

// GenerateFile generates or updates a file with regeneration-safe markers
func (g *Generator) GenerateFile(filePath, newContent string) error {
	if g.Verbose {
		fmt.Printf("Generating file: %s\n", filePath)
	}

	// Read existing content if file exists
	var existingContent string
	if _, err := os.Stat(filePath); err == nil {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read existing file: %w", err)
		}
		existingContent = string(content)
	}

	// Perform regeneration-safe update
	updatedContent, diff, err := g.updateContentSafely(existingContent, newContent)
	if err != nil {
		return fmt.Errorf("failed to update content safely: %w", err)
	}

	// Handle check mode
	if g.CheckMode {
		return g.handleCheckMode(filePath, existingContent, updatedContent, diff)
	}

	// Handle dry run
	if g.DryRun {
		return g.handleDryRun(filePath, existingContent, updatedContent, diff)
	}

	// Create backup if needed
	if g.BackupDir != "" && existingContent != "" {
		if err := g.createBackup(filePath, existingContent); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Write the updated content
	if err := g.writeFile(filePath, updatedContent); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if g.Verbose {
		fmt.Printf("✅ Updated %s: %s\n", filePath, diff.Summary)
	}

	return nil
}

// updateContentSafely updates content using regeneration-safe markers
func (g *Generator) updateContentSafely(existingContent, newContent string) (string, *DiffResult, error) {
	// If no existing content, return new content wrapped in markers
	if existingContent == "" {
		wrappedContent := g.wrapContentInMarkers(newContent)
		diff := ComputeDiff("", wrappedContent)
		return wrappedContent, diff, nil
	}

	// Find existing sections
	sections, err := FindSections(existingContent, g.Markers)
	if err != nil {
		return "", nil, fmt.Errorf("failed to find existing sections: %w", err)
	}

	// If no sections found, wrap new content and append
	if len(sections) == 0 {
		wrappedContent := existingContent + "\n\n" + g.wrapContentInMarkers(newContent)
		diff := ComputeDiff(existingContent, wrappedContent)
		return wrappedContent, diff, nil
	}

	// Update the first section with new content
	updatedContent, err := ReplaceSection(existingContent, sections[0], newContent)
	if err != nil {
		return "", nil, fmt.Errorf("failed to replace section: %w", err)
	}

	// Compute diff
	diff := ComputeDiff(existingContent, updatedContent)

	return updatedContent, diff, nil
}

// wrapContentInMarkers wraps content in bffgen markers
func (g *Generator) wrapContentInMarkers(content string) string {
	var result strings.Builder
	result.WriteString(g.Markers.Begin)
	result.WriteString("\n")
	if content != "" {
		result.WriteString(content)
		result.WriteString("\n")
	}
	result.WriteString(g.Markers.End)
	return result.String()
}

// handleCheckMode handles check mode (dry-run with diffing)
func (g *Generator) handleCheckMode(filePath, existingContent, updatedContent string, diff *DiffResult) error {
	fmt.Printf("🔍 Check mode: %s\n", filePath)
	
	if !diff.HasChanges {
		fmt.Printf("✅ No changes needed\n")
		return nil
	}

	fmt.Printf("📋 Changes detected: %s\n", diff.Summary)
	fmt.Printf("%s\n", diff.FormatDiff())
	
	return nil
}

// handleDryRun handles dry-run mode
func (g *Generator) handleDryRun(filePath, existingContent, updatedContent string, diff *DiffResult) error {
	fmt.Printf("🔍 Dry run: %s\n", filePath)
	
	if !diff.HasChanges {
		fmt.Printf("✅ No changes needed\n")
		return nil
	}

	fmt.Printf("📋 Would update: %s\n", diff.Summary)
	
	if g.Verbose {
		fmt.Printf("%s\n", diff.FormatDiff())
	}
	
	return nil
}

// createBackup creates a backup of the file
func (g *Generator) createBackup(filePath, content string) error {
	if g.BackupDir == "" {
		return nil
	}

	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(g.BackupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Generate backup filename
	baseName := filepath.Base(filePath)
	backupPath := filepath.Join(g.BackupDir, baseName+".backup")

	// Write backup
	if err := os.WriteFile(backupPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write backup: %w", err)
	}

	if g.Verbose {
		fmt.Printf("📁 Created backup: %s\n", backupPath)
	}

	return nil
}

// writeFile writes content to a file
func (g *Generator) writeFile(filePath, content string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// GenerateMultipleFiles generates multiple files with regeneration-safe markers
func (g *Generator) GenerateMultipleFiles(files map[string]string) error {
	var errors []string
	
	for filePath, content := range files {
		if err := g.GenerateFile(filePath, content); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", filePath, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to generate files:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

// ValidateFile validates that a file has proper marker structure
func (g *Generator) ValidateFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	return ValidateMarkers(string(content), g.Markers)
}

// GetFileSummary returns a summary of markers in a file
func (g *Generator) GetFileSummary(filePath string) ([]string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return GetMarkerSummary(string(content), g.Markers)
}

// CleanupBackups removes backup files
func (g *Generator) CleanupBackups() error {
	if g.BackupDir == "" {
		return nil
	}

	return os.RemoveAll(g.BackupDir)
}
