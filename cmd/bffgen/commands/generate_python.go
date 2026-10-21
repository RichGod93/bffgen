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
	"regexp"
	"strings"
	"text/template"

	"github.com/RichGod93/bffgen/v3/internal/templates"
	"github.com/RichGod93/bffgen/v3/internal/utils"
)

// pythonConfigFile is the bffgen config file of Python projects
const pythonConfigFile = "bffgen.config.py.json"

// generatePython generates routers and services for FastAPI projects
func generatePython() error {
	progress := utils.NewQuietProgress(verbose)
	write := !checkMode && !dryRun

	if checkMode {
		fmt.Println("🔍 Check mode: Analyzing what would be changed")
	} else if dryRun {
		fmt.Println("🔍 Dry run: Showing what would be changed")
	} else {
		fmt.Println("🔧 Generating Python routes from " + pythonConfigFile)
	}
	fmt.Println()

	progress.Start("Loading configuration")

	configData, err := os.ReadFile(pythonConfigFile)
	if os.IsNotExist(err) {
		fmt.Println("❌ " + pythonConfigFile + " not found in current directory")
		fmt.Println("💡 Run 'bffgen init <project-name> --lang python-fastapi' first or navigate to a BFF project directory")
		return fmt.Errorf("config file not found")
	}
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	backends, ok := config["backends"].([]interface{})
	if !ok || len(backends) == 0 {
		LogWarning("No backends configured in " + pythonConfigFile)
		fmt.Println("💡 Add backends manually to " + pythonConfigFile + " or use 'bffgen add-template'")
		return nil
	}

	if project, ok := config["project"].(map[string]interface{}); ok {
		if async, ok := project["async"].(bool); ok && !async {
			fmt.Println("ℹ️  Generated proxy routes are always async; project.async only affects your own code")
		}
	}

	progress.Success("Configuration loaded")

	// Build all backends first so an invalid config writes nothing
	var generated []*pythonBackendData
	names := backendNames{}
	for _, backend := range backends {
		backendMap, ok := backend.(map[string]interface{})
		if !ok {
			continue
		}
		data, err := buildPythonBackend(backendMap)
		if err != nil {
			return err
		}
		if data == nil {
			continue
		}
		if err := names.add(data.ServiceName, "module:"+data.Module, "env:"+data.EnvKey); err != nil {
			return err
		}
		generated = append(generated, data)
	}

	files := map[string]string{}
	var modules []string
	for _, data := range generated {
		router, err := renderPythonTemplate("python/fastapi/router_template.py.tmpl", data)
		if err != nil {
			return err
		}
		service, err := renderPythonTemplate("python/fastapi/service_template.py.tmpl", data)
		if err != nil {
			return err
		}
		files["routers/"+data.Module+"_router.py"] = router
		files["services/"+data.Module+"_service.py"] = service
		modules = append(modules, data.Module)
	}
	aggregates, err := parseAggregates(config, true)
	if err != nil {
		return err
	}
	anyRoles := false
	for _, data := range generated {
		anyRoles = anyRoles || data.AnyRoles
	}
	if len(aggregates) > 0 {
		if other, ok := names["module:"+aggregatesFile]; ok {
			return fmt.Errorf("backend %q: the name is reserved for routers/%s_router.py when \"aggregates\" are configured; rename the backend", other, aggregatesFile)
		}
		router, err := renderPythonAggregates(config, aggregates)
		if err != nil {
			return err
		}
		files["routers/"+aggregatesFile+"_router.py"] = router
		modules = append(modules, aggregatesFile)
		for _, a := range aggregates {
			anyRoles = anyRoles || len(a.Roles) > 0
		}
	}
	registry, err := renderPythonTemplate("python/fastapi/routers_init.py.tmpl", map[string]interface{}{"Modules": modules})
	if err != nil {
		return err
	}
	files["routers/__init__.py"] = registry

	// Plan every write before making any, so an error (e.g. a corrupt manifest) writes nothing
	actions, err := ensurePythonRuntimeFiles(".", false)
	if err != nil {
		return fmt.Errorf("failed to check runtime files: %w", err)
	}
	if conflict := runtimeConflict(actions, filepath.Join("utils", "authorize.py")); conflict != "" && anyRoles {
		return fmt.Errorf("routes use \"roles\", but %s is your own file (no bffgen:runtime marker); rename it and run 'bffgen generate' again", conflict)
	}
	plan, err := planGeneratedWrites(".", files, forceMode)
	if err != nil {
		return err
	}
	mainChanged, err := registerPythonRouters("main.py", false)
	if err != nil {
		return fmt.Errorf("failed to register routers: %w", err)
	}
	gitignoreChanged, err := migratePythonGitignore(".gitignore", false)
	if err != nil {
		return err
	}

	// Generated routers require the runtime files; create or upgrade them first
	if write {
		if actions, err = ensurePythonRuntimeFiles(".", true); err != nil {
			return fmt.Errorf("failed to update runtime files: %w", err)
		}
	}
	printRuntimeActions(actions, write)

	if write {
		if err := plan.apply("."); err != nil {
			return err
		}
	}
	plan.print(write)

	if !write {
		if mainChanged {
			fmt.Println("🔍 Would register the routers in main.py")
		}
		if gitignoreChanged {
			fmt.Println("🔍 Would update .gitignore so .bffgen/generated.json is committed")
		}
		if checkMode && (runtimeChangesPending(actions) || generatedChangesPending(plan.Actions) || mainChanged || gitignoreChanged) {
			return fmt.Errorf("generated code is out of date; run 'bffgen generate'")
		}
		fmt.Println()
		fmt.Println("🔍 Nothing was written.")
		return nil
	}

	if _, err := registerPythonRouters("main.py", true); err != nil {
		return fmt.Errorf("failed to register routers: %w", err)
	}
	if _, err := migratePythonGitignore(".gitignore", true); err != nil {
		return err
	}
	warnStrictPythonSettings("config.py")

	progress.Success(fmt.Sprintf("Generated %d routers and %d services", len(generated), len(generated)))
	fmt.Println()
	fmt.Println("✅ Code generation complete!")
	fmt.Println()
	fmt.Println("📁 Generated files:")
	fmt.Printf("   - %d router files in routers/ (registered in routers/__init__.py)\n", len(generated))
	fmt.Printf("   - %d service files in services/\n", len(generated))
	if len(aggregates) > 0 {
		fmt.Printf("   - %d aggregate routes in routers/%s_router.py\n", len(aggregates), aggregatesFile)
	}
	fmt.Println()
	fmt.Println("🚀 Next steps:")
	fmt.Println("   1. Review generated files")
	fmt.Println("   2. Run: uvicorn main:app --reload")
	fmt.Println("   3. Visit: http://localhost:8080/docs")
	fmt.Println()
	fmt.Println("   Files you edit are kept on the next 'bffgen generate' (see .bffgen/generated.json).")

	return nil
}

