package commands

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/RichGod93/bffgen/internal/scaffolding"
	"github.com/RichGod93/bffgen/internal/templates"
	"github.com/RichGod93/bffgen/internal/types"
	"github.com/RichGod93/bffgen/internal/utils"
)

// InfraTemplateData holds data for infrastructure templates
type InfraTemplateData struct {
	ProjectName     string
	Port            int
	Framework       string
	Version         string
	IncludeDocker   bool
	BackendServices []BackendServiceData
}

// BackendServiceData holds backend service information for templates
type BackendServiceData struct {
	Name    string
	BaseURL string
	EnvName string
}

// generateCIWorkflow generates GitHub Actions CI/CD workflow
func generateCIWorkflow(projectName string, langType scaffolding.LanguageType, includeDocker bool) error {
	// Create .github/workflows directory
	workflowDir := filepath.Join(projectName, ".github", "workflows")
	if err := os.MkdirAll(workflowDir, utils.ProjectDirPerm); err != nil {
		return fmt.Errorf("failed to create workflows directory: %w", err)
	}

	// Determine which template to use
	var templatePath string
	if langType == scaffolding.LanguageGo {
		templatePath = "infra/ci/github-actions-go.yml.tmpl"
	} else {
		templatePath = "infra/ci/github-actions-node.yml.tmpl"
	}

	// Load and render template
	content, err := templates.TemplateFS.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read CI template: %w", err)
	}

	tmpl, err := template.New("ci").Parse(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse CI template: %w", err)
	}

	data := InfraTemplateData{
		ProjectName:   projectName,
		IncludeDocker: includeDocker,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute CI template: %w", err)
	}

	// Write workflow file
	workflowPath := filepath.Join(workflowDir, "ci.yml")
	if err := os.WriteFile(workflowPath, buf.Bytes(), utils.ProjectFilePerm); err != nil {
		return fmt.Errorf("failed to write CI workflow: %w", err)
	}

	return nil
}

// generateDockerfile generates production Dockerfile and .dockerignore
func generateDockerfile(projectName string, langType scaffolding.LanguageType, framework string, port int) error {
	// Determine which templates to use
	var dockerfilePath, dockerignorePath string
	if langType == scaffolding.LanguageGo {
		dockerfilePath = "infra/docker/Dockerfile.go.tmpl"
		dockerignorePath = "infra/docker/.dockerignore.go.tmpl"
	} else {
		dockerfilePath = "infra/docker/Dockerfile.node.tmpl"
		dockerignorePath = "infra/docker/.dockerignore.node.tmpl"
	}

	// Load and render Dockerfile
	dockerContent, err := templates.TemplateFS.ReadFile(dockerfilePath)
	if err != nil {
		return fmt.Errorf("failed to read Dockerfile template: %w", err)
	}

	dockerTmpl, err := template.New("dockerfile").Parse(string(dockerContent))
	if err != nil {
		return fmt.Errorf("failed to parse Dockerfile template: %w", err)
	}

	data := InfraTemplateData{
		ProjectName: projectName,
		Port:        port,
		Framework:   framework,
	}

	var dockerBuf bytes.Buffer
	if err := dockerTmpl.Execute(&dockerBuf, data); err != nil {
		return fmt.Errorf("failed to execute Dockerfile template: %w", err)
	}

	// Write Dockerfile
	dockerfileDest := filepath.Join(projectName, "Dockerfile")
	if err := os.WriteFile(dockerfileDest, dockerBuf.Bytes(), utils.ProjectFilePerm); err != nil {
		return fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	// Load and write .dockerignore
	dockerignoreContent, err := templates.TemplateFS.ReadFile(dockerignorePath)
	if err != nil {
		return fmt.Errorf("failed to read .dockerignore template: %w", err)
	}

	dockerignoreDest := filepath.Join(projectName, ".dockerignore")
	if err := os.WriteFile(dockerignoreDest, dockerignoreContent, utils.ProjectFilePerm); err != nil {
		return fmt.Errorf("failed to write .dockerignore: %w", err)
	}

	return nil
}

// generateHealthChecks generates enhanced health check endpoints
func generateHealthChecks(projectName string, langType scaffolding.LanguageType, framework string, backends []types.BackendService) error {
	if langType == scaffolding.LanguageGo {
		return generateGoHealthChecks(projectName, backends)
	}
	return generateNodeHealthChecks(projectName, framework, backends)
}

// generateGoHealthChecks generates Go health check package
func generateGoHealthChecks(projectName string, backends []types.BackendService) error {
	// Create health package directory
	healthDir := filepath.Join(projectName, "internal", "health")
	if err := os.MkdirAll(healthDir, utils.ProjectDirPerm); err != nil {
		return fmt.Errorf("failed to create health directory: %w", err)
	}

	// Load health template
	content, err := templates.TemplateFS.ReadFile("go/health/health.go.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read health template: %w", err)
	}

	// Write health.go file
	healthPath := filepath.Join(healthDir, "health.go")
	if err := os.WriteFile(healthPath, content, utils.ProjectFilePerm); err != nil {
		return fmt.Errorf("failed to write health.go: %w", err)
	}

	return nil
}

