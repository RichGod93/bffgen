// File: init_python.go
// Purpose: Python-specific project initialization
// Contains all logic for scaffolding FastAPI BFF projects

package commands

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/RichGod93/bffgen/internal/templates"
	"github.com/RichGod93/bffgen/internal/utils"
)

// createPythonDependencyFiles creates dependency files for Python projects
func createPythonDependencyFiles(projectName string, opts ProjectOptions) error {
	pkgManager := opts.PkgManager
	if pkgManager == "" {
		pkgManager = "pip" // default
	}

	if pkgManager == "poetry" {
		return createPyprojectToml(projectName, opts)
	}
	return createRequirementsTxt(projectName, opts)
}

// createRequirementsTxt creates requirements.txt for pip
func createRequirementsTxt(projectName string, opts ProjectOptions) error {
	content, err := templates.TemplateFS.ReadFile("python/common/requirements.txt.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read requirements template: %w", err)
	}

	filePath := filepath.Join(projectName, "requirements.txt")
	return os.WriteFile(filePath, content, utils.ProjectFilePerm)
}

// createPyprojectToml creates pyproject.toml for Poetry
func createPyprojectToml(projectName string, opts ProjectOptions) error {
	tmplContent, err := templates.TemplateFS.ReadFile("python/common/pyproject.toml.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read pyproject template: %w", err)
	}

	tmpl, err := template.New("pyproject").Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse pyproject template: %w", err)
	}

	data := map[string]interface{}{
		"ProjectName": projectName,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute pyproject template: %w", err)
	}

	filePath := filepath.Join(projectName, "pyproject.toml")
	return os.WriteFile(filePath, buf.Bytes(), utils.ProjectFilePerm)
}

// createFastAPIMainFile creates main.py for FastAPI projects
func createFastAPIMainFile(projectName string, opts ProjectOptions) error {
	tmplContent, err := templates.TemplateFS.ReadFile("python/fastapi/main.py.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read main.py template: %w", err)
	}

	tmpl, err := template.New("main").Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse main.py template: %w", err)
	}

	data := map[string]interface{}{
		"ProjectName": projectName,
		"Async":       opts.AsyncEndpoints,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute main.py template: %w", err)
	}

	filePath := filepath.Join(projectName, "main.py")
	return os.WriteFile(filePath, buf.Bytes(), utils.ProjectFilePerm)
}

// createFastAPIConfig creates config.py for FastAPI projects
func createFastAPIConfig(projectName string) error {
	content, err := templates.TemplateFS.ReadFile("python/fastapi/config.py.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read config.py template: %w", err)
	}

	filePath := filepath.Join(projectName, "config.py")
	return os.WriteFile(filePath, content, utils.ProjectFilePerm)
}

// createFastAPIDependencies creates dependencies.py for FastAPI projects
func createFastAPIDependencies(projectName string) error {
	content, err := templates.TemplateFS.ReadFile("python/fastapi/dependencies.py.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read dependencies.py template: %w", err)
	}

	filePath := filepath.Join(projectName, "dependencies.py")
	return os.WriteFile(filePath, content, utils.ProjectFilePerm)
}

// createPythonEnvFile creates .env file for Python projects
func createPythonEnvFile(projectName string) error {
	content, err := templates.TemplateFS.ReadFile("python/common/env.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read .env template: %w", err)
	}

	filePath := filepath.Join(projectName, ".env")
	return os.WriteFile(filePath, content, utils.ProjectFilePerm)
}

// createPythonGitignore creates .gitignore file for Python projects
func createPythonGitignore(projectName string) error {
	content, err := templates.TemplateFS.ReadFile("python/common/gitignore.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read gitignore template: %w", err)
	}

	filePath := filepath.Join(projectName, ".gitignore")
	return os.WriteFile(filePath, content, utils.ProjectFilePerm)
}

