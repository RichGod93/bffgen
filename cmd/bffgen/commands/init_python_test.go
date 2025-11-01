package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreatePythonDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := filepath.Join(tmpDir, "test-python-project")

	err := createPythonDirectories(projectName)
	if err != nil {
		t.Fatalf("Failed to create Python directories: %v", err)
	}

	// Check main directories exist
	expectedDirs := []string{
		"routers",
		"services",
		"models",
		"middleware",
		"utils",
		"tests",
	}

	for _, dir := range expectedDirs {
		dirPath := filepath.Join(projectName, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			t.Errorf("Expected directory %s does not exist", dir)
		}
	}
}

func TestCreatePythonDependencyFiles(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := filepath.Join(tmpDir, "test-python-project")

	err := os.MkdirAll(projectName, 0755)
	if err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	opts := ProjectOptions{
		PkgManager:     "pip",
		AsyncEndpoints: true,
	}

	err = createPythonDependencyFiles(projectName, opts)
	if err != nil {
		t.Fatalf("Failed to create Python dependency files: %v", err)
	}

	// Check requirements.txt exists
	requirementsPath := filepath.Join(projectName, "requirements.txt")
	if _, err := os.Stat(requirementsPath); os.IsNotExist(err) {
		t.Error("requirements.txt does not exist")
	}

	// Verify content contains FastAPI
	content, err := os.ReadFile(requirementsPath)
	if err != nil {
		t.Fatalf("Failed to read requirements.txt: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "fastapi") {
		t.Error("requirements.txt should contain fastapi")
	}
	if !contains(contentStr, "uvicorn") {
		t.Error("requirements.txt should contain uvicorn")
	}
}

func TestCreateFastAPIMainFile(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := filepath.Join(tmpDir, "test-python-project")

	err := os.MkdirAll(projectName, 0755)
	if err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	opts := ProjectOptions{
		AsyncEndpoints: true,
	}

	err = createFastAPIMainFile(projectName, opts)
	if err != nil {
		t.Fatalf("Failed to create main.py: %v", err)
	}

	// Check main.py exists
	mainPath := filepath.Join(projectName, "main.py")
	if _, err := os.Stat(mainPath); os.IsNotExist(err) {
		t.Error("main.py does not exist")
	}

	// Verify content
	content, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatalf("Failed to read main.py: %v", err)
	}

	contentStr := string(content)
	expectedStrings := []string{
		"from fastapi import FastAPI",
		"app = FastAPI",
		"@app.get(\"/health\")",
		"BFFGEN:ROUTERS:START",
		"BFFGEN:ROUTERS:END",
	}

	for _, expected := range expectedStrings {
		if !contains(contentStr, expected) {
			t.Errorf("main.py should contain '%s'", expected)
		}
	}
}

func TestCreateFastAPIConfig(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := filepath.Join(tmpDir, "test-python-project")

	err := os.MkdirAll(projectName, 0755)
	if err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	err = createFastAPIConfig(projectName)
	if err != nil {
		t.Fatalf("Failed to create config.py: %v", err)
	}

	// Check config.py exists
	configPath := filepath.Join(projectName, "config.py")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config.py does not exist")
	}

	// Verify content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config.py: %v", err)
	}

	contentStr := string(content)
	expectedStrings := []string{
		"from pydantic_settings import BaseSettings",
		"class Settings(BaseSettings)",
		"settings = Settings()",
		"CORS_ORIGINS",
	}

	for _, expected := range expectedStrings {
		if !contains(contentStr, expected) {
			t.Errorf("config.py should contain '%s'", expected)
		}
	}
}

func TestCreateFastAPIDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := filepath.Join(tmpDir, "test-python-project")

	err := os.MkdirAll(projectName, 0755)
	if err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	err = createFastAPIDependencies(projectName)
	if err != nil {
		t.Fatalf("Failed to create dependencies.py: %v", err)
	}

	// Check dependencies.py exists
	depsPath := filepath.Join(projectName, "dependencies.py")
	if _, err := os.Stat(depsPath); os.IsNotExist(err) {
		t.Error("dependencies.py does not exist")
	}

	// Verify content
	content, err := os.ReadFile(depsPath)
	if err != nil {
		t.Fatalf("Failed to read dependencies.py: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "get_current_user") {
		t.Error("dependencies.py should contain get_current_user function")
	}
}

func TestCreatePythonEnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := filepath.Join(tmpDir, "test-python-project")

	err := os.MkdirAll(projectName, 0755)
	if err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	err = createPythonEnvFile(projectName)
	if err != nil {
		t.Fatalf("Failed to create .env: %v", err)
	}

	// Check .env exists
	envPath := filepath.Join(projectName, ".env")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		t.Error(".env does not exist")
	}

	// Verify content
	content, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("Failed to read .env: %v", err)
	}

	contentStr := string(content)
	expectedVars := []string{
		"PORT=",
		"DEBUG=",
		"CORS_ORIGINS=",
		"JWT_SECRET=",
	}

	for _, expected := range expectedVars {
		if !contains(contentStr, expected) {
			t.Errorf(".env should contain '%s'", expected)
		}
	}
}

func TestCreatePythonGitignore(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := filepath.Join(tmpDir, "test-python-project")

	err := os.MkdirAll(projectName, 0755)
	if err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	err = createPythonGitignore(projectName)
	if err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	// Check .gitignore exists
	gitignorePath := filepath.Join(projectName, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		t.Error(".gitignore does not exist")
	}

	// Verify Python-specific patterns
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("Failed to read .gitignore: %v", err)
	}

	contentStr := string(content)
	expectedPatterns := []string{
		"__pycache__",
		"*.py[cod]",
		"venv/",
		".pytest_cache",
		".env",
	}

	for _, expected := range expectedPatterns {
		if !contains(contentStr, expected) {
			t.Errorf(".gitignore should contain pattern '%s'", expected)
		}
	}
}

func TestCreatePythonLogger(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := filepath.Join(tmpDir, "test-python-project")

	err := os.MkdirAll(projectName, 0755)
	if err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	err = createPythonLogger(projectName)
	if err != nil {
		t.Fatalf("Failed to create logger.py: %v", err)
	}

	// Check logger.py exists
	loggerPath := filepath.Join(projectName, "logger.py")
	if _, err := os.Stat(loggerPath); os.IsNotExist(err) {
		t.Error("logger.py does not exist")
	}

	// Verify content
	content, err := os.ReadFile(loggerPath)
	if err != nil {
		t.Fatalf("Failed to read logger.py: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "setup_logger") {
		t.Error("logger.py should contain setup_logger function")
	}
}

func TestCreatePythonMiddleware(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := filepath.Join(tmpDir, "test-python-project")

	err := createPythonDirectories(projectName)
	if err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	err = createPythonMiddleware(projectName)
	if err != nil {
		t.Fatalf("Failed to create middleware: %v", err)
	}

	// Check middleware files exist
	middlewareDir := filepath.Join(projectName, "middleware")
	expectedFiles := []string{
		"__init__.py",
		"logging_middleware.py",
		"auth_middleware.py",
	}

	for _, file := range expectedFiles {
		filePath := filepath.Join(middlewareDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected middleware file %s does not exist", file)
		}
	}

	// Verify logging middleware content
	loggingPath := filepath.Join(middlewareDir, "logging_middleware.py")
	content, err := os.ReadFile(loggingPath)
	if err != nil {
		t.Fatalf("Failed to read logging_middleware.py: %v", err)
	}

	if !contains(string(content), "LoggingMiddleware") {
		t.Error("logging_middleware.py should contain LoggingMiddleware class")
	}
}

func TestCreatePythonTestFiles(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := filepath.Join(tmpDir, "test-python-project")

	err := createPythonDirectories(projectName)
	if err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	opts := ProjectOptions{
		SkipTests: false,
	}

	err = createPythonTestFiles(projectName, opts)
	if err != nil {
		t.Fatalf("Failed to create test files: %v", err)
	}

	// Check test files exist
	testsDir := filepath.Join(projectName, "tests")
	expectedFiles := []string{
		"__init__.py",
		"conftest.py",
		"pytest.ini",
	}

	for _, file := range expectedFiles {
		filePath := filepath.Join(testsDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected test file %s does not exist", file)
		}
	}
}

func TestCreatePythonBFFGenConfig(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := filepath.Join(tmpDir, "test-python-project")

	err := os.MkdirAll(projectName, 0755)
	if err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	opts := ProjectOptions{
		AsyncEndpoints: true,
		PkgManager:     "pip",
	}

	err = createPythonBFFGenConfig(projectName, opts)
	if err != nil {
		t.Fatalf("Failed to create bffgen config: %v", err)
	}

	// Check config exists
	configPath := filepath.Join(projectName, "bffgen.config.py.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("bffgen.config.py.json does not exist")
	}

	// Verify it's valid JSON
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, `"framework": "fastapi"`) {
		t.Error("Config should specify fastapi framework")
	}
	if !contains(contentStr, `"async": true`) {
		t.Error("Config should have async enabled")
	}
}

