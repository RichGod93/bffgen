// File: generate_nodejs.go
// Purpose: Node.js-specific code generation for Express and Fastify frameworks
// Contains all logic for generating Node.js BFF projects

package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/RichGod93/bffgen/internal/scaffolding"
	"github.com/RichGod93/bffgen/internal/templates"
	"github.com/RichGod93/bffgen/internal/utils"
)

// generateNodeJS generates code for Node.js projects
func generateNodeJS() error {
	// Create progress tracker
	progress := utils.NewQuietProgress(verbose)

	if checkMode {
		fmt.Println("🔍 Check mode: Analyzing what would be changed")
	} else if dryRun {
		fmt.Println("🔍 Dry run: Showing what would be changed")
	} else {
		fmt.Println("🔧 Generating Node.js routes from bffgen.config.json")
	}
	fmt.Println()

	progress.Start("Loading configuration")

	// Check if config file exists
	if _, err := os.Stat("bffgen.config.json"); os.IsNotExist(err) {
		fmt.Println("❌ bffgen.config.json not found in current directory")
		fmt.Println("💡 Run 'bffgen init <project-name>' first or navigate to a BFF project directory")
		return fmt.Errorf("config file not found")
	}

	// Load bffgen.config.json
	configData, err := os.ReadFile("bffgen.config.json")
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
		LogWarning("No backends configured in bffgen.config.json")
		fmt.Println("💡 Add backends using 'bffgen add-route' or 'bffgen add-template'")
		return nil
	}

	LogVerboseCommand("Found %d backends to generate", len(backends))

	// Detect framework (Express or Fastify)
	framework := "express"
	if project, ok := config["project"].(map[string]interface{}); ok {
		if fw, ok := project["framework"].(string); ok {
			framework = fw
		}
	}

	progress.Success("Configuration loaded")

	// Generated routes require the runtime files; create or upgrade them first
	write := !checkMode && !dryRun
	actions, err := ensureNodeRuntimeFiles(".", framework, write)
	if err != nil {
		return fmt.Errorf("failed to update runtime files: %w", err)
	}
	for _, a := range actions {
		switch {
		case !write:
			fmt.Printf("🔍 Would %s %s\n", strings.TrimSuffix(a.Action, "d"), a.Path)
		case a.Backup != "":
			fmt.Printf("♻️  Updated %s to the current bffgen runtime (previous copy: %s)\n", a.Path, a.Backup)
		default:
			fmt.Printf("✅ Created %s\n", a.Path)
		}
	}

	fmt.Printf("📝 Generating routes for %s\n", framework)
	progress.Start(fmt.Sprintf("Generating files for %d backends", len(backends)))

	// Generate route files, controllers, and services for each backend
	routesGenerated := 0
	controllersGenerated := 0
	servicesGenerated := 0

	// Default controller type
	controllerType := "both" // Can be made configurable via flag

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
			fmt.Printf("⏭️  Skipped service '%s' (no endpoints defined)\n", serviceName)
			continue
		}

		// Generate route file
		if err := generateNodeJSRouteFile(serviceName, backendMap, framework); err != nil {
			fmt.Printf("⚠️  Warning: Failed to generate routes for %s: %v\n", serviceName, err)
			continue
		}
		routesGenerated++
		fmt.Printf("✅ Generated routes for service: %s\n", serviceName)

		// Generate controller file
		if err := generateNodeJSControllerFile(serviceName, backendMap, framework, controllerType); err != nil {
			fmt.Printf("⚠️  Warning: Failed to generate controller for %s: %v\n", serviceName, err)
		} else {
			controllersGenerated++
			fmt.Printf("✅ Generated controller for service: %s\n", serviceName)
		}

		// Generate service file
		if err := generateNodeJSServiceFile(serviceName, backendMap, framework); err != nil {
			fmt.Printf("⚠️  Warning: Failed to generate service for %s: %v\n", serviceName, err)
		} else {
			servicesGenerated++
			fmt.Printf("✅ Generated service for service: %s\n", serviceName)
		}
	}

	// Auto-register routes in index.js
	if !checkMode && !dryRun {
		if err := autoRegisterRoutes(framework, backends); err != nil {
			fmt.Printf("⚠️  Warning: Failed to auto-register routes: %v\n", err)
			fmt.Println("💡 You can manually import routes in src/index.js")
		} else {
			fmt.Println("✅ Auto-registered routes in src/index.js")
		}
	}

	if !checkMode && !dryRun {
		progress.Success("All files generated")

		fmt.Println()
		fmt.Printf("✅ Code generation completed!\n")
		fmt.Printf("   📁 Generated %d route files in src/routes/\n", routesGenerated)
		fmt.Printf("   🎮 Generated %d controller files in src/controllers/\n", controllersGenerated)
		fmt.Printf("   🔧 Generated %d service files in src/services/\n", servicesGenerated)
		fmt.Println()
		fmt.Println("🚀 Routes, controllers, and services are ready to use")
		fmt.Println()
		fmt.Println("💡 Next steps:")
		fmt.Println("   1. Review generated files:")
		fmt.Println("      - src/routes/     (route handlers)")
		fmt.Println("      - src/controllers/ (business logic)")
		fmt.Println("      - src/services/    (HTTP communication)")
		fmt.Println("   2. Start your backend services")
		fmt.Println("   3. Run: npm run dev")
		fmt.Println("   4. Test your endpoints")
	}

	LogVerboseCommand("Code generation completed successfully")

	return nil
}

