package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/RichGod93/bffgen/v3/cmd/bffgen/commands"
)

// Version information - set during build
var (
	version   = "dev"
	buildTime = "unknown"
	commit    = "unknown"
)

func main() {
	if err := commands.Execute(resolveVersion(version, debug.ReadBuildInfo), buildTime, commit); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// resolveVersion returns the ldflags version, or for `go install …@vX.Y.Z` builds (no ldflags)
// the module version recorded by the Go toolchain.
func resolveVersion(ldflagsVersion string, buildInfo func() (*debug.BuildInfo, bool)) string {
	if ldflagsVersion != "dev" {
		return ldflagsVersion
	}
	if info, ok := buildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return ldflagsVersion
}
