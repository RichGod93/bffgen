package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
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
		if err := registerPythonRouters(path); err != nil {
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

	dryRun = true
	defer func() { dryRun = false }()
	if err := generatePython(); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
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