// generateNodeJSRouteFile generates a route file for a Node.js service
func generateNodeJSRouteFile(serviceName string, backend map[string]interface{}, framework string) error {
	// Create routes directory if it doesn't exist
	routesDir := filepath.Join("src", "routes")
	if err := os.MkdirAll(routesDir, utils.ProjectDirPerm); err != nil {
		return err
	}

	content, err := renderNodeJSRouteFile(serviceName, backend, framework)
	if err != nil {
		return err
	}

	filename := filepath.Join(routesDir, fmt.Sprintf("%s.js", serviceName))
	return os.WriteFile(filename, []byte(content), utils.ProjectFilePerm)
}

// renderNodeJSRouteFile renders the route file for one backend from routes-template.js.tmpl
func renderNodeJSRouteFile(serviceName string, backend map[string]interface{}, framework string) (string, error) {
	baseURL, _ := backend["baseUrl"].(string)
	endpoints, _ := backend["endpoints"].([]interface{})

	data := &templates.RouteTemplateData{
		ServiceName:  serviceName,
		FunctionName: nodeIdentifier(serviceName) + "Routes",
		EnvKey:       nodeEnvKey(serviceName),
		BaseURL:      baseURL,
	}
	if err := applyResilienceConfig(data, backend); err != nil {
		return "", fmt.Errorf("backend %s: %w", serviceName, err)
	}

	for _, ep := range endpoints {
		endpoint, ok := ep.(map[string]interface{})
		if !ok {
			continue
		}

		path, _ := endpoint["exposeAs"].(string)
		method, _ := endpoint["method"].(string)
		backendPath, _ := endpoint["path"].(string)
		requiresAuth, _ := endpoint["requiresAuth"].(bool)

		if path == "" || method == "" {
			continue
		}

		method = strings.ToUpper(method)
		if !nodeRouteMethods[method] {
			return "", fmt.Errorf("endpoint %s %s: unsupported method", method, path)
		}
		path, backendPath = normalizeRouteParams(path), normalizeRouteParams(backendPath)
		if err := validateNodeRoutePaths(path, backendPath); err != nil {
			return "", fmt.Errorf("endpoint %s %s: %w", method, path, err)
		}

		data.AnyAuth = data.AnyAuth || requiresAuth
		data.Endpoints = append(data.Endpoints, templates.RouteEndpointData{
			Method:       method,
			MethodLower:  strings.ToLower(method),
			Path:         path,
			BackendPath:  backendPath,
			RequiresAuth: requiresAuth,
		})
	}

	langType := scaffolding.LanguageNodeExpress
	if framework == "fastify" {
		langType = scaffolding.LanguageNodeFastify
	}

	tmplContent, err := templates.NewTemplateLoader(langType).LoadTemplate(framework, "routes-template.js.tmpl")
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("routes-template.js.tmpl").Parse(tmplContent)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render routes for %s: %w", serviceName, err)
	}

	return buf.String(), nil
}

