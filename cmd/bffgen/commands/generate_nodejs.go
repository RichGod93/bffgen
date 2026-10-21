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
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/RichGod93/bffgen/v3/internal/scaffolding"
	"github.com/RichGod93/bffgen/v3/internal/templates"
	"github.com/RichGod93/bffgen/v3/internal/utils"
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

	// Render every file first, so an invalid config writes nothing
	files := map[string]string{}
	var services []string
	names := backendNames{}
	usesRoles := false
	for _, backend := range backends {
		backendMap, ok := backend.(map[string]interface{})
		if !ok {
			continue
		}

		serviceName, _ := backendMap["name"].(string)
		if serviceName == "" {
			continue
		}
		if err := names.add(serviceName, "file:"+strings.ToLower(serviceName), "env:"+nodeEnvKey(serviceName), "id:"+nodeIdentifier(serviceName)); err != nil {
			return err
		}

		// Skip backends with no endpoints
		endpoints, ok := backendMap["endpoints"].([]interface{})
		if !ok || len(endpoints) == 0 {
			LogVerboseCommand("Skipping %s (no endpoints defined)", serviceName)
			fmt.Printf("⏭️  Skipped service '%s' (no endpoints defined)\n", serviceName)
			continue
		}

		if err := renderNodeJSBackendFiles(files, serviceName, backendMap, framework); err != nil {
			return fmt.Errorf("backend %s: %w", serviceName, err)
		}
		services = append(services, serviceName)
		for _, e := range endpoints {
			if endpoint, ok := e.(map[string]interface{}); ok && endpoint["roles"] != nil {
				usesRoles = true // validated by renderNodeJSBackendFiles
			}
		}
	}

	aggregates, err := parseAggregates(config, false)
	if err != nil {
		return err
	}
	if len(aggregates) > 0 {
		if other, ok := names["file:"+aggregatesFile]; ok {
			return fmt.Errorf("backend %q: the name is reserved for src/routes/%s.js when \"aggregates\" are configured; rename the backend", other, aggregatesFile)
		}
		content, err := renderNodeAggregates(config, aggregates, framework)
		if err != nil {
			return err
		}
		files["src/routes/"+aggregatesFile+".js"] = content
		services = append(services, aggregatesFile)
		for _, a := range aggregates {
			usesRoles = usesRoles || len(a.Roles) > 0
		}
	}

	progress.Success("Configuration loaded")

	// Plan every write before making any, so an error (e.g. a corrupt manifest) writes nothing
	write := !checkMode && !dryRun
	runtimeActions, err := ensureNodeRuntimeFiles(".", framework, false)
	if err != nil {
		return fmt.Errorf("failed to check runtime files: %w", err)
	}
	if conflict := runtimeConflict(runtimeActions, filepath.Join("src", "middleware", "authorize.js")); conflict != "" && usesRoles {
		return fmt.Errorf("routes use \"roles\", but %s is your own file (no bffgen:runtime marker); rename it and run 'bffgen generate' again", conflict)
	}
	plan, err := planGeneratedWrites(".", files, forceMode)
	if err != nil {
		return err
	}
	indexPath := filepath.Join("src", "index.js")
	var indexFix *legacyFix
	if usesRoles {
		if indexFix, err = planInvalidTokenStatusFix(indexPath); err != nil {
			return err
		}
	}

	// Generated routes require the runtime files; create or upgrade them first
	if write {
		if runtimeActions, err = ensureNodeRuntimeFiles(".", framework, true); err != nil {
			return fmt.Errorf("failed to update runtime files: %w", err)
		}
	}
	printRuntimeActions(runtimeActions, write)

	fmt.Printf("📝 Generating %s routes, controllers and services for %d backends\n", framework, len(services)-min(len(aggregates), 1))
	if len(aggregates) > 0 {
		fmt.Printf("📝 Generating %d aggregate routes in src/routes/%s.js\n", len(aggregates), aggregatesFile)
	}
	if write {
		if err := plan.apply("."); err != nil {
			return err
		}
	}
	plan.print(write)
	actions := plan.Actions

	if indexFix != nil {
		if write {
			if err := indexFix.apply(); err != nil {
				return err
			}
		}
		fmt.Printf("%s %s: invalid or expired tokens get 401 instead of 403, since 403 now means a missing role", pick(write, "♻️  Updated", "🔍 Would update"), indexPath)
		fmt.Printf(" (previous copy: %s)\n", indexFix.backup)
	}

	indexChanged, err := autoRegisterRoutes(framework, services, write)
	switch {
	case err != nil:
		fmt.Printf("⚠️  Warning: Failed to auto-register routes: %v\n", err)
		fmt.Println("💡 You can manually import routes in src/index.js")
	case indexChanged && write:
		fmt.Println("✅ Registered routes in src/index.js")
	case indexChanged:
		fmt.Println("🔍 Would update the route registrations in src/index.js")
	}

	if !write {
		if checkMode && (runtimeChangesPending(runtimeActions) || generatedChangesPending(actions) || indexChanged || indexFix != nil) {
			return fmt.Errorf("generated code is out of date; run 'bffgen generate'")
		}
		fmt.Println()
		fmt.Println("🔍 Nothing was written.")
		return nil
	}

	progress.Success("All files generated")
	fmt.Println()
	fmt.Println("✅ Code generation completed!")
	fmt.Println()
	fmt.Println("💡 Next steps:")
	fmt.Println("   1. Review generated files:")
	fmt.Println("      - src/routes/      (route handlers)")
	fmt.Println("      - src/controllers/ (business logic)")
	fmt.Println("      - src/services/    (HTTP communication)")
	fmt.Println("   2. Start your backend services")
	fmt.Println("   3. Run: npm run dev")
	fmt.Println("   4. Test your endpoints")
	fmt.Println()
	fmt.Println("   Files you edit are kept on the next 'bffgen generate' (see .bffgen/generated.json).")

	LogVerboseCommand("Code generation completed successfully")

	return nil
}

