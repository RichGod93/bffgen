package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

// ToolInfo represents information about an installed tool
type ToolInfo struct {
	Name      string
	Command   string
	Installed bool
	Version   string
	Required  bool
}

// CheckTool checks if a tool is installed and gets its version
func CheckTool(name, command string) *ToolInfo {
	info := &ToolInfo{
		Name:    name,
		Command: command,
	}

	// Try to get version
	cmd := exec.Command(command, "--version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		info.Installed = false
		return info
	}

	info.Installed = true
	info.Version = strings.TrimSpace(string(output))

	// Clean up version string (first line only)
	lines := strings.Split(info.Version, "\n")
	if len(lines) > 0 {
		info.Version = strings.TrimSpace(lines[0])
	}

	return info
}

// CheckRequiredTools checks all required tools for a project type
func CheckRequiredTools(projectType string) ([]ToolInfo, []ToolInfo) {
	var required, optional []ToolInfo

	switch projectType {
	case "go":
		goTool := CheckTool("Go", "go")
		goTool.Required = true
		required = append(required, *goTool)

		optional = append(optional, *CheckTool("Docker", "docker"))
		optional = append(optional, *CheckTool("Make", "make"))

	case "nodejs", "nodejs-express", "nodejs-fastify":
		nodeTool := CheckTool("Node.js", "node")
		nodeTool.Required = true
		required = append(required, *nodeTool)

		npmTool := CheckTool("npm", "npm")
		npmTool.Required = true
		required = append(required, *npmTool)

		optional = append(optional, *CheckTool("Docker", "docker"))
		optional = append(optional, *CheckTool("npx", "npx"))

	default:
		// Unknown project type, no specific requirements
	}

	return required, optional
}

// GetToolInstallInstructions returns installation instructions for a tool
func GetToolInstallInstructions(toolName string) string {
	instructions := map[string]string{
		"Go": `Install Go:
   - macOS: brew install go
   - Linux: https://go.dev/doc/install
   - Windows: https://go.dev/dl/`,

		"Node.js": `Install Node.js:
   - macOS: brew install node
   - Linux: curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash - && sudo apt-get install -y nodejs
   - Windows: https://nodejs.org/en/download/
   - Or use nvm: https://github.com/nvm-sh/nvm`,

		"npm": `npm is included with Node.js. Install Node.js to get npm.`,

		"Docker": `Install Docker:
   - macOS: brew install --cask docker
   - Linux: https://docs.docker.com/engine/install/
   - Windows: https://docs.docker.com/desktop/install/windows-install/`,

		"Make": `Install Make:
   - macOS: xcode-select --install
   - Linux: sudo apt-get install build-essential
   - Windows: choco install make`,

		"npx": `npx is included with npm 5.2+. Update npm to get npx:
   npm install -g npm@latest`,
	}

	if instr, ok := instructions[toolName]; ok {
		return instr
	}

	return fmt.Sprintf("No installation instructions available for %s", toolName)
}

// PrintToolStatus prints the status of tools
func PrintToolStatus(required, optional []ToolInfo) {
	fmt.Println("🔧 Tool Check:")
	fmt.Println()

	allOk := true

	// Check required tools
	if len(required) > 0 {
		fmt.Println("Required:")
		for _, tool := range required {
			if tool.Installed {
				fmt.Printf("   ✅ %s: %s\n", tool.Name, tool.Version)
			} else {
				fmt.Printf("   ❌ %s: Not installed\n", tool.Name)
				allOk = false
			}
		}
		fmt.Println()
	}

	// Check optional tools
	if len(optional) > 0 {
		fmt.Println("Optional:")
		for _, tool := range optional {
			if tool.Installed {
				fmt.Printf("   ✅ %s: %s\n", tool.Name, tool.Version)
			} else {
				fmt.Printf("   ⚪ %s: Not installed (optional)\n", tool.Name)
			}
		}
		fmt.Println()
	}

	// Print installation instructions for missing required tools
	if !allOk {
		fmt.Println("⚠️  Missing required tools. Installation instructions:")
		fmt.Println()
		for _, tool := range required {
			if !tool.Installed {
				fmt.Println(GetToolInstallInstructions(tool.Name))
				fmt.Println()
			}
		}
	}
}

// HasRequiredTools checks if all required tools are installed
func HasRequiredTools(projectType string) bool {
	required, _ := CheckRequiredTools(projectType)

	for _, tool := range required {
		if !tool.Installed {
			return false
		}
	}

	return true
}