// generateNodeJSControllerFile generates controller files for a Node.js service
func generateNodeJSControllerFile(serviceName string, backend map[string]interface{}, framework, controllerType string) error {
	// Create controllers directory if it doesn't exist
	controllersDir := filepath.Join("src", "controllers")
	if err := os.MkdirAll(controllersDir, utils.ProjectDirPerm); err != nil {
		return err
	}

	endpoints, ok := backend["endpoints"].([]interface{})
	if !ok {
		return nil // No endpoints, skip controller generation
	}

	// Build endpoint data
	var endpointData []map[string]interface{}
	for _, ep := range endpoints {
		endpoint, ok := ep.(map[string]interface{})
		if !ok {
			continue
		}

		path, _ := endpoint["exposeAs"].(string)
		method, _ := endpoint["method"].(string)
		backendPath, _ := endpoint["path"].(string)
		requiresAuth, _ := endpoint["requiresAuth"].(bool)

		if path == "" || method == "" {
			continue
		}

		handlerName := nodeHandlerName(method, path)

		endpointData = append(endpointData, map[string]interface{}{
			"Path":              path,
			"Method":            method,
			"BackendPath":       backendPath,
			"RequiresAuth":      requiresAuth,
			"HandlerName":       handlerName,
			"HandlerNamePascal": strings.ToUpper(handlerName[:1]) + handlerName[1:],
		})
	}

	// Determine template to use
	templateType := "basic"
	if controllerType == "aggregator" || controllerType == "both" {
		templateType = "aggregator"
	}

	// Load template loader
	var langType scaffolding.LanguageType
	if framework == "fastify" {
		langType = scaffolding.LanguageNodeFastify
	} else {
		langType = scaffolding.LanguageNodeExpress
	}

	loader := templates.NewTemplateLoader(langType)

	// Prepare template data
	data := &templates.ControllerTemplateData{
		ServiceName:       serviceName,
		ServiceNamePascal: templates.ToPascalCase(serviceName),
		Endpoints:         convertToEndpointData(endpointData),
	}

	// Render controller template
	templateName := fmt.Sprintf("controller-%s.js.tmpl", templateType)
	content, err := renderControllerTemplate(loader, framework, templateName, data)
	if err != nil {
		return fmt.Errorf("failed to render controller: %w", err)
	}

	// Write controller file
	filename := filepath.Join(controllersDir, fmt.Sprintf("%s.controller.js", serviceName))
	if err := os.WriteFile(filename, []byte(content), utils.ProjectFilePerm); err != nil {
		return err
	}

	// If "both", also generate basic controller
	if controllerType == "both" {
		basicContent, err := renderControllerTemplate(loader, framework, "controller-basic.js.tmpl", data)
		if err != nil {
			return fmt.Errorf("failed to render basic controller: %w", err)
		}

		basicFilename := filepath.Join(controllersDir, fmt.Sprintf("%s.controller.basic.js", serviceName))
		if err := os.WriteFile(basicFilename, []byte(basicContent), utils.ProjectFilePerm); err != nil {
			return err
		}
	}

	return nil
}