// jsStringArray renders values as a JS array literal, e.g. ['admin', 'support']
func jsStringArray(values []string) string {
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = "'" + template.JSEscapeString(v) + "'"
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// pythonStringArgs renders values as Python string arguments, e.g. "admin", "support"
func pythonStringArgs(values []string) string {
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = strconv.Quote(v)
	}
	return strings.Join(parts, ", ")
}

// legacyInvalidTokenStatus matches the line projects made before bffgen 3.8 used to answer an
// invalid or expired token with 403 (in src/index.js; generated Fastify routes use that check).
// RFC 9110: 401 means not authenticated, 403 authenticated but not allowed (a missing role).
var legacyInvalidTokenStatus = regexp.MustCompile(`(return (?:reply|res)\.status\()403(\)\.(?:send|json)\(\{ error: 'Invalid or expired token' \}\);)`)

// legacyFix is a planned one-line change to a file the user owns; apply saves the old copy first
type legacyFix struct {
	path, backup string
	old, updated []byte
}

func (f *legacyFix) apply() error {
	if err := os.WriteFile(f.backup, f.old, utils.ProjectFilePerm); err != nil {
		return err
	}
	return os.WriteFile(f.path, f.updated, utils.ProjectFilePerm)
}

// planInvalidTokenStatusFix plans replacing that exact line with its 401 version; the rest of
// src/index.js is the user's. Only projects that use roles get it: there, a 403 for a bad token
// would be indistinguishable from a missing role. It returns nil if nothing needs to change.
func planInvalidTokenStatusFix(path string) (*legacyFix, error) {
	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !legacyInvalidTokenStatus.Match(content) {
		return nil, nil
	}
	return &legacyFix{
		path:    path,
		backup:  backupPath("", path),
		old:     content,
		updated: legacyInvalidTokenStatus.ReplaceAll(content, []byte("${1}401${2}")),
	}, nil
}

// endpointRoles returns an endpoint's "roles": a non-empty list of role names, any one of which
// is required. Duplicates are dropped; order is kept so output is deterministic. Role names can't
// contain whitespace or commas, because a roles claim may be one space/comma-separated string.
func endpointRoles(endpoint map[string]interface{}) ([]string, error) {
	raw, ok := endpoint["roles"]
	if !ok || raw == nil {
		return nil, nil
	}
	list, ok := raw.([]interface{})
	if !ok || len(list) == 0 {
		return nil, fmt.Errorf(`"roles" must be a non-empty list of role names, e.g. ["admin"]`)
	}
	var roles []string
	seen := map[string]bool{}
	for _, r := range list {
		role, ok := r.(string)
		if !ok || role == "" || strings.ContainsAny(role, " \t\r\n,") {
			return nil, fmt.Errorf(`"roles" must contain non-empty role names without spaces or commas, got %v`, r)
		}
		if !seen[role] {
			seen[role] = true
			roles = append(roles, role)
		}
	}
	return roles, nil
}

// backendNames rejects backends whose names collide: same name ignoring case (one file on
// case-insensitive disks such as macOS and Windows), same env var, or same module/identifier
type backendNames map[string]string

func (b backendNames) add(name string, keys ...string) error {
	for _, key := range keys {
		if other, ok := b[key]; ok {
			if other == name {
				return fmt.Errorf("backend %q is defined twice; backend names must be unique", name)
			}
			return fmt.Errorf("backends %q and %q would generate the same files or variables; rename one", other, name)
		}
	}
	for _, key := range keys {
		b[key] = name
	}
	return nil
}

// nodeServiceName matches backend names that are safe as file names and in require() paths
var nodeServiceName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)