// createPythonLogger creates logger utility for Python projects
func createPythonLogger(projectName string) error {
	content, err := templates.TemplateFS.ReadFile("python/common/logger.py.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read logger template: %w", err)
	}

	// Create utils directory
	utilsDir := filepath.Join(projectName, "utils")
	if err := os.MkdirAll(utilsDir, utils.ProjectDirPerm); err != nil {
		return fmt.Errorf("failed to create utils directory: %w", err)
	}

	// Create __init__.py
	initPath := filepath.Join(utilsDir, "__init__.py")
	if err := os.WriteFile(initPath, []byte(""), utils.ProjectFilePerm); err != nil {
		return fmt.Errorf("failed to create utils/__init__.py: %w", err)
	}

	filePath := filepath.Join(utilsDir, "logger.py")
	return os.WriteFile(filePath, content, utils.ProjectFilePerm)
}

// createPythonCacheManager creates cache manager utility
func createPythonCacheManager(projectName string) error {
	content, err := templates.TemplateFS.ReadFile("python/common/cache_manager.py.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read cache_manager template: %w", err)
	}

	utilsDir := filepath.Join(projectName, "utils")
	filePath := filepath.Join(utilsDir, "cache_manager.py")
	return os.WriteFile(filePath, content, utils.ProjectFilePerm)
}

// createPythonCircuitBreaker creates circuit breaker utility
func createPythonCircuitBreaker(projectName string) error {
	content, err := templates.TemplateFS.ReadFile("python/common/circuit_breaker.py.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read circuit_breaker template: %w", err)
	}

	utilsDir := filepath.Join(projectName, "utils")
	filePath := filepath.Join(utilsDir, "circuit_breaker.py")
	return os.WriteFile(filePath, content, utils.ProjectFilePerm)
}

// createPythonMiddleware creates middleware files
func createPythonMiddleware(projectName string) error {
	middlewareDir := filepath.Join(projectName, "middleware")
	if err := os.MkdirAll(middlewareDir, utils.ProjectDirPerm); err != nil {
		return fmt.Errorf("failed to create middleware directory: %w", err)
	}

	// Create __init__.py
	initPath := filepath.Join(middlewareDir, "__init__.py")
	if err := os.WriteFile(initPath, []byte(""), utils.ProjectFilePerm); err != nil {
		return fmt.Errorf("failed to create middleware/__init__.py: %w", err)
	}

	// Create auth middleware
	authContent, err := templates.TemplateFS.ReadFile("python/fastapi/middleware_auth.py.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read auth middleware template: %w", err)
	}
	authPath := filepath.Join(middlewareDir, "auth.py")
	if err := os.WriteFile(authPath, authContent, utils.ProjectFilePerm); err != nil {
		return fmt.Errorf("failed to create auth middleware: %w", err)
	}

	// Create logging middleware
	loggingContent, err := templates.TemplateFS.ReadFile("python/fastapi/middleware_logging.py.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read logging middleware template: %w", err)
	}
	loggingPath := filepath.Join(middlewareDir, "logging.py")
	if err := os.WriteFile(loggingPath, loggingContent, utils.ProjectFilePerm); err != nil {
		return fmt.Errorf("failed to create logging middleware: %w", err)
	}

	return nil
}

// createPythonDirectories creates standard Python project directories
func createPythonDirectories(projectName string) error {
	dirs := []string{
		filepath.Join(projectName, "routers"),
		filepath.Join(projectName, "services"),
		filepath.Join(projectName, "models"),
		filepath.Join(projectName, "tests"),
		filepath.Join(projectName, "utils"),
		filepath.Join(projectName, "middleware"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, utils.ProjectDirPerm); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		// Create __init__.py in each directory
		initPath := filepath.Join(dir, "__init__.py")
		if err := os.WriteFile(initPath, []byte(""), utils.ProjectFilePerm); err != nil {
			return fmt.Errorf("failed to create %s: %w", initPath, err)
		}
	}

	return nil
}

