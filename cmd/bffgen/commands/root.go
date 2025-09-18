package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bffgen",
	Short: "A CLI tool for generating Backend-for-Frontend (BFF) services",
	Long: `bffgen is a Go-based CLI tool that helps developers quickly scaffold 
Backend-for-Frontend (BFF) services. It enables teams to aggregate backend 
endpoints and expose them in a frontend-friendly way, with minimal setup.

Global Installation:
  macOS/Linux: sudo cp bffgen /usr/local/bin/
  Windows: Add the bffgen directory to your PATH
  Or use: go install github.com/RichGod93/bffgen/cmd/bffgen`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Add all subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addRouteCmd)
	rootCmd.AddCommand(addTemplateCmd)
	rootCmd.AddCommand(addAggregatorCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(devCmd)
	rootCmd.AddCommand(postmanCmd)
}