// pythonBackendData is the template data for one backend's router and service
type pythonBackendData struct {
	ServiceName string
	Module      string // snake_case module name, e.g. user_service
	ClassName   string // service class, e.g. UserServiceService
	EnvKey      string
	BaseURL     string
	AnyAuth     bool // an endpoint requires authentication but no roles
	AnyRoles    bool
	Endpoints   []pythonEndpointData

	Timeout          int
	Retries          int
	BreakerEnabled   bool
	FailureThreshold int
	ResetTimeout     int
}

// pythonEndpointData is one proxied route
type pythonEndpointData struct {
	Method       string   // upper case
	Path         string   // exposed path, {param} style
	BackendPath  string   // backend path, {param} style
	Params       []string // path parameters, in order
	HandlerName  string
	RequiresAuth bool
	Roles        []string // any one of these is required (implies RequiresAuth)
	RolesArgs    string   // Roles as Python string arguments, e.g. "admin", "support"
	HasBody      bool
}

// pythonReservedNames can't be used as path parameter names: Python keywords and the
// argument names of generated route handlers and service methods
var pythonReservedNames = map[string]bool{
	"False": true, "None": true, "True": true, "and": true, "as": true, "assert": true, "async": true,
	"await": true, "break": true, "class": true, "continue": true, "def": true, "del": true, "elif": true,
	"else": true, "except": true, "finally": true, "for": true, "from": true, "global": true, "if": true,
	"import": true, "in": true, "is": true, "lambda": true, "nonlocal": true, "not": true, "or": true,
	"pass": true, "raise": true, "return": true, "try": true, "while": true, "with": true, "yield": true,
	"request": true, "self": true, "query": true, "body": true, "headers": true, "path": true,
}