// createPythonTestFiles creates test configuration and sample tests
func createPythonTestFiles(projectName string, opts ProjectOptions) error {
	if opts.SkipTests {
		return nil
	}

	// Create pytest.ini
	pytestContent, err := templates.TemplateFS.ReadFile("python/tests/pytest.ini.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read pytest.ini template: %w", err)
	}
	pytestPath := filepath.Join(projectName, "pytest.ini")
	if err := os.WriteFile(pytestPath, pytestContent, utils.ProjectFilePerm); err != nil {
		return fmt.Errorf("failed to create pytest.ini: %w", err)
	}

	testsDir := filepath.Join(projectName, "tests")

	// Create conftest.py
	conftestContent, err := templates.TemplateFS.ReadFile("python/tests/conftest.py.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read conftest.py template: %w", err)
	}
	conftestPath := filepath.Join(testsDir, "conftest.py")
	if err := os.WriteFile(conftestPath, conftestContent, utils.ProjectFilePerm); err != nil {
		return fmt.Errorf("failed to create conftest.py: %w", err)
	}

	// Create test_router.py
	testRouterContent, err := templates.TemplateFS.ReadFile("python/tests/test_router.py.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read test_router.py template: %w", err)
	}
	testRouterPath := filepath.Join(testsDir, "test_main.py")
	if err := os.WriteFile(testRouterPath, testRouterContent, utils.ProjectFilePerm); err != nil {
		return fmt.Errorf("failed to create test_main.py: %w", err)
	}

	return nil
}

// createPythonBFFGenConfig creates bffgen.config.py.json
func createPythonBFFGenConfig(projectName string, opts ProjectOptions) error {
	tmplContent, err := templates.TemplateFS.ReadFile("python/common/bffgen.config.py.json.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read bffgen config template: %w", err)
	}

	tmpl, err := template.New("config").Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse bffgen config template: %w", err)
	}

	data := map[string]interface{}{
		"ProjectName": projectName,
		"Async":       opts.AsyncEndpoints,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute bffgen config template: %w", err)
	}

	filePath := filepath.Join(projectName, "bffgen.config.py.json")
	return os.WriteFile(filePath, buf.Bytes(), utils.ProjectFilePerm)
}

// createPythonSetupScript creates setup.sh script
func createPythonSetupScript(projectName string, opts ProjectOptions) error {
	pkgManager := opts.PkgManager
	if pkgManager == "" {
		pkgManager = "pip"
	}

	var installCmd string
	if pkgManager == "poetry" {
		installCmd = `# Install dependencies with Poetry
poetry install

echo "✅ Dependencies installed successfully"
echo "📝 Activate environment with: poetry shell"
echo "🚀 Run with: poetry run uvicorn main:app --reload"`
	} else {
		installCmd = `# Install dependencies with pip
pip install -r requirements.txt

echo "✅ Dependencies installed successfully"
echo "🚀 Run with: uvicorn main:app --reload"`
	}

	script := `#!/bin/bash
set -e

echo "🔧 Setting up Python BFF project: ` + projectName + `"

# Check Python version
python_version=$(python3 --version 2>&1 | awk '{print $2}')
required_version="3.9"

if ! python3 -c "import sys; exit(0 if sys.version_info >= (3, 9) else 1)"; then
    echo "❌ Error: Python 3.9+ required (found $python_version)"
    exit 1
fi

echo "✅ Python version: $python_version"

# Create virtual environment
if [ ! -d "venv" ]; then
    echo "📦 Creating virtual environment..."
    python3 -m venv venv
fi

# Activate virtual environment
echo "🔌 Activating virtual environment..."
source venv/bin/activate

` + installCmd + `

echo ""
echo "🎉 Setup complete!"
echo "📚 Next steps:"
echo "  1. source venv/bin/activate (or poetry shell)"
echo "  2. Edit .env file with your configuration"
echo "  3. uvicorn main:app --reload (or poetry run uvicorn main:app --reload)"
echo "  4. Visit http://localhost:8080/docs for API documentation"
`

	filePath := filepath.Join(projectName, "setup.sh")
	if err := os.WriteFile(filePath, []byte(script), 0755); err != nil {
		return fmt.Errorf("failed to create setup.sh: %w", err)
	}

	return nil
}

