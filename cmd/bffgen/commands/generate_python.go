// File: generate_python.go
// Purpose: Python-specific code generation for FastAPI framework
// Contains all logic for generating FastAPI routers and services

package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/RichGod93/bffgen/internal/templates"
	"github.com/RichGod93/bffgen/internal/utils"
)

// generatePython generates code for Python projects
func generatePython() error {
	// Create progress tracker
	progress := utils.NewQuietProgress(verbose)

	if checkMode {
		fmt.Println("🔍 Check mode: Analyzing what would be changed")
	} else if dryRun {
		fmt.Println("🔍 Dry run: Showing what would be changed")
	} else {
		fmt.Println("🔧 Generating Python routes from bffgen.config.py.json")
	}
	fmt.Println()

	progress.Start("Loading configuration")

	// Check if config file exists
	if _, err := os.Stat("bffgen.config.py.json"); os.IsNotExist(err) {
		fmt.Println("❌ bffgen.config.py.json not found in current directory")
		fmt.Println("💡 Run 'bffgen init <project-name> --lang python-fastapi' first or navigate to a BFF project directory")
		return fmt.Errorf("config file not found")
	}

	// Load bffgen.config.py.json
	configData, err := os.ReadFile("bffgen.config.py.json")
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Get backends
	backends, ok := config["backends"].([]interface{})
	if !ok || len(backends) == 0 {
		LogWarning("No backends configured in bffgen.config.py.json")
		fmt.Println("💡 Add backends manually to bffgen.config.py.json or use 'bffgen add-template'")
		return nil
	}

	LogVerboseCommand("Found %d backends to generate", len(backends))

	// Detect async mode
	asyncMode := true
	if project, ok := config["project"].(map[string]interface{}); ok {
		if async, ok := project["async"].(bool); ok {
			asyncMode = async
		}
	}

	progress.Success("Configuration loaded")

	fmt.Printf("📝 Generating routes for FastAPI (async: %v)\n", asyncMode)
	progress.Start(fmt.Sprintf("Generating files for %d backends", len(backends)))

	// Generate router files, services for each backend
	routersGenerated := 0
	servicesGenerated := 0

	for _, backend := range backends {
		backendMap, ok := backend.(map[string]interface{})
		if !ok {
			continue
		}

		serviceName, _ := backendMap["name"].(string)
		if serviceName == "" {
			continue
		}

		// Skip backends with no endpoints
		endpoints, ok := backendMap["endpoints"].([]interface{})
		if !ok || len(endpoints) == 0 {
			LogVerboseCommand("Skipping %s (no endpoints defined)", serviceName)
			continue
		}

		baseURL, _ := backendMap["url"].(string)

		// Generate router file
		if err := generateFastAPIRouter(serviceName, endpoints, baseURL, asyncMode); err != nil {
			return fmt.Errorf("failed to generate router for %s: %w", serviceName, err)
		}
		routersGenerated++

		// Generate service file
		if err := generateFastAPIService(serviceName, endpoints, baseURL, asyncMode); err != nil {
			return fmt.Errorf("failed to generate service for %s: %w", serviceName, err)
		}
		servicesGenerated++

		LogVerboseCommand("Generated router and service for %s", serviceName)
	}

	progress.Success(fmt.Sprintf("Generated %d routers and %d services", routersGenerated, servicesGenerated))

	// Auto-register routers in main.py
	if !checkMode && !dryRun {
		progress.Start("Registering routers in main.py")
		if err := autoRegisterPythonRouters(backends); err != nil {
			return fmt.Errorf("failed to register routers: %w", err)
		}
		progress.Success("Routers registered in main.py")
	}

	fmt.Println()
	fmt.Println("✅ Code generation complete!")
	fmt.Println()
	fmt.Println("📁 Generated files:")
	fmt.Printf("   - %d router files in routers/\n", routersGenerated)
	fmt.Printf("   - %d service files in services/\n", servicesGenerated)
	fmt.Println("   - Updated main.py with router imports")
	fmt.Println()
	fmt.Println("🚀 Next steps:")
	fmt.Println("   1. Review generated files")
	fmt.Println("   2. Run: uvicorn main:app --reload")
	fmt.Println("   3. Visit: http://localhost:8080/docs")

	return nil
}

// generateFastAPIRouter generates router file for a service
func generateFastAPIRouter(serviceName string, endpoints []interface{}, baseURL string, async bool) error {
	tmplContent, err := templates.TemplateFS.ReadFile("python/fastapi/router_template.py.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read router template: %w", err)
	}

	tmpl, err := template.New("router").Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse router template: %w", err)
	}

	// Determine router prefix
	routerPrefix := fmt.Sprintf("/api/%s", strings.ToLower(serviceName))

	// Convert endpoints
	convertedEndpoints := make([]map[string]interface{}, 0)
	for _, ep := range endpoints {
		epMap, ok := ep.(map[string]interface{})
		if !ok {
			continue
		}

		method, _ := epMap["method"].(string)
		path, _ := epMap["path"].(string)
		name, _ := epMap["name"].(string)
		upstreamPath, _ := epMap["upstreamPath"].(string)

		if method == "" || path == "" {
			continue
		}

		// Strip router prefix from path to avoid duplication
		routePath := path
		if strings.HasPrefix(routePath, routerPrefix) {
			routePath = strings.TrimPrefix(routePath, routerPrefix)
		}
		if routePath == "" {
			routePath = "/"
		}

		method = strings.ToLower(method)
		functionName := sanitizeFunctionName(name, path, method)
		hasBody := method == "post" || method == "put" || method == "patch"

		convertedEndpoints = append(convertedEndpoints, map[string]interface{}{
			"Method":       method,
			"Path":         routePath,
			"FunctionName": functionName,
			"Description":  fmt.Sprintf("%s %s endpoint", strings.ToUpper(method), path),
			"HasBody":      hasBody,
			"UpstreamPath": upstreamPath,
		})
	}

	data := map[string]interface{}{
		"ServiceName":      serviceName,
		"ServiceNameLower": strings.ToLower(serviceName),
		"ServiceNameCamel": toCamelCase(serviceName),
		"ServiceNameUpper": strings.ToUpper(serviceName),
		"Endpoints":        convertedEndpoints,
		"Async":            async,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute router template: %w", err)
	}

	// Write to routers/ directory
	routerFile := filepath.Join("routers", fmt.Sprintf("%s_router.py", strings.ToLower(serviceName)))
	return os.WriteFile(routerFile, buf.Bytes(), utils.ProjectFilePerm)
}

