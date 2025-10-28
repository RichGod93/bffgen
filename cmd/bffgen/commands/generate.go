// File: generate.go
// Purpose: Core code generation command and orchestration
// Routes generation requests to language-specific handlers (Go or Node.js)

package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Go code for routes from config",
	Long:  `Generate Go code for routes from bff.config.yaml configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := generate(); err != nil {
			HandleError(err, "code generation")
		}
	},
}

var (
	checkMode bool
	dryRun    bool
	verbose   bool
	forceMode bool
)

func init() {
	generateCmd.Flags().BoolVar(&checkMode, "check", false, "Check mode: show what would be changed without making changes")
	generateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Dry run: show what would be changed without making changes")
	generateCmd.Flags().BoolVar(&verbose, "verbose", false, "Verbose output")
	generateCmd.Flags().BoolVar(&forceMode, "force", false, "Force overwrite of existing files without markers")
}

func generate() error {
	LogVerboseCommand("Starting code generation")

	// Detect project type
	projectType := detectProjectType()

	if projectType == "unknown" {
		LogInfo("No BFF project found in current directory")
		LogInfo("Run 'bffgen init <project-name>' first or navigate to a BFF project directory")
		return fmt.Errorf("no project configuration found")
	}

	// Handle based on project type
	if projectType == "nodejs" {
		return generateNodeJS()
	}

	// Default: Go project
	return generateGo()
}
