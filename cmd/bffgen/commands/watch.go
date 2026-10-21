package commands

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/RichGod93/bffgen/v3/internal/watcher"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch for config changes and auto-regenerate code",
	Long: `Start development mode with automatic code regeneration.
	
Watches configuration files (bff.config.yaml, bffgen.config.json) for changes
and automatically regenerates affected code. Perfect for rapid development.`,
	RunE: runWatch,
}

var (
	watchVerbose    bool
	watchNoRestart  bool
	watchNoDiff     bool
	watchConfigPath string
)

func init() {
	watchCmd.Flags().BoolVarP(&watchVerbose, "verbose", "v", false, "Show detailed logs")
	watchCmd.Flags().BoolVar(&watchNoRestart, "no-restart", false, "Skip server restart")
	watchCmd.Flags().BoolVar(&watchNoDiff, "no-diff", false, "Skip diff display")
	watchCmd.Flags().StringVar(&watchConfigPath, "config", "", "Custom config file path")
}

func runWatch(cmd *cobra.Command, args []string) error {
	// Determine config files to watch
	configPaths := []string{
		"bff.config.yaml",
		"bffgen.config.json",
	}

	if watchConfigPath != "" {
		configPaths = []string{watchConfigPath}
	}

	// Filter to existing files
	existingConfigs := []string{}
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			existingConfigs = append(existingConfigs, absPath)
		}
	}

	if len(existingConfigs) == 0 {
		return fmt.Errorf("no config files found to watch. Run 'bffgen init' first")
	}

	fmt.Println("🔍 BFFGen Watch Mode")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📁 Monitoring: %v\n\n", existingConfigs)

	// Create watcher
	configWatcher, err := watcher.NewConfigWatcher(existingConfigs, func(path string) error {
		return handleConfigChange(path)
	})
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer func() { _ = configWatcher.Stop() }()

	// Start watching
	if err := configWatcher.Start(); err != nil {
		return fmt.Errorf("failed to start watcher: %w", err)
	}

	fmt.Println("⏳ Watching for changes... (Press Ctrl+C to stop)")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n\n👋 Stopping watch mode...")
	return nil
}

func handleConfigChange(path string) error {
	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("\n[%s] 🔄 Config changed: %s\n", timestamp, filepath.Base(path))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Create regenerator
	regenerator := watcher.NewRegenerator(path, watchVerbose)

	// Show diff if not disabled
	if !watchNoDiff {
		if err := regenerator.ShowDiff(); err != nil {
			fmt.Printf("⚠️  Could not show diff: %v\n", err)
		}
	}

	// Perform regeneration
	if err := regenerator.Regenerate(); err != nil {
		fmt.Printf("❌ Regeneration failed: %v\n", err)
		return err
	}

	fmt.Println("\n✅ Processing complete")
	fmt.Println("⏳ Watching for changes...")

	return nil
}