// createPythonREADME creates README.md for Python projects
func createPythonREADME(projectName string, opts ProjectOptions) error {
	pkgManager := opts.PkgManager
	if pkgManager == "" {
		pkgManager = "pip"
	}

	installInstructions := ""
	if pkgManager == "poetry" {
		installInstructions = "## Setup\n\n" +
			"1. Install dependencies:\n" +
			"   ```bash\n" +
			"   poetry install\n" +
			"   ```\n\n" +
			"2. Activate environment:\n" +
			"   ```bash\n" +
			"   poetry shell\n" +
			"   ```\n\n" +
			"3. Run the development server:\n" +
			"   ```bash\n" +
			"   poetry run uvicorn main:app --reload\n" +
			"   ```"
	} else {
		installInstructions = "## Setup\n\n" +
			"1. Create and activate virtual environment:\n" +
			"   ```bash\n" +
			"   python -m venv venv\n" +
			"   source venv/bin/activate  # On Windows: venv\\Scripts\\activate\n" +
			"   ```\n\n" +
			"2. Install dependencies:\n" +
			"   ```bash\n" +
			"   pip install -r requirements.txt\n" +
			"   ```\n\n" +
			"3. Run the development server:\n" +
			"   ```bash\n" +
			"   uvicorn main:app --reload\n" +
			"   ```"
	}

	asyncInfo := ""
	if opts.AsyncEndpoints {
		asyncInfo = "\n## Async/Await\n\n" +
			"This project uses async/await for all endpoint handlers, providing better performance for I/O-bound operations."
	}

	readme := "# " + projectName + "\n\n" +
		"Backend-for-Frontend service generated by [bffgen](https://github.com/RichGod93/bffgen).\n\n" +
		"## Description\n\n" +
		"This is a FastAPI-based BFF service that aggregates and transforms backend APIs for frontend consumption.\n\n" +
		installInstructions + "\n\n" +
		"## API Documentation\n\n" +
		"Once running, visit:\n" +
		"- Swagger UI: http://localhost:8080/docs\n" +
		"- ReDoc: http://localhost:8080/redoc\n\n" +
		"## Project Structure\n\n" +
		"```\n" +
		projectName + "/\n" +
		"├── main.py              # FastAPI application entry point\n" +
		"├── config.py            # Configuration management\n" +
		"├── dependencies.py      # FastAPI dependencies\n" +
		"├── routers/             # API route handlers\n" +
		"├── services/            # Business logic and external API calls\n" +
		"├── models/              # Pydantic models\n" +
		"├── middleware/          # Custom middleware\n" +
		"├── utils/               # Utility functions\n" +
		"├── tests/               # Test files\n" +
		"├── .env                 # Environment variables\n" +
		"├── requirements.txt     # Python dependencies (pip)\n" +
		"└── pyproject.toml       # Python dependencies (Poetry)\n" +
		"```" +
		asyncInfo + "\n\n" +
		"## Configuration\n\n" +
		"Edit the `.env` file to configure:\n" +
		"- Server settings (PORT, HOST)\n" +
		"- CORS origins\n" +
		"- JWT secrets\n" +
		"- Backend service URLs\n" +
		"- Redis configuration\n\n" +
		"## Testing\n\n" +
		"Run tests with pytest:\n" +
		"```bash\n" +
		"pytest\n" +
		"```\n\n" +
		"With coverage:\n" +
		"```bash\n" +
		"pytest --cov=. --cov-report=html\n" +
		"```\n\n" +
		"## Code Generation\n\n" +
		"Add new routes with bffgen:\n" +
		"```bash\n" +
		"bffgen add-route\n" +
		"bffgen generate\n" +
		"```\n\n" +
		"## License\n\nMIT\n"

	filePath := filepath.Join(projectName, "README.md")
	return os.WriteFile(filePath, []byte(readme), utils.ProjectFilePerm)
}