var colonToBraceParam = regexp.MustCompile(`:([A-Za-z_][A-Za-z0-9_]*)`)

// buildPythonBackend validates one backend from bffgen.config.py.json and returns its
// template data, or nil if it has no endpoints.
//
// Endpoints use the Python config format, where "path" is the path the BFF exposes and
// "upstreamPath" the backend path (default: the same). Node-style endpoints with
// "exposeAs" (exposed) and "path" (backend) are accepted too. Both :id and {id}
// parameters work.
func buildPythonBackend(backend map[string]interface{}) (*pythonBackendData, error) {
	name, _ := backend["name"].(string)
	endpoints, _ := backend["endpoints"].([]interface{})
	if name == "" || len(endpoints) == 0 {
		return nil, nil
	}

	baseURL, _ := backend["baseUrl"].(string)
	if baseURL == "" {
		baseURL, _ = backend["url"].(string)
	}
	if baseURL == "" {
		return nil, fmt.Errorf("backend %s: baseUrl is required", name)
	}

	data := &pythonBackendData{
		ServiceName: name,
		Module:      pythonModuleName(name),
		ClassName:   toCamelCase(pythonModuleName(name)) + "Service",
		EnvKey:      nodeEnvKey(name),
		BaseURL:     baseURL,
	}

	if err := applyPythonResilienceConfig(data, backend); err != nil {
		return nil, fmt.Errorf("backend %s: %w", name, err)
	}

	usedNames := map[string]bool{}
	for _, ep := range endpoints {
		endpoint, ok := ep.(map[string]interface{})
		if !ok {
			continue
		}

		method, _ := endpoint["method"].(string)
		exposed, _ := endpoint["path"].(string)
		backendPath, _ := endpoint["upstreamPath"].(string)
		if exposeAs, ok := endpoint["exposeAs"].(string); ok && exposeAs != "" {
			exposed, backendPath = exposeAs, exposed
		}
		if backendPath == "" {
			backendPath = exposed
		}
		requiresAuth, _ := endpoint["requiresAuth"].(bool)
		endpointName, _ := endpoint["name"].(string)

		if exposed == "" || method == "" {
			continue
		}
		roles, err := endpointRoles(endpoint)
		if err != nil {
			return nil, fmt.Errorf("backend %s, endpoint %s %s: %w", name, method, exposed, err)
		}
		requiresAuth = requiresAuth || len(roles) > 0

		method = strings.ToUpper(method)
		if !nodeRouteMethods[method] {
			return nil, fmt.Errorf("backend %s, endpoint %s %s: unsupported method", name, method, exposed)
		}

		// Validate in :param form (shared with Node.js), then render in FastAPI's {param} form
		exposed, backendPath = normalizeRouteParams(exposed), normalizeRouteParams(backendPath)
		if err := validateNodeRoutePaths(exposed, backendPath); err != nil {
			return nil, fmt.Errorf("backend %s, endpoint %s %s: %w", name, method, exposed, err)
		}
		params := routeParams(exposed)
		for _, param := range params {
			if pythonReservedNames[param] {
				return nil, fmt.Errorf("backend %s, endpoint %s %s: path parameter %q is a reserved name in Python; rename it", name, method, exposed, param)
			}
		}

		handler := toSnakeCase(endpointName)
		if handler == "" {
			handler = toSnakeCase(nodeHandlerName(method, exposed))
		}
		if handler[0] >= '0' && handler[0] <= '9' || pythonReservedNames[handler] {
			handler = "route_" + handler
		}
		for base, n := handler, 2; usedNames[handler]; n++ {
			handler = fmt.Sprintf("%s_%d", base, n)
		}
		usedNames[handler] = true

		data.AnyAuth = data.AnyAuth || requiresAuth && len(roles) == 0
		data.AnyRoles = data.AnyRoles || len(roles) > 0
		data.Endpoints = append(data.Endpoints, pythonEndpointData{
			Method:       method,
			Path:         colonToBraceParam.ReplaceAllString(exposed, "{$1}"),
			BackendPath:  colonToBraceParam.ReplaceAllString(backendPath, "{$1}"),
			Params:       params,
			HandlerName:  handler,
			RequiresAuth: requiresAuth,
			Roles:        roles,
			RolesArgs:    pythonStringArgs(roles),
			HasBody:      method != "GET" && method != "HEAD",
		})
	}

	if len(data.Endpoints) == 0 {
		return nil, nil
	}
	return data, nil
}