func TestCreatePythonSetupScript(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := filepath.Join(tmpDir, "test-python-project")

	err := os.MkdirAll(projectName, 0755)
	if err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	opts := ProjectOptions{
		PkgManager: "pip",
	}

	err = createPythonSetupScript(projectName, opts)
	if err != nil {
		t.Fatalf("Failed to create setup script: %v", err)
	}

	// Check setup.sh exists
	setupPath := filepath.Join(projectName, "setup.sh")
	if _, err := os.Stat(setupPath); os.IsNotExist(err) {
		t.Error("setup.sh does not exist")
	}

	// Verify content
	content, err := os.ReadFile(setupPath)
	if err != nil {
		t.Fatalf("Failed to read setup.sh: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "#!/bin/bash") {
		t.Error("setup.sh should have bash shebang")
	}
	if !contains(contentStr, "pip install") {
		t.Error("setup.sh should contain pip install command")
	}
	if !contains(contentStr, "venv") {
		t.Error("setup.sh should create virtual environment")
	}
}

func TestCreatePythonREADME(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := filepath.Join(tmpDir, "test-python-project")

	err := os.MkdirAll(projectName, 0755)
	if err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	opts := ProjectOptions{
		AsyncEndpoints: true,
		PkgManager:     "pip",
	}

	err = createPythonREADME(projectName, opts)
	if err != nil {
		t.Fatalf("Failed to create README: %v", err)
	}

	// Check README.md exists
	readmePath := filepath.Join(projectName, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		t.Error("README.md does not exist")
	}

	// Verify content
	content, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("Failed to read README.md: %v", err)
	}

	contentStr := string(content)
	expectedSections := []string{
		"# " + filepath.Base(projectName),
		"## Features",
		"## Quick Start",
		"./setup.sh",
		"uvicorn main:app",
	}

	for _, expected := range expectedSections {
		if !contains(contentStr, expected) {
			t.Errorf("README.md should contain '%s'", expected)
		}
	}
}

func TestPythonProjectInitializationOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    ProjectOptions
		wantErr bool
	}{
		{
			name: "pip with async",
			opts: ProjectOptions{
				PkgManager:     "pip",
				AsyncEndpoints: true,
			},
			wantErr: false,
		},
		{
			name: "poetry with sync",
			opts: ProjectOptions{
				PkgManager:     "poetry",
				AsyncEndpoints: false,
			},
			wantErr: false,
		},
		{
			name: "skip tests",
			opts: ProjectOptions{
				PkgManager: "pip",
				SkipTests:  true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			projectName := filepath.Join(tmpDir, "test-project")

			err := os.MkdirAll(projectName, 0755)
			if err != nil {
				t.Fatalf("Failed to create project directory: %v", err)
			}

			err = createPythonDependencyFiles(projectName, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("createPythonDependencyFiles() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPythonUtilityFiles(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := filepath.Join(tmpDir, "test-python-project")

	err := createPythonDirectories(projectName)
	if err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	// Test cache manager
	err = createPythonCacheManager(projectName)
	if err != nil {
		t.Fatalf("Failed to create cache manager: %v", err)
	}

	cachePath := filepath.Join(projectName, "utils", "cache_manager.py")
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Error("cache_manager.py does not exist")
	}

	// Test circuit breaker
	err = createPythonCircuitBreaker(projectName)
	if err != nil {
		t.Fatalf("Failed to create circuit breaker: %v", err)
	}

	cbPath := filepath.Join(projectName, "utils", "circuit_breaker.py")
	if _, err := os.Stat(cbPath); os.IsNotExist(err) {
		t.Error("circuit_breaker.py does not exist")
	}

	// Verify cache manager content
	content, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("Failed to read cache_manager.py: %v", err)
	}

	if !contains(string(content), "CacheManager") {
		t.Error("cache_manager.py should contain CacheManager class")
	}

	// Verify circuit breaker content
	content, err = os.ReadFile(cbPath)
	if err != nil {
		t.Fatalf("Failed to read circuit_breaker.py: %v", err)
	}

	if !contains(string(content), "CircuitBreaker") {
		t.Error("circuit_breaker.py should contain CircuitBreaker class")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
