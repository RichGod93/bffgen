package main

import (
	"fmt"
	"os"

	"github.com/RichGod93/bffgen/cmd/bffgen/commands"
)

// Version information - set during build
var (
	version  = "dev"
	buildTime = "unknown"
	commit    = "unknown"
)

func main() {
	// Add version command if --version flag is passed
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("bffgen version %s\n", version)
		fmt.Printf("Build time: %s\n", buildTime)
		fmt.Printf("Commit: %s\n", commit)
		os.Exit(0)
	}
	
	commands.Execute()
}