// applyPythonResilienceConfig reads timeout, retries and circuitBreaker like the Node.js
// generator (times in ms or duration strings). Python configs used to give times in seconds,
// so times under 100 ms are rejected rather than silently making every call time out.
func applyPythonResilienceConfig(data *pythonBackendData, backend map[string]interface{}) error {
	backendCopy := map[string]interface{}{}
	for k, v := range backend {
		backendCopy[k] = v
	}
	// The old Python example used circuitBreaker.recoveryTimeout
	if breaker, ok := backend["circuitBreaker"].(map[string]interface{}); ok {
		if _, ok := breaker["resetTimeout"]; !ok && breaker["recoveryTimeout"] != nil {
			b := map[string]interface{}{}
			for k, v := range breaker {
				b[k] = v
			}
			b["resetTimeout"] = breaker["recoveryTimeout"]
			backendCopy["circuitBreaker"] = b
		}
	}

	var r templates.RouteTemplateData
	if err := applyResilienceConfig(&r, backendCopy); err != nil {
		return err
	}
	if r.Timeout < 100 {
		return fmt.Errorf("timeout %d is in milliseconds; did you mean %d or \"%ds\"?", r.Timeout, r.Timeout*1000, r.Timeout)
	}
	if r.ResetTimeout < 100 {
		return fmt.Errorf("circuitBreaker.resetTimeout %d is in milliseconds; did you mean %d or \"%ds\"?", r.ResetTimeout, r.ResetTimeout*1000, r.ResetTimeout)
	}
	data.Timeout, data.Retries = r.Timeout, r.Retries
	data.BreakerEnabled, data.FailureThreshold, data.ResetTimeout = r.BreakerEnabled, r.FailureThreshold, r.ResetTimeout
	return nil
}

// pythonModuleName returns a valid Python module name for a backend, e.g. user-service -> user_service
func pythonModuleName(serviceName string) string {
	module := toSnakeCase(serviceName)
	if module == "" || module[0] >= '0' && module[0] <= '9' || pythonReservedNames[module] {
		module = "service_" + module
	}
	return module
}

// renderPythonTemplate renders an embedded Python template
func renderPythonTemplate(name string, data interface{}) (string, error) {
	content, err := templates.TemplateFS.ReadFile(name)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", name, err)
	}
	tmpl, err := template.New(filepath.Base(name)).Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("failed to parse %s: %w", name, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render %s: %w", name, err)
	}
	return buf.String(), nil
}

