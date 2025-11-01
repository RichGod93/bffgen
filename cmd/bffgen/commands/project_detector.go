package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/RichGod93/bffgen/internal/utils"
)

// detectProjectType detects if this is a Go or Node.js project
func detectProjectType() string {
	// Check for global runtime override first
	if globalConfig.RuntimeOverride != "" {
		runtime := normalizeRuntime(globalConfig.RuntimeOverride)
		if runtime != "unknown" {
			return runtime
		}
	}

	// Check for stored runtime in project metadata
	if storedRuntime := getStoredRuntime(); storedRuntime != "" {
		return storedRuntime
	}

	// Priority: Config files > runtime files

	// Check for bffgen.config.py.json (Python - highest priority)
	if _, err := os.Stat("bffgen.config.py.json"); err == nil {
		return "python"
	}

	// Check for bffgen.config.json (Node.js - highest priority)
	if _, err := os.Stat("bffgen.config.json"); err == nil {
		return "nodejs"
	}

	// Check for bff.config.yaml (Go - highest priority)
	if _, err := os.Stat("bff.config.yaml"); err == nil {
		return "go"
	}

	// Check for package.json (Node.js - lower priority)
	if _, err := os.Stat("package.json"); err == nil {
		return "nodejs"
	}

	// Check for go.mod (Go - lower priority)
	if _, err := os.Stat("go.mod"); err == nil {
		return "go"
	}

	// Check for requirements.txt or pyproject.toml (Python - lower priority)
	if _, err := os.Stat("requirements.txt"); err == nil {
		return "python"
	}
	if _, err := os.Stat("pyproject.toml"); err == nil {
		return "python"
	}

	return "unknown"
}

// getStoredRuntime retrieves stored runtime from project metadata
func getStoredRuntime() string {
	metadataPath := filepath.Join(utils.GetStateDir(), "metadata.json")

	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return ""
	}

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return ""
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		return ""
	}

	runtime, _ := metadata["runtime"].(string)
	return normalizeRuntime(runtime)
}

// normalizeRuntime normalizes runtime strings to standard format
func normalizeRuntime(runtime string) string {
	runtime = strings.ToLower(strings.TrimSpace(runtime))

	switch runtime {
	case "go", "golang":
		return "go"
	case "node", "nodejs", "nodejs-express", "express":
		return "nodejs"
	case "nodejs-fastify", "fastify":
		return "nodejs"
	case "python", "python-fastapi", "fastapi-python":
		return "python"
	default:
		return "unknown"
	}
}

