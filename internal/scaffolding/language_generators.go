package scaffolding

import (
	"fmt"
	"path/filepath"
)

// LanguageType represents the programming language/runtime
type LanguageType string

const (
	LanguageGo            LanguageType = "go"
	LanguageNodeExpress   LanguageType = "nodejs-express"
	LanguageNodeFastify   LanguageType = "nodejs-fastify"
	LanguagePythonFastAPI LanguageType = "python-fastapi"
	LanguageNodeApollo    LanguageType = "nodejs-apollo" // GraphQL Apollo Server
	LanguageGoGqlgen      LanguageType = "go-graphql"    // GraphQL gqlgen
)

// LanguageConfig holds configuration for different languages
type LanguageConfig struct {
	Type         LanguageType
	Name         string
	Framework    string
	DepsPath     string
	MainFile     string
	TemplatePath string
}

// GetLanguageConfig returns configuration for a language type
func GetLanguageConfig(langType LanguageType) LanguageConfig {
	switch langType {
	case LanguageGo:
		return LanguageConfig{
			Type:         LanguageGo,
			Name:         "Go (Chi)",
			Framework:    "chi", // default framework
			DepsPath:     "go.mod",
			MainFile:     "main.go",
			TemplatePath: "templates/go",
		}
	case LanguageNodeExpress:
		return LanguageConfig{
			Type:         LanguageNodeExpress,
			Name:         "Node.js (Express)",
			Framework:    "express",
			DepsPath:     "package.json",
			MainFile:     "index.js",
			TemplatePath: "node/express",
		}
	case LanguageNodeFastify:
		return LanguageConfig{
			Type:         LanguageNodeFastify,
			Name:         "Node.js (Fastify)",
			Framework:    "fastify",
			DepsPath:     "package.json",
			MainFile:     "index.js",
			TemplatePath: "node/fastify",
		}
	case LanguagePythonFastAPI:
		return LanguageConfig{
			Type:         LanguagePythonFastAPI,
			Name:         "Python (FastAPI)",
			Framework:    "fastapi",
			DepsPath:     "requirements.txt",
			MainFile:     "main.py",
			TemplatePath: "templates/python-fastapi",
		}
	case LanguageNodeApollo:
		return LanguageConfig{
			Type:         LanguageNodeApollo,
			Name:         "Node.js (Apollo GraphQL)",
			Framework:    "apollo",
			DepsPath:     "package.json",
			MainFile:     "index.js",
			TemplatePath: "node/apollo",
		}
	case LanguageGoGqlgen:
		return LanguageConfig{
			Type:         LanguageGoGqlgen,
			Name:         "Go (gqlgen GraphQL)",
			Framework:    "gqlgen",
			DepsPath:     "go.mod",
			MainFile:     "server.go",
			TemplatePath: "templates/go-graphql",
		}
	default:
		return LanguageConfig{
			Type:         LanguageGo,
			Name:         "Go",
			Framework:    "chi",
			DepsPath:     "go.mod",
			MainFile:     "main.go",
			TemplatePath: "templates/go",
		}
	}
}

// GetSupportedLanguages returns all supported language types
func GetSupportedLanguages() []LanguageConfig {
	return []LanguageConfig{
		GetLanguageConfig(LanguageGo),
		GetLanguageConfig(LanguageNodeExpress),
		GetLanguageConfig(LanguageNodeFastify),
		GetLanguageConfig(LanguagePythonFastAPI),
		GetLanguageConfig(LanguageNodeApollo),
		GetLanguageConfig(LanguageGoGqlgen),
	}
}

// IsValidLanguage checks if a language type is supported
func IsValidLanguage(lang string) bool {
	langType := LanguageType(lang)
	for _, config := range GetSupportedLanguages() {
		if string(config.Type) == string(langType) {
			return true
		}
	}
	return false
}

// GetLanguageFromFramework maps framework names to language types
func GetLanguageFromFramework(framework string) LanguageType {
	switch framework {
	case "chi", "echo", "fiber":
		return LanguageGo
	case "express":
		return LanguageNodeExpress
	case "fastify":
		return LanguageNodeFastify
	case "apollo":
		return LanguageNodeApollo
	case "gqlgen":
		return LanguageGoGqlgen
	case "python-fastapi", "fastapi":
		return LanguagePythonFastAPI
	default:
		return LanguageGo
	}
}

// GenerateProjectFile creates project files based on language type
func (g *Generator) GenerateProjectFile(projectName string, langConfig LanguageConfig) error {
	switch langConfig.Type {
	case LanguageGo:
		return g.generateGoProject(projectName, langConfig)
	case LanguageNodeExpress:
		return g.generateNodeExpressProject(projectName, langConfig)
	case LanguageNodeFastify:
		return g.generateNodeFastifyProject(projectName, langConfig)
	case LanguagePythonFastAPI:
		return g.generatePythonFastAPIProject(projectName, langConfig)
	default:
		return fmt.Errorf("unsupported language type: %s", langConfig.Type)
	}
}

// generateGoProject creates Go-specific project structure
func (g *Generator) generateGoProject(projectName string, config LanguageConfig) error {
	// This will contain the Go project generation logic
	// Implementation will be moved from init.go
	return nil
}

// generateNodeExpressProject creates Node.js Express project structure
func (g *Generator) generateNodeExpressProject(projectName string, config LanguageConfig) error {
	// Create package.json
	packageJson := g.generatePackageContent(projectName, "express")
	return g.generateFile(filepath.Join(projectName, "package.json"), packageJson)
}

// generateNodeFastifyProject creates Node.js Fastify project structure
func (g *Generator) generateNodeFastifyProject(projectName string, config LanguageConfig) error {
	// Create package.json
	packageJson := g.generatePackageContent(projectName, "fastify")
	return g.generateFile(filepath.Join(projectName, "package.json"), packageJson)
}

// generatePythonFastAPIProject creates Python FastAPI project structure
func (g *Generator) generatePythonFastAPIProject(projectName string, config LanguageConfig) error {
	// Create requirements.txt
	requirements := g.generatePythonRequirements(projectName)
	return g.generateFile(filepath.Join(projectName, "requirements.txt"), requirements)
}

// generatePythonRequirements creates requirements.txt content for Python projects
// DEPRECATED: This function uses hardcoded versions. Use template files instead:
// internal/templates/python/common/requirements.txt.tmpl
// This function is kept for backward compatibility only.
func (g *Generator) generatePythonRequirements(projectName string) string {
	// TODO: Remove this function and use template loading instead
	// For now, return empty to force template usage
	return `# See requirements.txt.tmpl for dependencies
# This hardcoded version is deprecated
`
}

// generatePackageContent creates package.json content for Node.js projects
// DEPRECATED: This function uses hardcoded versions. Use template files instead:
// internal/templates/node/express/package.json.tmpl
// internal/templates/node/fastify/package.json.tmpl
// This function is kept for backward compatibility only.
func (g *Generator) generatePackageContent(projectName, framework string) string {
	// TODO: Remove this function and use template loading instead
	return fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "description": "Backend-for-Frontend service generated by bffgen - USE TEMPLATES",
  "main": "index.js",
  "engines": {
    "node": ">=18.0.0"
  }
}`, projectName)
}

// generateFile creates a file with given content
func (g *Generator) generateFile(filepath, content string) error {
	// Implementation will handle file creation
	return nil
}