// renderNodeJSBackendFiles renders the route, controller and service files for one backend into files
func renderNodeJSBackendFiles(files map[string]string, serviceName string, backend map[string]interface{}, framework string) error {
	if !nodeServiceName.MatchString(serviceName) {
		return fmt.Errorf("name %q must contain only letters, digits, '-' and '_'", serviceName)
	}

	route, err := renderNodeJSRouteFile(serviceName, backend, framework)
	if err != nil {
		return err
	}
	files["src/routes/"+serviceName+".js"] = route

	if err := renderNodeJSControllerFiles(files, serviceName, backend, framework); err != nil {
		return err
	}

	service, err := renderNodeJSServiceFile(serviceName, backend, framework)
	if err != nil {
		return err
	}
	files["src/services/"+serviceName+".service.js"] = service
	return nil
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
		roles, err := endpointRoles(endpoint)
		if err != nil {
			return "", fmt.Errorf("endpoint %s %s: %w", method, path, err)
		}
		requiresAuth = requiresAuth || len(roles) > 0

		method = strings.ToUpper(method)
		if !nodeRouteMethods[method] {
			return "", fmt.Errorf("endpoint %s %s: unsupported method", method, path)
		}
		path, backendPath = normalizeRouteParams(path), normalizeRouteParams(backendPath)
		if err := validateNodeRoutePaths(path, backendPath); err != nil {
			return "", fmt.Errorf("endpoint %s %s: %w", method, path, err)
		}

		data.AnyAuth = data.AnyAuth || requiresAuth
		data.AnyRoles = data.AnyRoles || len(roles) > 0
		data.Endpoints = append(data.Endpoints, templates.RouteEndpointData{
			Method:       method,
			MethodLower:  strings.ToLower(method),
			Path:         path,
			BackendPath:  backendPath,
			RequiresAuth: requiresAuth,
			Roles:        roles,
			RolesJS:      jsStringArray(roles),
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

// renderNodeJSControllerFiles renders a backend's controllers: <name>.controller.js (aggregator
// pattern) and <name>.controller.basic.js
func renderNodeJSControllerFiles(files map[string]string, serviceName string, backend map[string]interface{}, framework string) error {
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

	content, err := renderControllerTemplate(loader, framework, "controller-aggregator.js.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render controller: %w", err)
	}
	files["src/controllers/"+serviceName+".controller.js"] = content

	basicContent, err := renderControllerTemplate(loader, framework, "controller-basic.js.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render basic controller: %w", err)
	}
	files["src/controllers/"+serviceName+".controller.basic.js"] = basicContent

	return nil
}

// renderNodeJSServiceFile renders the service file for a backend
func renderNodeJSServiceFile(serviceName string, backend map[string]interface{}, framework string) (string, error) {
	baseURL, _ := backend["baseUrl"].(string)
	endpoints, _ := backend["endpoints"].([]interface{})

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

	content, err := renderServiceTemplate(loader, framework, "service-template.js.tmpl", data)
	if err != nil {
		return "", fmt.Errorf("failed to render service: %w", err)
	}
	return content, nil
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

// autoRegisterRoutes rewrites the route registrations between the bffgen:begin:routes and
// bffgen:end:routes markers in src/index.js. It reports whether the block changes (or would
// change, when write is false). Projects without the markers are left alone.
func autoRegisterRoutes(framework string, services []string, write bool) (bool, error) {
	indexPath := filepath.Join("src", "index.js")

	content, err := os.ReadFile(indexPath)
	if err != nil {
		return false, fmt.Errorf("failed to read index.js: %w", err)
	}
	contentStr := string(content)

	var routeContent strings.Builder
	for _, serviceName := range services {
		if framework == "fastify" {
			routeContent.WriteString(fmt.Sprintf("    // %s routes\n", serviceName))
			routeContent.WriteString(fmt.Sprintf("    await fastify.register(require('./routes/%s'));\n", serviceName))
		} else {
			routeContent.WriteString(fmt.Sprintf("// %s routes\n", serviceName))
			routeContent.WriteString(fmt.Sprintf("app.use(require('./routes/%s'));\n", serviceName))
		}
	}

	marker := scaffolding.CustomMarkers("routes")
	sections, err := scaffolding.FindSections(contentStr, marker)
	if err != nil || len(sections) == 0 {
		LogVerboseCommand("No bffgen:begin:routes markers in src/index.js; register routes manually")
		return false, nil
	}

	updatedContent, err := scaffolding.ReplaceSection(contentStr, sections[0], routeContent.String())
	if err != nil {
		return false, fmt.Errorf("failed to replace section: %w", err)
	}
	if updatedContent == contentStr {
		return false, nil
	}
	if !write {
		return true, nil
	}
	if err := os.WriteFile(indexPath, []byte(updatedContent), utils.ProjectFilePerm); err != nil {
		return false, fmt.Errorf("failed to write index.js: %w", err)
	}
	return true, nil
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
