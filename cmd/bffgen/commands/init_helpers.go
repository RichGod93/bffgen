// File: init_helpers.go
// Purpose: Core project initialization and orchestration
// Routes initialization requests to language-specific handlers (Go or Node.js)

package commands

import (
	"os"
	"path/filepath"

	"github.com/RichGod93/bffgen/internal/scaffolding"
	"github.com/RichGod93/bffgen/internal/utils"
)

// createProjectDirectories creates directories based on language type
func createProjectDirectories(projectName string, langType scaffolding.LanguageType) error {
	var dirs []string

	if langType == scaffolding.LanguageGo {
		// Go-specific directories
		dirs = []string{
			filepath.Join(projectName, "internal", "routes"),
			filepath.Join(projectName, "internal", "aggregators"),
			filepath.Join(projectName, "internal", "templates"),
			filepath.Join(projectName, "cmd", "server"),
		}
	} else if langType == scaffolding.LanguagePythonFastAPI {
		// Python-specific directories
		return createPythonDirectories(projectName)
	} else {
		// Node.js-specific directories with src/ structure
		dirs = []string{
			filepath.Join(projectName, "src"),
			filepath.Join(projectName, "src", "routes"),
			filepath.Join(projectName, "src", "middleware"),
			filepath.Join(projectName, "src", "controllers"),
			filepath.Join(projectName, "src", "services"),
			filepath.Join(projectName, "src", "utils"),
			filepath.Join(projectName, "src", "config"),
			filepath.Join(projectName, "tests"),
			filepath.Join(projectName, "tests", "unit"),
			filepath.Join(projectName, "tests", "integration"),
		}
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, utils.ProjectDirPerm); err != nil {
			return err
		}
	}

	return nil
}

// createDependencyFiles creates language-specific dependency files
func createDependencyFiles(projectName string, langType scaffolding.LanguageType, framework string) error {
	return createDependencyFilesWithOptions(projectName, langType, framework, ProjectOptions{})
}

// createDependencyFilesWithOptions creates language-specific dependency files with options
func createDependencyFilesWithOptions(projectName string, langType scaffolding.LanguageType, framework string, opts ProjectOptions) error {
	switch langType {
	case scaffolding.LanguageGo:
		return createGoModFile(projectName, framework)
	case scaffolding.LanguageNodeExpress, scaffolding.LanguageNodeFastify:
		return createPackageJsonFile(projectName, langType, framework)
	case scaffolding.LanguagePythonFastAPI:
		return createPythonDependencyFiles(projectName, opts)
	default:
		return nil
	}
}

// createMainFile creates the main server file based on language/framework
func createMainFile(projectName string, langType scaffolding.LanguageType, framework string, corsConfig string, backendServs []BackendService) error {
	return createMainFileWithOptions(projectName, langType, framework, corsConfig, backendServs, ProjectOptions{})
}

// createMainFileWithOptions creates the main server file with options
func createMainFileWithOptions(projectName string, langType scaffolding.LanguageType, framework string, corsConfig string, backendServs []BackendService, opts ProjectOptions) error {
	switch langType {
	case scaffolding.LanguageGo:
		return createGoMainFile(projectName, framework, corsConfig)
	case scaffolding.LanguageNodeExpress:
		return createNodeExpressMainFileWithOptions(projectName, backendServs, opts)
	case scaffolding.LanguageNodeFastify:
		return createNodeFastifyMainFileWithOptions(projectName, backendServs, opts)
	case scaffolding.LanguagePythonFastAPI:
		return createFastAPIMainFile(projectName, opts)
	default:
		return nil
	}
}