// generateFastAPIService generates service file for a service
func generateFastAPIService(serviceName string, endpoints []interface{}, baseURL string, async bool) error {
	tmplContent, err := templates.TemplateFS.ReadFile("python/fastapi/service_template.py.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read service template: %w", err)
	}

	tmpl, err := template.New("service").Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse service template: %w", err)
	}

	// Convert endpoints
	convertedEndpoints := make([]map[string]interface{}, 0)
	for _, ep := range endpoints {
		epMap, ok := ep.(map[string]interface{})
		if !ok {
			continue
		}

		method, _ := epMap["method"].(string)
		path, _ := epMap["path"].(string)
		name, _ := epMap["name"].(string)
		upstreamPath, _ := epMap["upstreamPath"].(string)

		if method == "" || path == "" {
			continue
		}

		if upstreamPath == "" {
			upstreamPath = path
		}

		method = strings.ToLower(method)
		functionName := sanitizeFunctionName(name, path, method)
		hasBody := method == "post" || method == "put" || method == "patch"

		convertedEndpoints = append(convertedEndpoints, map[string]interface{}{
			"Method":       method,
			"Path":         path,
			"FunctionName": functionName,
			"Description":  fmt.Sprintf("Call %s endpoint", upstreamPath),
			"HasBody":      hasBody,
			"UpstreamPath": upstreamPath,
		})
	}

	data := map[string]interface{}{
		"ServiceName":      serviceName,
		"ServiceNameCamel": toCamelCase(serviceName),
		"ServiceNameUpper": strings.ToUpper(serviceName),
		"BaseURL":          baseURL,
		"Endpoints":        convertedEndpoints,
		"Async":            async,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute service template: %w", err)
	}

	// Write to services/ directory
	serviceFile := filepath.Join("services", fmt.Sprintf("%s_service.py", strings.ToLower(serviceName)))
	return os.WriteFile(serviceFile, buf.Bytes(), utils.ProjectFilePerm)
}

// autoRegisterPythonRouters registers routers in main.py
func autoRegisterPythonRouters(backends []interface{}) error {
	mainFile := "main.py"
	content, err := os.ReadFile(mainFile)
	if err != nil {
		return fmt.Errorf("failed to read main.py: %w", err)
	}

	contentStr := string(content)

	// Generate import statements
	imports := []string{}
	includes := []string{}

	for _, backend := range backends {
		backendMap, ok := backend.(map[string]interface{})
		if !ok {
			continue
		}

		serviceName, _ := backendMap["name"].(string)
		if serviceName == "" {
			continue
		}

		serviceNameLower := strings.ToLower(serviceName)
		imports = append(imports, fmt.Sprintf("from routers.%s_router import router as %s_router", serviceNameLower, serviceNameLower))
		includes = append(includes, fmt.Sprintf("app.include_router(%s_router)", serviceNameLower))
	}

	// Find and replace import marker
	importMarker := "# BFFGEN_ROUTER_IMPORTS - Auto-generated router imports will be inserted here by bffgen generate"
	if strings.Contains(contentStr, importMarker) {
		importBlock := strings.Join(imports, "\n")
		contentStr = strings.Replace(contentStr, importMarker, importMarker+"\n"+importBlock, 1)
	}

	// Find and replace include marker
	includeMarker := "# BFFGEN_ROUTER_INCLUDES - Auto-generated router includes will be inserted here by bffgen generate"
	if strings.Contains(contentStr, includeMarker) {
		includeBlock := strings.Join(includes, "\n")
		contentStr = strings.Replace(contentStr, includeMarker, includeMarker+"\n"+includeBlock, 1)
	}

	return os.WriteFile(mainFile, []byte(contentStr), utils.ProjectFilePerm)
}

// sanitizeFunctionName creates a valid Python function name from endpoint details
func sanitizeFunctionName(name, path, method string) string {
	if name != "" {
		// Use provided name, convert to snake_case
		return toSnakeCase(name)
	}

	// Generate from method and path
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	parts := []string{method}

	for _, part := range pathParts {
		// Skip path parameters
		if strings.HasPrefix(part, "{") || strings.HasPrefix(part, ":") {
			continue
		}
		parts = append(parts, part)
	}

	return toSnakeCase(strings.Join(parts, "_"))
}

// capitalizeWord capitalizes the first letter of a word (replaces deprecated strings.Title)
func capitalizeWord(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

// toCamelCase converts string to CamelCase
func toCamelCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})

	for i, word := range words {
		words[i] = capitalizeWord(word)
	}

	return strings.Join(words, "")
}

// toSnakeCase converts string to snake_case
func toSnakeCase(s string) string {
	// Replace special characters with underscores
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}, s)

	// Convert camelCase to snake_case
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}

	return strings.ToLower(strings.Trim(string(result), "_"))
}