// generateNodeJSServiceFile generates service files for a Node.js service
func generateNodeJSServiceFile(serviceName string, backend map[string]interface{}, framework string) error {
	// Create services directory if it doesn't exist
	servicesDir := filepath.Join("src", "services")
	if err := os.MkdirAll(servicesDir, utils.ProjectDirPerm); err != nil {
		return err
	}

	baseURL, _ := backend["baseUrl"].(string)
	endpoints, ok := backend["endpoints"].([]interface{})
	if !ok {
		return nil // No endpoints, skip service generation
	}

	// Build endpoint data
	var endpointData []map[string]interface{}
	for _, ep := range endpoints {
		endpoint, ok := ep.(map[string]interface{})
		if !ok {
			continue
		}

		path, _ := endpoint["exposeAs"].(string)
		method, _ := endpoint["method"].(string)
		backendPath, _ := endpoint["path"].(string)

		if path == "" || method == "" {
			continue
		}

		handlerName := nodeHandlerName(method, path)

		endpointData = append(endpointData, map[string]interface{}{
			"Method":      method,
			"BackendPath": normalizeRouteParams(backendPath),
			"HandlerName": handlerName,
		})
	}

	// Load template loader
	var langType scaffolding.LanguageType
	if framework == "fastify" {
		langType = scaffolding.LanguageNodeFastify
	} else {
		langType = scaffolding.LanguageNodeExpress
	}

	loader := templates.NewTemplateLoader(langType)

	// Prepare template data
	data := &templates.ServiceTemplateData{
		ServiceName:       serviceName,
		ServiceNamePascal: templates.ToPascalCase(serviceName),
		BaseURL:           baseURL,
		EnvKey:            nodeEnvKey(serviceName),
		Endpoints:         convertToEndpointData(endpointData),
	}

	// Render service template
	content, err := renderServiceTemplate(loader, framework, "service-template.js.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render service: %w", err)
	}

	// Write service file
	filename := filepath.Join(servicesDir, fmt.Sprintf("%s.service.js", serviceName))
	if err := os.WriteFile(filename, []byte(content), utils.ProjectFilePerm); err != nil {
		return err
	}

	return nil
}

// Helper function to convert endpoint data
func convertToEndpointData(data []map[string]interface{}) []templates.EndpointData {
	result := make([]templates.EndpointData, 0, len(data))
	for _, d := range data {
		endpoint := templates.EndpointData{}
		if v, ok := d["Path"].(string); ok {
			endpoint.Path = v
		}
		if v, ok := d["Method"].(string); ok {
			endpoint.Method = v
		}
		if v, ok := d["BackendPath"].(string); ok {
			endpoint.BackendPath = v
		}
		if v, ok := d["RequiresAuth"].(bool); ok {
			endpoint.RequiresAuth = v
		}
		if v, ok := d["HandlerName"].(string); ok {
			endpoint.HandlerName = v
		}
		if v, ok := d["HandlerNamePascal"].(string); ok {
			endpoint.HandlerNamePascal = v
		}
		result = append(result, endpoint)
	}
	return result
}

