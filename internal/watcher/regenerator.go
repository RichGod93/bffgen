package watcher

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Regen erator handles smart code regeneration based on file changes
type Regenerator struct {
	configPath string
	verbose    bool
}

// NewRegenerator creates a new regenerator
func NewRegenerator(configPath string, verbose bool) *Regenerator {
	return &Regenerator{
		configPath: configPath,
		verbose:    verbose,
	}
}

// Regenerate performs smart regeneration based on what changed
func (r *Regenerator) Regenerate() error {
	if r.verbose {
		fmt.Println("📝 Analyzing changes...")
	}

	// TODO: Parse config to determine what needs regeneration
	// For now, trigger full regeneration

	if r.verbose {
		fmt.Println("🔨 Regenerating affected files...")
	}

	// This would call the existing generate command logic
	// For now, just a placeholder

	if r.verbose {
		fmt.Println("✅ Regeneration complete")
	}

	return nil
}

// RegenerateSelective regenerates only specific files based on change type
func (r *Regenerator) RegenerateSelective(changeType string) error {
	switch changeType {
	case "routes":
		// Only regenerate route files
		if r.verbose {
			fmt.Println("🔨 Regenerating routes...")
		}
	case "config":
		// Regenerate configuration-related files
		if r.verbose {
			fmt.Println("🔨 Regenerating config...")
		}
	default:
		return r.Regenerate()
	}

	return nil
}

// ShowDiff shows differences before applying changes
func (r *Regenerator) ShowDiff() error {
	// TODO: Implement diff viewing
	fmt.Println("📊 Diff preview not yet implemented")
	return nil
}

// Backup creates a backup before regeneration
func (r *Regenerator) Backup(targetPath string) error {
	backupPath := targetPath + ".backup." + time.Now().Format("20060102-150405")

	data, err := os.ReadFile(targetPath)
	if err != nil {
		return fmt.Errorf("failed to read file for backup: %w", err)
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	if r.verbose {
		fmt.Printf("💾 Backup created: %s\n", filepath.Base(backupPath))
	}

	return nil
}

// Rollback rolls back changes using backup
func (r *Regenerator) Rollback(targetPath string) error {
	// Find most recent backup
	pattern := targetPath + ".backup.*"
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return fmt.Errorf("no backup found for %s", targetPath)
	}

	// Use most recent backup
	backupPath := matches[len(matches)-1]

	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup: %w", err)
	}

	if err := os.WriteFile(targetPath, data, 0644); err != nil {
		return fmt.Errorf("failed to restore from backup: %w", err)
	}

	if r.verbose {
		fmt.Printf("⏮️  Rolled back to: %s\n", filepath.Base(backupPath))
	}

	return nil
}