// Router registration markers in main.py files created before bffgen 3.5
const (
	legacyImportMarker  = "# BFFGEN_ROUTER_IMPORTS - Auto-generated router imports will be inserted here by bffgen generate"
	legacyIncludeMarker = "# BFFGEN_ROUTER_INCLUDES - Auto-generated router includes will be inserted here by bffgen generate"
	legacyMarkerEnd     = "# Do not modify this marker comment"
)

// registerPythonRouters makes sure main.py registers the generated routers. Current projects
// call include_routers(app) from routers/__init__.py and need no change. In older projects
// the lines between each legacy marker and its "Do not modify" line are replaced (not
// appended to), so repeated runs don't duplicate imports.
func registerPythonRouters(mainFile string, write bool) (bool, error) {
	content, err := os.ReadFile(mainFile)
	if err != nil {
		return false, fmt.Errorf("failed to read %s: %w", mainFile, err)
	}
	text := string(content)
	if strings.Contains(text, "include_routers(app)") {
		return false, nil
	}

	updated, okImports := replaceLegacyBlock(text, legacyImportMarker, "from routers import include_routers")
	updated, okIncludes := replaceLegacyBlock(updated, legacyIncludeMarker, "include_routers(app)")
	if !okImports || !okIncludes {
		if !write { // generate checks before it writes; warn once
			fmt.Printf("⚠️  %s doesn't register routers; add these lines to it:\n", mainFile)
			fmt.Println("      from routers import include_routers")
			fmt.Println("      include_routers(app)")
		}
		return false, nil
	}
	if !write {
		return true, nil
	}
	return true, os.WriteFile(mainFile, []byte(updated), utils.ProjectFilePerm)
}

// legacyBffgenIgnore matches the .gitignore line of Python projects made before bffgen 3.7,
// which ignored all of .bffgen/ including the manifest that should be committed
var legacyBffgenIgnore = regexp.MustCompile(`(?m)^/?\.bffgen/?[ \t]*\r?$`)

const bffgenIgnoreRules = `.bffgen/*
!.bffgen/generated.json`

// migratePythonGitignore replaces that line so .bffgen/generated.json is committed. It reports
// whether the file needs (or, with write, got) the change; a missing .gitignore is left alone.
func migratePythonGitignore(path string, write bool) (bool, error) {
	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !legacyBffgenIgnore.Match(content) {
		return false, nil
	}
	if !write {
		return true, nil
	}
	updated := legacyBffgenIgnore.ReplaceAllLiteral(content, []byte(bffgenIgnoreRules))
	if err := os.WriteFile(path, updated, utils.ProjectFilePerm); err != nil {
		return true, err
	}
	fmt.Println("♻️  Updated .gitignore so .bffgen/generated.json is committed")
	return true, nil
}

var settingsIgnoresExtra = regexp.MustCompile(`extra\s*=\s*["']ignore["']`)

// warnStrictPythonSettings warns when config.py comes from a project created before bffgen
// 3.5, whose Settings rejected unknown .env keys: such projects crash at startup on the
// DEBUG=true line init wrote to .env, and on <NAME>_URL entries. config.py belongs to the
// user, so it is not changed.
func warnStrictPythonSettings(configFile string) {
	content, err := os.ReadFile(configFile)
	if err != nil || !strings.Contains(string(content), "BaseSettings") || settingsIgnoresExtra.Match(content) {
		return
	}
	fmt.Printf("⚠️  %s rejects unknown .env keys, so the app fails to start on DEBUG=true or <NAME>_URL\n", configFile)
	fmt.Println(`   lines in .env. Add  extra = "ignore"  to the Settings Config class (or model_config).`)
}

// replaceLegacyBlock replaces the lines between marker and the next legacyMarkerEnd line with block
func replaceLegacyBlock(text, marker, block string) (string, bool) {
	start := strings.Index(text, marker)
	if start == -1 {
		return text, false
	}
	bodyStart := start + len(marker)
	end := strings.Index(text[bodyStart:], legacyMarkerEnd)
	if end == -1 {
		return text, false
	}
	return text[:bodyStart] + "\n" + block + "\n" + text[bodyStart+end:], true
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
