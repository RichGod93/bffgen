package commands

import (
	"encoding/json"
	"fmt"
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

// storeRuntime stores runtime information in project metadata
func storeRuntime(runtime string) error {
	metadataDir := utils.GetStateDir()
	if err := os.MkdirAll(metadataDir, utils.ProjectDirPerm); err != nil {
		return err
	}

	metadataPath := filepath.Join(metadataDir, "metadata.json")

	metadata := map[string]interface{}{
		"runtime":   runtime,
		"createdAt": utils.GetCurrentTimestamp(),
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metadataPath, data, utils.ProjectFilePerm)
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
	default:
		return "unknown"
	}
}

// detectProjectTypeWithFeedback detects project type and provides feedback
func detectProjectTypeWithFeedback() string {
	projectType := detectProjectType()

	if globalConfig.RuntimeOverride != "" && projectType != "unknown" {
		// Check if override conflicts with detected type
		detectedWithoutOverride := detectProjectTypeWithoutOverride()
		if detectedWithoutOverride != "unknown" && detectedWithoutOverride != projectType {
			fmt.Printf("⚠️  Runtime override (%s) differs from detected type (%s)\n",
				globalConfig.RuntimeOverride, detectedWithoutOverride)
		}
	}

	return projectType
}

// detectProjectTypeWithoutOverride detects project type ignoring override
func detectProjectTypeWithoutOverride() string {
	// Temporarily clear override
	originalOverride := globalConfig.RuntimeOverride
	globalConfig.RuntimeOverride = ""

	projectType := detectProjectType()

	// Restore override
	globalConfig.RuntimeOverride = originalOverride

	return projectType
}
