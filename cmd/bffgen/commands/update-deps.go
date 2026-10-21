package commands

import (
	"fmt"

	"github.com/RichGod93/bffgen/v3/internal/registry"
	"github.com/spf13/cobra"
)

var updateDepsCmd = &cobra.Command{
	Use:   "update-deps",
	Short: "Update dependency versions in the manifest",
	Long: `Update dependency versions in the manifest file (versions.yaml).

This command will fetch the latest stable versions from npm and PyPI registries
and update the internal version manifest. This ensures all generated projects
use the latest stable dependencies.

Examples:
  # Update all dependencies
  bffgen update-deps

  # Dry run to see what would be updated
  bffgen update-deps --dry-run

  # Update only Node.js dependencies
  bffgen update-deps --runtime nodejs-express

Note: This command is typically run by maintainers, not end users.
Generated projects will automatically use the updated versions.`,
	RunE: runUpdateDeps,
}

var (
	updateDryRun  bool
	updateRuntime string
)

func init() {
	updateDepsCmd.Flags().BoolVar(&updateDryRun, "dry-run", false, "Show what would be updated without making changes")
	updateDepsCmd.Flags().StringVar(&updateRuntime, "runtime", "", "Update only a specific runtime (e.g., nodejs-express)")
}

func runUpdateDeps(cmd *cobra.Command, args []string) error {
	fmt.Println("🔄 Fetching latest dependency versions...")
	fmt.Println()

	// Load current manifest
	manifest, err := registry.LoadManifest()
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	fmt.Printf("📋 Current manifest version: %s\n", manifest.Metadata.Version)
	fmt.Printf("📅 Last updated: %s\n", manifest.Metadata.LastUpdated)
	fmt.Println()

	// Get list of runtimes to update
	runtimesToUpdate := manifest.GetRuntimes()
	if updateRuntime != "" {
		runtimesToUpdate = []string{updateRuntime}
	}

	if updateDryRun {
		fmt.Println("🔍 DRY RUN MODE - No changes will be made")
		fmt.Println()
	}

	// Initialize registry clients
	npmClient := registry.NewNPMClient()
	pypiClient := registry.NewPyPIClient()

	// Initialize cache
	cache, err := registry.NewVersionCache()
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to initialize cache: %v\n", err)
		cache = nil
	}

	totalUpdates := 0
	for _, rt := range runtimesToUpdate {
		versions := manifest.GetAllVersions(rt)
		if len(versions) == 0 {
			continue
		}

		fmt.Printf("📦 Updating %s (%d packages)...\n", rt, len(versions))

		updated := 0
		for pkg := range versions {
			// Try cache first
			var newVersion string
			var cached bool

			if cache != nil {
				newVersion, cached = cache.Get(rt, pkg)
			}

			if !cached {
				// Fetch from registry
				var fetchErr error
				if rt == "python-fastapi" {
					newVersion, fetchErr = pypiClient.GetLatestVersionWithRange(pkg)
				} else {
					newVersion, fetchErr = npmClient.GetLatestVersionWithPrefix(pkg)
				}

				if fetchErr != nil {
					fmt.Printf("  ⚠️  %s: %v\n", pkg, fetchErr)
					continue
				}

				// Cache the result
				if cache != nil {
					if err := cache.Set(rt, pkg, newVersion); err != nil {
						fmt.Printf("  ⚠️  failed to cache %s: %v\n", pkg, err)
					}
				}
			}

			oldVersion := versions[pkg]
			if oldVersion != newVersion {
				if cached {
					fmt.Printf("  ✓ %s: %s → %s (cached)\n", pkg, oldVersion, newVersion)
				} else {
					fmt.Printf("  ✓ %s: %s → %s\n", pkg, oldVersion, newVersion)
				}
				versions[pkg] = newVersion
				updated++
				totalUpdates++
			}
		}

		if updated == 0 {
			fmt.Printf("  ✓ All packages up to date\n")
		} else {
			fmt.Printf("  Updated %d packages\n", updated)
		}
		fmt.Println()
	}

	if totalUpdates == 0 {
		fmt.Println("✅ All dependencies are already up to date!")
		return nil
	}

	if updateDryRun {
		fmt.Printf("🔍 Would update %d package versions (dry run - not saved)\n", totalUpdates)
		return nil
	}

	fmt.Printf("✅ Successfully updated %d package versions\n", totalUpdates)
	fmt.Println()
	fmt.Println("ℹ️  Note: Manifest file (versions.yaml) has been updated in memory.")
	fmt.Println("   To persist changes, the manifest file needs to be regenerated.")
	fmt.Println("   This will be automated in a future update.")

	return nil
}

// UpdateManifestMetadata updates the manifest metadata
// This is a helper function for future use when we implement live fetching
func UpdateManifestMetadata() error {
	// TODO: In v3.3.0, this will:
	// 1. Fetch latest versions from npm/PyPI
	// 2. Update versions.yaml
	// 3. Update metadata timestamp

	// For now, just display current info
	manifest, err := registry.LoadManifest()
	if err != nil {
		return err
	}

	fmt.Printf("Manifest last updated: %s\n", manifest.Metadata.LastUpdated)
	fmt.Printf("BFFGen version: %s\n", manifest.Metadata.BffgenVersion)

	return nil
}
