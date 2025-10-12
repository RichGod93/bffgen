package commands

import "os"

// detectProjectType detects if this is a Go or Node.js project
func detectProjectType() string {
	// Check for bffgen.config.json (Node.js)
	if _, err := os.Stat("bffgen.config.json"); err == nil {
		return "nodejs"
	}

	// Check for package.json (Node.js)
	if _, err := os.Stat("package.json"); err == nil {
		return "nodejs"
	}

	// Check for bff.config.yaml (Go)
	if _, err := os.Stat("bff.config.yaml"); err == nil {
		return "go"
	}

	// Check for go.mod (Go)
	if _, err := os.Stat("go.mod"); err == nil {
		return "go"
	}

	return "unknown"
}
