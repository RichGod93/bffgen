package main

import (
	"fmt"
	"os"

	"github.com/RichGod93/bffgen/cmd/bffgen/commands"
)

// Version information - set during build
var (
	version   = "dev"
	buildTime = "unknown"
	commit    = "unknown"
)

func main() {
	if err := commands.Execute(version, buildTime, commit); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