// generateNodeHealthChecks generates Node.js health check utility
func generateNodeHealthChecks(projectName string, framework string, backends []types.BackendService) error {
	// Health check utility already exists in node/common/health.js.tmpl
	// It's created during init, so we just ensure it exists
	utilsDir := filepath.Join(projectName, "src", "utils")
	if err := os.MkdirAll(utilsDir, utils.ProjectDirPerm); err != nil {
		return fmt.Errorf("failed to create utils directory: %w", err)
	}

	// Load health template
	content, err := templates.TemplateFS.ReadFile("node/common/health.js.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read health template: %w", err)
	}

	// Write health.js file
	healthPath := filepath.Join(utilsDir, "health.js")
	if err := os.WriteFile(healthPath, content, utils.ProjectFilePerm); err != nil {
		return fmt.Errorf("failed to write health.js: %w", err)
	}

	return nil
}

// generateGracefulShutdown generates graceful shutdown handlers
func generateGracefulShutdown(projectName string, langType scaffolding.LanguageType, framework string) error {
	if langType == scaffolding.LanguageGo {
		return generateGoGracefulShutdown(projectName)
	}
	return generateNodeGracefulShutdown(projectName, framework)
}

// generateGoGracefulShutdown generates Go graceful shutdown package
func generateGoGracefulShutdown(projectName string) error {
	// Create shutdown package directory
	shutdownDir := filepath.Join(projectName, "internal", "shutdown")
	if err := os.MkdirAll(shutdownDir, utils.ProjectDirPerm); err != nil {
		return fmt.Errorf("failed to create shutdown directory: %w", err)
	}

	// Load shutdown template
	content, err := templates.TemplateFS.ReadFile("go/shutdown/graceful.go.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read shutdown template: %w", err)
	}

	// Write graceful.go file
	shutdownPath := filepath.Join(shutdownDir, "graceful.go")
	if err := os.WriteFile(shutdownPath, content, utils.ProjectFilePerm); err != nil {
		return fmt.Errorf("failed to write graceful.go: %w", err)
	}

	return nil
}

// generateNodeGracefulShutdown generates Node.js graceful shutdown utility
func generateNodeGracefulShutdown(projectName string, framework string) error {
	// Create utils directory if it doesn't exist
	utilsDir := filepath.Join(projectName, "src", "utils")
	if err := os.MkdirAll(utilsDir, utils.ProjectDirPerm); err != nil {
		return fmt.Errorf("failed to create utils directory: %w", err)
	}

	// Load graceful shutdown template
	content, err := templates.TemplateFS.ReadFile("node/common/graceful-shutdown.js.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read graceful shutdown template: %w", err)
	}

	// Write graceful-shutdown.js file
	shutdownPath := filepath.Join(utilsDir, "graceful-shutdown.js")
	if err := os.WriteFile(shutdownPath, content, utils.ProjectFilePerm); err != nil {
		return fmt.Errorf("failed to write graceful-shutdown.js: %w", err)
	}

	return nil
}

// generateDockerCompose generates development docker-compose.yml
func generateDockerCompose(projectName string, langType scaffolding.LanguageType, backends []types.BackendService, port int) error {
	// Determine which template to use
	var templatePath string
	if langType == scaffolding.LanguageGo {
		templatePath = "infra/compose/docker-compose.dev.go.tmpl"
	} else {
		templatePath = "infra/compose/docker-compose.dev.node.tmpl"
	}

	// Load and render template
	content, err := templates.TemplateFS.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read docker-compose template: %w", err)
	}

	tmpl, err := template.New("compose").Parse(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse docker-compose template: %w", err)
	}

	// Prepare backend service data
	backendData := make([]BackendServiceData, 0, len(backends))
	for _, backend := range backends {
		envName := strings.ToUpper(strings.ReplaceAll(backend.Name, "-", "_"))
		backendData = append(backendData, BackendServiceData{
			Name:    backend.Name,
			BaseURL: backend.BaseURL,
			EnvName: envName,
		})
	}

	data := InfraTemplateData{
		ProjectName:     projectName,
		Port:            port,
		BackendServices: backendData,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute docker-compose template: %w", err)
	}

	// Write docker-compose.yml file
	composePath := filepath.Join(projectName, "docker-compose.yml")
	if err := os.WriteFile(composePath, buf.Bytes(), utils.ProjectFilePerm); err != nil {
		return fmt.Errorf("failed to write docker-compose.yml: %w", err)
	}

	return nil
}