// Helper to render controller template
func renderControllerTemplate(loader *templates.TemplateLoader, framework, templateName string, data *templates.ControllerTemplateData) (string, error) {
	tmplContent, err := loader.LoadTemplate(framework, templateName)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New(templateName).Parse(tmplContent)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Helper to render service template
func renderServiceTemplate(loader *templates.TemplateLoader, framework, templateName string, data *templates.ServiceTemplateData) (string, error) {
	tmplContent, err := loader.LoadTemplate(framework, templateName)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New(templateName).Parse(tmplContent)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// autoRegisterRoutes automatically imports and registers routes in the main index.js file
func autoRegisterRoutes(framework string, backends []interface{}) error {
	indexPath := filepath.Join("src", "index.js")

	// Read existing index.js
	content, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("failed to read index.js: %w", err)
	}

	contentStr := string(content)

	// Generate route imports and registrations
	var routeContent strings.Builder

	for _, b := range backends {
		backend, ok := b.(map[string]interface{})
		if !ok {
			continue
		}

		serviceName, _ := backend["name"].(string)
		if serviceName == "" {
			continue
		}

		endpoints, ok := backend["endpoints"].([]interface{})
		if !ok || len(endpoints) == 0 {
			continue
		}

		// Add import and registration
		if framework == "fastify" {
			routeContent.WriteString(fmt.Sprintf("    // %s routes\n", serviceName))
			routeContent.WriteString(fmt.Sprintf("    await fastify.register(require('./routes/%s'));\n", serviceName))
		} else {
			routeContent.WriteString(fmt.Sprintf("// %s routes\n", serviceName))
			routeContent.WriteString(fmt.Sprintf("app.use(require('./routes/%s'));\n", serviceName))
		}
	}

	// Find and replace the route registration section using markers
	marker := scaffolding.CustomMarkers("routes")
	sections, err := scaffolding.FindSections(contentStr, marker)
	if err != nil || len(sections) == 0 {
		// Markers not found, skip auto-registration
		return nil
	}

	// Replace the section with new content
	updatedContent, err := scaffolding.ReplaceSection(contentStr, sections[0], routeContent.String())
	if err != nil {
		return fmt.Errorf("failed to replace section: %w", err)
	}

	// Write updated index.js
	if err := os.WriteFile(indexPath, []byte(updatedContent), utils.ProjectFilePerm); err != nil {
		return fmt.Errorf("failed to write index.js: %w", err)
	}

	return nil
}

// nodeHandlerName builds a valid camelCase JavaScript identifier from an HTTP method and route path,
// e.g. GET /api/user-profiles/:id -> getApiUserProfilesById
func nodeHandlerName(method, path string) string {
	var b strings.Builder
	b.WriteString(strings.ToLower(method))

	for _, segment := range strings.Split(path, "/") {
		isParam := strings.HasPrefix(segment, ":") || strings.HasPrefix(segment, "{")
		words := strings.FieldsFunc(segment, func(r rune) bool {
			return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9')
		})
		if isParam && len(words) > 0 {
			b.WriteString("By")
		}
		for _, word := range words {
			b.WriteString(strings.ToUpper(word[:1]) + word[1:])
		}
	}

	return b.String()
}

// Defaults for backends that don't configure timeout, retries or circuitBreaker
const (
	defaultBackendTimeoutMs      = 10000
	defaultBackendRetries        = 2
	defaultBreakerFailures       = 5
	defaultBreakerResetTimeoutMs = 30000
)

// applyResilienceConfig fills the timeout, retry and circuit breaker settings of a route file
// from a backend's config:
//
//	"timeout": 5000,          per-attempt timeout in ms (or a duration string such as "5s")
//	"retries": 2,             extra attempts for GET/HEAD/OPTIONS
//	"circuitBreaker": {"enabled": true, "failureThreshold": 5, "resetTimeout": 30000}
func applyResilienceConfig(data *templates.RouteTemplateData, backend map[string]interface{}) error {
	var err error
	if data.Timeout, err = configMillis(backend["timeout"], defaultBackendTimeoutMs); err != nil {
		return fmt.Errorf("timeout: %w", err)
	}
	if data.Retries, err = configInt(backend["retries"], defaultBackendRetries); err != nil {
		return fmt.Errorf("retries: %w", err)
	}

	breaker, _ := backend["circuitBreaker"].(map[string]interface{})
	data.BreakerEnabled = true
	if enabled, ok := breaker["enabled"].(bool); ok {
		data.BreakerEnabled = enabled
	}
	if data.FailureThreshold, err = configInt(breaker["failureThreshold"], defaultBreakerFailures); err != nil {
		return fmt.Errorf("circuitBreaker.failureThreshold: %w", err)
	}
	if data.ResetTimeout, err = configMillis(breaker["resetTimeout"], defaultBreakerResetTimeoutMs); err != nil {
		return fmt.Errorf("circuitBreaker.resetTimeout: %w", err)
	}
	if data.Timeout == 0 || data.FailureThreshold == 0 {
		return fmt.Errorf("timeout and circuitBreaker.failureThreshold must be greater than 0")
	}
	return nil
}

// configInt reads a non-negative whole number from JSON config, or returns def if unset
func configInt(v interface{}, def int) (int, error) {
	if v == nil {
		return def, nil
	}
	n, ok := v.(float64)
	if !ok || n < 0 || n != float64(int(n)) {
		return 0, fmt.Errorf("%v is not a non-negative whole number", v)
	}
	return int(n), nil
}

// configMillis reads a duration in ms from JSON config: a number of ms or a Go duration
// string such as "5s". Returns def if unset.
func configMillis(v interface{}, def int) (int, error) {
	if s, ok := v.(string); ok {
		d, err := time.ParseDuration(s)
		if err != nil || d < 0 {
			return 0, fmt.Errorf("%q is not a duration such as \"5s\" or a number of ms", s)
		}
		return int(d.Milliseconds()), nil
	}
	return configInt(v, def)
}

// nodeRouteMethods are the HTTP methods a generated route can proxy
var nodeRouteMethods = map[string]bool{
	"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true, "HEAD": true, "OPTIONS": true,
}

var (
	braceParam = regexp.MustCompile(`\{([A-Za-z_$][\w$]*)\}`)
	colonParam = regexp.MustCompile(`:([A-Za-z_$][\w$]*)`)
	// Characters allowed in a route path: RFC 3986 path characters, minus quotes and backslashes,
	// which would break out of the generated JS string literals
	safeRoutePath = regexp.MustCompile(`^/[A-Za-z0-9\-._~!$&()*+,;=:@%/]*$`)
)

// normalizeRouteParams rewrites OpenAPI-style {id} parameters to Express/Fastify :id parameters
func normalizeRouteParams(path string) string {
	return braceParam.ReplaceAllString(path, ":$1")
}

// routeParams returns the :param names in a route path, in order
func routeParams(path string) []string {
	var names []string
	for _, m := range colonParam.FindAllStringSubmatch(path, -1) {
		names = append(names, m[1])
	}
	return names
}

// validateNodeRoutePaths checks that both paths are safe to embed in generated JS and that every
// parameter the backend path needs is provided by the exposed path
func validateNodeRoutePaths(exposeAs, backendPath string) error {
	if !safeRoutePath.MatchString(exposeAs) {
		return fmt.Errorf("exposeAs %q must start with / and contain only URL path characters", exposeAs)
	}
	if !safeRoutePath.MatchString(backendPath) {
		return fmt.Errorf("path %q must start with / and contain only URL path characters", backendPath)
	}

	exposed := map[string]bool{}
	for _, name := range routeParams(exposeAs) {
		exposed[name] = true
	}
	for _, name := range routeParams(backendPath) {
		if !exposed[name] {
			return fmt.Errorf("backend path %s uses :%s, which exposeAs %s does not define", backendPath, name, exposeAs)
		}
	}
	return nil
}

// nodeEnvKey returns the environment variable prefix for a backend, e.g. user-service -> USER_SERVICE
func nodeEnvKey(serviceName string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' {
			return r - 'a' + 'A'
		}
		if r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			return r
		}
		return '_'
	}, serviceName)
}

// nodeIdentifier returns a camelCase JS identifier for a backend name, e.g. user-service -> userService
func nodeIdentifier(serviceName string) string {
	words := strings.FieldsFunc(serviceName, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9')
	})
	var b strings.Builder
	for i, word := range words {
		if i == 0 {
			b.WriteString(strings.ToLower(word[:1]) + word[1:])
		} else {
			b.WriteString(strings.ToUpper(word[:1]) + word[1:])
		}
	}
	id := b.String()
	if id == "" || id[0] >= '0' && id[0] <= '9' {
		id = "service" + id
	}
	return id
}
