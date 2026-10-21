package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSanitizeFunctionName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"get_users", "get_users"},
		{"get-users", "get_users"},
		{"get users", "get_users"},
		{"getUsers", "get_users"},
		{"GetUsers", "get_users"},
		{"get/users", "get_users"},
		{"get.users", "get_users"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// sanitizeFunctionName requires name, path, method - passing name with empty path/method
			result := sanitizeFunctionName(tt.input, "", "")
			if result != tt.expected {
				t.Errorf("sanitizeFunctionName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user_service", "UserService"},
		{"product_api", "ProductApi"},
		{"simple", "Simple"},
		{"", ""},
		{"user", "User"},
		{"my_long_service_name", "MyLongServiceName"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toCamelCase(tt.input)
			if result != tt.expected {
				t.Errorf("toCamelCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGeneratePythonWithConfig(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := filepath.Join(tmpDir, "test-python-project")

	// Create project structure
	err := os.MkdirAll(projectName, 0755)
	if err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	err = createPythonDirectories(projectName)
	if err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	// Create main.py
	opts := ProjectOptions{AsyncEndpoints: true}
	err = createFastAPIMainFile(projectName, opts)
	if err != nil {
		t.Fatalf("Failed to create main.py: %v", err)
	}

	// Create config file
	config := map[string]interface{}{
		"project": map[string]interface{}{
			"name":      "test-project",
			"framework": "fastapi",
			"async":     true,
		},
		"backends": []map[string]interface{}{
			{
				"name":    "api",
				"baseUrl": "http://localhost:5000",
				"endpoints": []map[string]interface{}{
					{
						"name":         "health",
						"method":       "GET",
						"path":         "/api/health",
						"upstreamPath": "/health",
					},
				},
			},
		},
	}

	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	configPath := filepath.Join(projectName, "bffgen.config.py.json")
	err = os.WriteFile(configPath, configData, 0644)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Change to project directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(projectName)

	// Generate code
	err = generatePython()
	if err != nil {
		t.Fatalf("generatePython() failed: %v", err)
	}

	// Verify router was created
	routerPath := filepath.Join(projectName, "routers", "api_router.py")
	if _, err := os.Stat(routerPath); os.IsNotExist(err) {
		t.Error("Router was not generated")
	}

	// Verify service was created
	servicePath := filepath.Join(projectName, "services", "api_service.py")
	if _, err := os.Stat(servicePath); os.IsNotExist(err) {
		t.Error("Service was not generated")
	}

	// Verify the router is registered
	registry, err := os.ReadFile(filepath.Join(projectName, "routers", "__init__.py"))
	if err != nil {
		t.Fatalf("Failed to read routers/__init__.py: %v", err)
	}
	if !contains(string(registry), "from routers.api_router import router as api_router") {
		t.Error("routers/__init__.py should import api_router")
	}
	if !contains(string(registry), "ROUTERS: List[APIRouter] = [api_router]") {
		t.Error("routers/__init__.py should list api_router")
	}
}

func pyBackend(endpoints ...map[string]interface{}) map[string]interface{} {
	eps := make([]interface{}, len(endpoints))
	for i, ep := range endpoints {
		eps[i] = ep
	}
	return map[string]interface{}{"name": "user-service", "baseUrl": "http://localhost:4001/api", "endpoints": eps}
}

func TestBuildPythonBackend(t *testing.T) {
	data, err := buildPythonBackend(pyBackend(
		map[string]interface{}{"name": "getUser", "method": "get", "path": "/api/users/{id}", "upstreamPath": "/users/{id}", "requiresAuth": true},
		// Node-style endpoint with :params
		map[string]interface{}{"method": "GET", "exposeAs": "/api/orders/:orderId/items/:itemId", "path": "/orders/:orderId/items/:itemId"},
		// upstreamPath defaults to path
		map[string]interface{}{"method": "POST", "path": "/api/users"},
		// duplicate name gets a suffix
		map[string]interface{}{"name": "getUser", "method": "DELETE", "path": "/api/users/{id}", "upstreamPath": "/users/{id}"},
	))
	if err != nil {
		t.Fatal(err)
	}

	if data.Module != "user_service" || data.EnvKey != "USER_SERVICE" || data.ClassName != "UserServiceService" {
		t.Errorf("names: module=%q envKey=%q class=%q", data.Module, data.EnvKey, data.ClassName)
	}
	if !data.AnyAuth || data.Timeout != 10000 || data.Retries != 2 || !data.BreakerEnabled {
		t.Errorf("defaults: %+v", data)
	}

	want := []pythonEndpointData{
		{Method: "GET", Path: "/api/users/{id}", BackendPath: "/users/{id}", Params: []string{"id"}, HandlerName: "get_user", RequiresAuth: true},
		{Method: "GET", Path: "/api/orders/{orderId}/items/{itemId}", BackendPath: "/orders/{orderId}/items/{itemId}", Params: []string{"orderId", "itemId"}, HandlerName: "get_api_orders_by_order_id_items_by_item_id"},
		{Method: "POST", Path: "/api/users", BackendPath: "/api/users", HandlerName: "post_api_users", HasBody: true},
		{Method: "DELETE", Path: "/api/users/{id}", BackendPath: "/users/{id}", Params: []string{"id"}, HandlerName: "get_user_2", HasBody: true},
	}
	if !reflect.DeepEqual(data.Endpoints, want) {
		t.Errorf("endpoints:\n got %+v\nwant %+v", data.Endpoints, want)
	}
}

func TestBuildPythonBackend_Config(t *testing.T) {
	ep := map[string]interface{}{"method": "GET", "path": "/api/users"}

	b := pyBackend(ep)
	delete(b, "baseUrl")
	b["url"] = "http://legacy:5000"
	data, err := buildPythonBackend(b)
	if err != nil || data.BaseURL != "http://legacy:5000" {
		t.Errorf("legacy url key: %v %+v", err, data)
	}

	b = pyBackend(ep)
	b["timeout"] = "5s"
	b["circuitBreaker"] = map[string]interface{}{"failureThreshold": 3.0, "recoveryTimeout": 60000.0}
	data, err = buildPythonBackend(b)
	if err != nil || data.Timeout != 5000 || data.FailureThreshold != 3 || data.ResetTimeout != 60000 {
		t.Errorf("duration string / recoveryTimeout alias: %v %+v", err, data)
	}

	if data, err := buildPythonBackend(map[string]interface{}{"name": "x", "baseUrl": "http://x", "endpoints": []interface{}{}}); data != nil || err != nil {
		t.Errorf("backend without endpoints should be skipped: %v %v", data, err)
	}
}

func TestBuildPythonBackend_Rejects(t *testing.T) {
	tests := map[string]map[string]interface{}{
		"seconds as timeout":      {"timeout": 30.0},
		"seconds as resetTimeout": {"circuitBreaker": map[string]interface{}{"resetTimeout": 60.0}},
		"no base URL":             {"baseUrl": ""},
	}
	for name, overrides := range tests {
		t.Run(name, func(t *testing.T) {
			b := pyBackend(map[string]interface{}{"method": "GET", "path": "/api/users"})
			for k, v := range overrides {
				b[k] = v
			}
			if _, err := buildPythonBackend(b); err == nil {
				t.Error("expected an error")
			}
		})
	}

	endpoints := map[string]map[string]interface{}{
		"keyword param":      {"method": "GET", "path": "/api/items/{class}", "upstreamPath": "/items/{class}"},
		"reserved param":     {"method": "GET", "path": "/api/items/{request}", "upstreamPath": "/items/{request}"},
		"param not exposed":  {"method": "GET", "path": "/api/items", "upstreamPath": "/items/{id}"},
		"unsupported method": {"method": "TRACE", "path": "/api/items"},
		"quote in path":      {"method": "GET", "path": "/api/items\""},
	}
	for name, ep := range endpoints {
		t.Run(name, func(t *testing.T) {
			if _, err := buildPythonBackend(pyBackend(ep)); err == nil {
				t.Error("expected an error")
			}
		})
	}
}

// main.py files from before the router registry had marker comments that generate appended to;
// they are now rewritten in place, so repeated runs don't duplicate imports
func TestRegisterPythonRouters_Legacy(t *testing.T) {
	legacy := "app = FastAPI()\n\n" + legacyImportMarker + "\nfrom routers.a_router import router as a_router\nfrom routers.a_router import router as a_router\n" + legacyMarkerEnd +
		"\n\n" + legacyIncludeMarker + "\napp.include_router(a_router)\n" + legacyMarkerEnd + "\n"
	path := filepath.Join(t.TempDir(), "main.py")
	if err := os.WriteFile(path, []byte(legacy), 0o600); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 2; i++ {
		if _, err := registerPythonRouters(path, true); err != nil {
			t.Fatal(err)
		}
	}
	content, _ := os.ReadFile(path)
	want := "app = FastAPI()\n\n" + legacyImportMarker + "\nfrom routers import include_routers\n" + legacyMarkerEnd +
		"\n\n" + legacyIncludeMarker + "\ninclude_routers(app)\n" + legacyMarkerEnd + "\n"
	if string(content) != want {
		t.Errorf("got:\n%s\nwant:\n%s", content, want)
	}
}

func TestGeneratePython_DryRunWritesNothing(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	config := `{"backends": [{"name": "api", "baseUrl": "http://localhost:5000", "endpoints": [{"method": "GET", "path": "/api/x"}]}]}`
	if err := os.WriteFile(pythonConfigFile, []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("main.py", []byte("include_routers(app)\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	dryRun = true
	defer func() { dryRun = false }()
	if err := generatePython(); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 2 {
		t.Errorf("dry run wrote files: %v", entries)
	}
}

func TestGeneratePython_Deterministic(t *testing.T) {
	config := `{"backends": [
		{"name": "users", "baseUrl": "http://localhost:4001", "endpoints": [{"method": "GET", "path": "/api/users/{id}", "upstreamPath": "/users/{id}", "requiresAuth": true}]},
		{"name": "orders", "baseUrl": "http://localhost:4002", "endpoints": [{"method": "POST", "path": "/api/orders"}]}]}`

	snapshot := func() map[string]string {
		dir := t.TempDir()
		t.Chdir(dir)
		if err := os.WriteFile("main.py", []byte("include_routers(app)\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(pythonConfigFile, []byte(config), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := generatePython(); err != nil {
			t.Fatal(err)
		}
		files := map[string]string{}
		_ = filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
			if err == nil && !d.IsDir() {
				content, _ := os.ReadFile(path)
				files[path] = string(content)
			}
			return nil
		})
		return files
	}

	first, second := snapshot(), snapshot()
	if !reflect.DeepEqual(first, second) {
		t.Error("generating twice from the same config gave different output")
	}
	if len(first) < 8 {
		t.Errorf("expected routers, services, registry and runtime files, got %d files", len(first))
	}
}

// Note: contains helper function is defined in init_python_test.go and shared across test files

// pythonTestProject writes a config and main.py into a new directory and changes into it
func pythonTestProject(t *testing.T, config, mainPy string) string {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile(pythonConfigFile, []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("main.py", []byte(mainPy), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

const pythonTestConfig = `{"backends": [{"name": "api", "baseUrl": "http://localhost:5000", "endpoints": [{"method": "GET", "path": "/api/x"}]}]}`

func TestGeneratePython_RejectsCollidingBackendNames(t *testing.T) {
	config := `{"backends": [
		{"name": "user-service", "baseUrl": "http://localhost:4001", "endpoints": [{"method": "GET", "path": "/a"}]},
		{"name": "user_service", "baseUrl": "http://localhost:4002", "endpoints": [{"method": "GET", "path": "/b"}]}]}`
	dir := pythonTestProject(t, config, "include_routers(app)\n")
	err := generatePython()
	if err == nil || !strings.Contains(err.Error(), "user_service") {
		t.Errorf("backends generating the same module accepted: %v", err)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 2 {
		t.Errorf("rejected config still wrote: %v", entries)
	}
}

func TestGeneratePython_CorruptManifestWritesNothing(t *testing.T) {
	dir := pythonTestProject(t, pythonTestConfig, "include_routers(app)\n")
	if err := os.MkdirAll(".bffgen", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(generatedManifestPath, []byte("<<<<<<< HEAD\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := generatePython(); err == nil {
		t.Error("corrupt manifest accepted")
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 3 {
		t.Errorf("failed generate still wrote: %v", entries)
	}
}

func TestGeneratePython_CheckIncludesMainPy(t *testing.T) {
	pythonTestProject(t, pythonTestConfig, "include_routers(app)\n")
	if err := generatePython(); err != nil {
		t.Fatal(err)
	}
	checkMode = true
	defer func() { checkMode = false }()
	if err := generatePython(); err != nil {
		t.Fatalf("--check failed on an up-to-date project: %v", err)
	}

	// main.py from an older project: generate would rewrite the router registration
	legacy := "app = FastAPI()\n" + legacyImportMarker + "\nfrom routers.a_router import router\n" + legacyMarkerEnd +
		"\n" + legacyIncludeMarker + "\napp.include_router(router)\n" + legacyMarkerEnd + "\n"
	if err := os.WriteFile("main.py", []byte(legacy), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := generatePython(); err == nil {
		t.Error("--check passed although main.py would change")
	}
	if content, _ := os.ReadFile("main.py"); string(content) != legacy {
		t.Error("--check wrote main.py")
	}
}

func TestGeneratePython_MigratesLegacyGitignore(t *testing.T) {
	pythonTestProject(t, pythonTestConfig, "include_routers(app)\n")
	legacy := "__pycache__/\n\n# bffgen\n.bffgen/\n"
	if err := os.WriteFile(".gitignore", []byte(legacy), 0o600); err != nil {
		t.Fatal(err)
	}

	dryRun = true
	if err := generatePython(); err != nil {
		t.Fatal(err)
	}
	dryRun = false
	if content, _ := os.ReadFile(".gitignore"); string(content) != legacy {
		t.Fatal("dry run changed .gitignore")
	}

	for i := 0; i < 2; i++ { // idempotent
		if err := generatePython(); err != nil {
			t.Fatal(err)
		}
	}
	want := "__pycache__/\n\n# bffgen\n.bffgen/*\n!.bffgen/generated.json\n"
	if content, _ := os.ReadFile(".gitignore"); string(content) != want {
		t.Errorf(".gitignore = %q, want %q", content, want)
	}
}

func TestPythonRouter_Roles(t *testing.T) {
	data, err := buildPythonBackend(pyBackend(
		map[string]interface{}{"method": "GET", "path": "/api/users", "requiresAuth": true},
		map[string]interface{}{"method": "DELETE", "path": "/api/users/{id}", "roles": []interface{}{"admin", "support"}},
	))
	if err != nil {
		t.Fatal(err)
	}
	router, err := renderPythonTemplate("python/fastapi/router_template.py.tmpl", data)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"from fastapi import APIRouter, Depends, Request, Response",
		"from middleware.auth import require_authentication\nfrom utils.authorize import require_roles\n",
		`dependencies=[Depends(require_roles("admin", "support"))],`,
		"dependencies=[Depends(require_authentication)],",
	} {
		if !strings.Contains(router, want) {
			t.Errorf("router missing %q\n%s", want, router)
		}
	}

	// Only roles: require_authentication isn't imported (ruff flags unused imports)
	data, _ = buildPythonBackend(pyBackend(map[string]interface{}{"method": "GET", "path": "/api/x", "roles": []interface{}{"admin"}}))
	router, _ = renderPythonTemplate("python/fastapi/router_template.py.tmpl", data)
	if strings.Contains(router, "require_authentication") || !strings.Contains(router, "from utils.authorize import require_roles") {
		t.Errorf("roles-only router imports\n%s", router)
	}
}

func TestValidatePythonConfig(t *testing.T) {
	valid := `{"backends": [
		{"name": "users", "baseUrl": "http://localhost:4001", "endpoints": [
			{"method": "GET", "path": "/api/users/{id}", "requiresAuth": true},
			{"method": "DELETE", "path": "/api/users/{id}", "roles": ["admin"]}]},
		{"name": "orders", "baseUrl": "http://localhost:4002", "endpoints": [{"method": "GET", "path": "/api/orders"}]}]}`
	dir := pythonTestProject(t, valid, "include_routers(app)\n")
	if got := detectProjectType(); got != "python" {
		t.Fatalf("detectProjectType = %q, want python", got)
	}
	if err := validateProjectConfig(); err != nil {
		t.Errorf("valid Python config rejected: %v", err)
	}

	for name, tc := range map[string]struct{ config, want string }{
		"no baseUrl":      {`{"backends": [{"name": "a", "endpoints": [{"method": "GET", "path": "/x"}]}]}`, "baseUrl is required"},
		"invalid roles":   {`{"backends": [{"name": "a", "baseUrl": "http://h", "endpoints": [{"method": "GET", "path": "/x", "roles": "admin"}]}]}`, "roles"},
		"bad method":      {`{"backends": [{"name": "a", "baseUrl": "http://h", "endpoints": [{"method": "TRACE", "path": "/x"}]}]}`, "unsupported method"},
		"colliding names": {`{"backends": [{"name": "a-b", "baseUrl": "http://h", "endpoints": [{"method": "GET", "path": "/x"}]}, {"name": "a_b", "baseUrl": "http://h", "endpoints": [{"method": "GET", "path": "/y"}]}]}`, "a_b"},
		"no endpoints":    {`{"backends": [{"name": "a", "baseUrl": "http://h", "endpoints": []}]}`, "no endpoints"},
		"no backends":     {`{"project": {}}`, "backends"},
	} {
		t.Run(name, func(t *testing.T) {
			var config map[string]interface{}
			if err := json.Unmarshal([]byte(tc.config), &config); err != nil {
				t.Fatal(err)
			}
			errs, _, _ := validatePythonConfigStructure(config)
			if len(errs) == 0 || !strings.Contains(strings.Join(errs, "\n"), tc.want) {
				t.Errorf("got %v, want an error containing %q", errs, tc.want)
			}
		})
	}

	// A config generate rejects is rejected by validate too
	if err := os.WriteFile(filepath.Join(dir, pythonConfigFile), []byte(`{"backends": [{"name": "a", "endpoints": [{"method": "GET", "path": "/x"}]}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := validateProjectConfig(); err == nil {
		t.Error("validate accepted a config without baseUrl")
	}
	if err := generatePython(); err == nil {
		t.Error("generate accepted a config without baseUrl")
	}
}

func TestGeneratePython_KeepsUserAuthorizeFile(t *testing.T) {
	dir := pythonTestProject(t, pythonTestConfig, "include_routers(app)\n")
	own := "def require_admin():\n    pass\n"
	if err := os.MkdirAll("utils", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join("utils", "authorize.py"), []byte(own), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := generatePython(); err != nil {
		t.Fatal(err)
	}
	if content, _ := os.ReadFile(filepath.Join("utils", "authorize.py")); string(content) != own {
		t.Fatal("user's utils/authorize.py was replaced")
	}

	roles := strings.Replace(pythonTestConfig, `"path": "/api/x"`, `"path": "/api/x", "roles": ["admin"]`, 1)
	if err := os.WriteFile(pythonConfigFile, []byte(roles), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := generatePython(); err == nil || !strings.Contains(err.Error(), "authorize.py") {
		t.Errorf("roles with the user's authorize.py in the way: err = %v", err)
	}
	if content, _ := os.ReadFile(filepath.Join(dir, "routers", "api_router.py")); strings.Contains(string(content), "require_roles") {
		t.Error("router with roles written although generate failed")
	}
}
