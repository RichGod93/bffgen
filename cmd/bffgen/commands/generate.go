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
	generateCmd.Flags().BoolVar(&checkMode, "check", false, "Show what would change and exit non-zero if generated code is out of date; writes nothing")
	generateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would change without writing anything")
	generateCmd.Flags().BoolVar(&verbose, "verbose", false, "Verbose output")
	generateCmd.Flags().BoolVar(&forceMode, "force", false, "Regenerate code you have edited (Node.js/Python: your versions are kept as <file>.bak; Go: regenerate existing routes)")
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

	if projectType == "python" {
		return generatePython()
	}

	// Default: Go project
	return generateGo()
}
