package commands

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/RichGod93/bffgen/v3/internal/scaffolding"
	"github.com/RichGod93/bffgen/v3/internal/templates"
)

func TestNodeHandlerName(t *testing.T) {
	tests := []struct {
		method, path, want string
	}{
		{"GET", "/api/users", "getApiUsers"},
		{"GET", "/api/users/:id", "getApiUsersById"},
		{"PUT", "/api/users/:userId/orders/:orderId", "putApiUsersByUserIdOrdersByOrderId"},
		{"POST", "/api/user-profiles", "postApiUserProfiles"},
		{"DELETE", "/api/items/{id}", "deleteApiItemsById"},
	}

	validIdentifier := regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$]*$`)
	for _, tt := range tests {
		got := nodeHandlerName(tt.method, tt.path)
		if got != tt.want {
			t.Errorf("nodeHandlerName(%q, %q) = %q, want %q", tt.method, tt.path, got, tt.want)
		}
		if !validIdentifier.MatchString(got) {
			t.Errorf("nodeHandlerName(%q, %q) = %q is not a valid JS identifier", tt.method, tt.path, got)
		}
	}
}

func testBackend(endpoints ...map[string]interface{}) map[string]interface{} {
	eps := make([]interface{}, len(endpoints))
	for i, ep := range endpoints {
		eps[i] = ep
	}
	return map[string]interface{}{"name": "users", "baseUrl": "http://localhost:4001/api", "endpoints": eps}
}

func TestRenderNodeJSRouteFile(t *testing.T) {
	backend := testBackend(
		map[string]interface{}{"path": "/users/:id", "method": "GET", "exposeAs": "/api/users/:id", "requiresAuth": true},
		map[string]interface{}{"path": "/orders/{orderId}/items/{itemId}", "method": "get", "exposeAs": "/api/orders/{orderId}/items/{itemId}"},
		map[string]interface{}{"path": "/users", "method": "POST", "exposeAs": "/api/users"},
	)

	for _, framework := range []string{"express", "fastify"} {
		t.Run(framework, func(t *testing.T) {
			content, err := renderNodeJSRouteFile("users", backend, framework)
			if err != nil {
				t.Fatal(err)
			}

			for _, want := range []string{
				"require('../utils/upstream')",
				"envKey: 'USERS'",
				"{ method: 'GET', path: '/users/:id' }",
				// {param} is normalized to :param on both sides
				"'/api/orders/:orderId/items/:itemId'",
				"{ method: 'GET', path: '/orders/:orderId/items/:itemId' }",
				"{ method: 'POST', path: '/users' }",
			} {
				if !strings.Contains(content, want) {
					t.Errorf("route file missing %q\n%s", want, content)
				}
			}
			if strings.Count(content, "authenticate") != map[string]int{"express": 2, "fastify": 1}[framework] {
				t.Errorf("expected authentication on exactly one route\n%s", content)
			}

			again, _ := renderNodeJSRouteFile("users", backend, framework)
			if again != content {
				t.Error("rendering the same backend twice gave different output")
			}
		})
	}
}

// An unused import fails ESLint (no-unused-vars), so the auth middleware is only imported when needed
func TestRenderNodeJSRouteFile_NoAuthImportWithoutAuthRoutes(t *testing.T) {
	backend := testBackend(map[string]interface{}{"path": "/users", "method": "GET", "exposeAs": "/api/users"})
	content, err := renderNodeJSRouteFile("users", backend, "express")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(content, "authenticate") {
		t.Errorf("auth middleware imported although no route requires it\n%s", content)
	}
}

func TestRenderNodeJSRouteFile_RejectsInvalidEndpoints(t *testing.T) {
	tests := map[string]map[string]interface{}{
		"param missing from exposeAs": {"path": "/users/:id", "method": "GET", "exposeAs": "/api/users"},
		"quote in path":               {"path": "/users'", "method": "GET", "exposeAs": "/api/users"},
		"no leading slash":            {"path": "users", "method": "GET", "exposeAs": "/api/users"},
		"unsupported method":          {"path": "/users", "method": "TRACE", "exposeAs": "/api/users"},
	}
	for name, endpoint := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := renderNodeJSRouteFile("users", testBackend(endpoint), "express"); err == nil {
				t.Error("expected an error")
			}
		})
	}
}

func TestNodeEnvKeyAndIdentifier(t *testing.T) {
	tests := []struct{ name, envKey, ident string }{
		{"users", "USERS", "users"},
		{"user-service", "USER_SERVICE", "userService"},
		{"order_items", "ORDER_ITEMS", "orderItems"},
		{"2fa", "2FA", "service2fa"},
	}
	for _, tt := range tests {
		if got := nodeEnvKey(tt.name); got != tt.envKey {
			t.Errorf("nodeEnvKey(%q) = %q, want %q", tt.name, got, tt.envKey)
		}
		if got := nodeIdentifier(tt.name); got != tt.ident {
			t.Errorf("nodeIdentifier(%q) = %q, want %q", tt.name, got, tt.ident)
		}
	}
}

// TestNodeRuntimeJS runs testdata/node_runtime_test.js against freshly rendered runtime files
func TestNodeRuntimeJS(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not installed")
	}

	dir := t.TempDir()
	if _, err := ensureNodeRuntimeFiles(dir, "express", true); err != nil {
		t.Fatal(err)
	}
	// The runtime logs through src/utils/logger.js, which needs winston or pino; stub it
	stub := "module.exports = { debug() {}, info() {}, warn() {}, error() {} };\n"
	if err := os.WriteFile(filepath.Join(dir, "src", "utils", "logger.js"), []byte(stub), 0o600); err != nil {
		t.Fatal(err)
	}

	script, err := filepath.Abs(filepath.Join("testdata", "node_runtime_test.js"))
	if err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(node, script, dir).CombinedOutput()
	if err != nil {
		t.Fatalf("runtime tests failed: %v\n%s", err, out)
	}
}

func TestEnsureNodeRuntimeFiles(t *testing.T) {
	dir := t.TempDir()
	upstream := filepath.Join(dir, "src", "utils", "upstream.js")

	// Dry run reports but writes nothing
	actions, err := ensureNodeRuntimeFiles(dir, "express", false)
	if err != nil || len(actions) == 0 || actions[0].Action != "created" {
		t.Fatalf("dry run: actions=%v err=%v", actions, err)
	}
	if _, err := os.Stat(upstream); !os.IsNotExist(err) {
		t.Fatal("dry run wrote a file")
	}

	// Missing files are created
	if _, err := ensureNodeRuntimeFiles(dir, "express", true); err != nil {
		t.Fatal(err)
	}
	created, err := os.ReadFile(upstream)
	if err != nil {
		t.Fatal(err)
	}
	if runtimeVersionOf(created) != nodeRuntimeVersion {
		t.Fatalf("created file has runtime version %d, want %d", runtimeVersionOf(created), nodeRuntimeVersion)
	}

	// Current files are left alone, local edits included
	edited := append(created, []byte("\n// local edit\n")...)
	if err := os.WriteFile(upstream, edited, 0o600); err != nil {
		t.Fatal(err)
	}
	if actions, _ := ensureNodeRuntimeFiles(dir, "express", true); len(actions) != 0 {
		t.Fatalf("current file was touched: %v", actions)
	}

	// Outdated files are replaced and the old copy is kept
	old := []byte("// bffgen:runtime 0\nmodule.exports = {};\n")
	if err := os.WriteFile(upstream, old, 0o600); err != nil {
		t.Fatal(err)
	}
	actions, err = ensureNodeRuntimeFiles(dir, "express", true)
	if err != nil || len(actions) != 1 || actions[0].Action != "updated" {
		t.Fatalf("outdated file: actions=%v err=%v", actions, err)
	}
	if backup, _ := os.ReadFile(upstream + ".bak"); string(backup) != string(old) {
		t.Error("previous copy not kept as .bak")
	}
	if current, _ := os.ReadFile(upstream); string(current) != string(created) {
		t.Error("outdated file not replaced with the current runtime")
	}
}

func TestNodeInit_WritesESLintConfig(t *testing.T) {
	for name, create := range map[string]func(string, ProjectOptions) error{
		"express": func(p string, o ProjectOptions) error { return createNodeExpressMainFileWithOptions(p, nil, o) },
		"fastify": func(p string, o ProjectOptions) error { return createNodeFastifyMainFileWithOptions(p, nil, o) },
	} {
		t.Run(name, func(t *testing.T) {
			projectName := filepath.Join(t.TempDir(), "app")
			if err := createProjectDirectories(projectName, scaffolding.LanguageNodeExpress); err != nil {
				t.Fatal(err)
			}
			if err := create(projectName, ProjectOptions{SkipTests: true, SkipDocs: true, MiddlewareFlag: "none"}); err != nil {
				t.Fatal(err)
			}

			content, err := os.ReadFile(filepath.Join(projectName, "eslint.config.js"))
			if err != nil {
				t.Fatalf("eslint.config.js not generated: %v", err)
			}
			if !strings.Contains(string(content), "js.configs.recommended") {
				t.Error("eslint.config.js does not extend the recommended config")
			}
		})
	}
}

func TestApplyResilienceConfig(t *testing.T) {
	tests := []struct {
		name    string
		backend map[string]interface{}
		want    templates.RouteTemplateData
		wantErr bool
	}{
		{
			name:    "defaults",
			backend: map[string]interface{}{},
			want:    templates.RouteTemplateData{Timeout: 10000, Retries: 2, BreakerEnabled: true, FailureThreshold: 5, ResetTimeout: 30000},
		},
		{
			name: "explicit values",
			backend: map[string]interface{}{"timeout": 2000.0, "retries": 0.0,
				"circuitBreaker": map[string]interface{}{"enabled": false, "failureThreshold": 3.0, "resetTimeout": "1m"}},
			want: templates.RouteTemplateData{Timeout: 2000, Retries: 0, BreakerEnabled: false, FailureThreshold: 3, ResetTimeout: 60000},
		},
		{
			name:    "duration string timeout",
			backend: map[string]interface{}{"timeout": "1.5s"},
			want:    templates.RouteTemplateData{Timeout: 1500, Retries: 2, BreakerEnabled: true, FailureThreshold: 5, ResetTimeout: 30000},
		},
		{name: "negative retries", backend: map[string]interface{}{"retries": -1.0}, wantErr: true},
		{name: "fractional retries", backend: map[string]interface{}{"retries": 1.5}, wantErr: true},
		{name: "bad duration", backend: map[string]interface{}{"timeout": "soon"}, wantErr: true},
		{name: "zero timeout", backend: map[string]interface{}{"timeout": 0.0}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got templates.RouteTemplateData
			err := applyResilienceConfig(&got, tt.backend)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected an error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

// TestPythonRuntime runs testdata/python_runtime_test.py against freshly written Python runtime
// files. It needs a Python with fastapi, httpx and python-dotenv (BFFGEN_TEST_PYTHON selects
// the interpreter; default python3) and is skipped otherwise.
func TestPythonRuntime(t *testing.T) {
	python := os.Getenv("BFFGEN_TEST_PYTHON")
	if python == "" {
		python = "python3"
	}
	if err := exec.Command(python, "-c", "import fastapi, httpx, dotenv").Run(); err != nil {
		t.Skipf("%s with fastapi, httpx and python-dotenv not available", python)
	}

	dir := t.TempDir()
	if _, err := ensurePythonRuntimeFiles(dir, true); err != nil {
		t.Fatal(err)
	}
	for _, pkg := range []string{"utils", "services"} {
		if err := os.WriteFile(filepath.Join(dir, pkg, "__init__.py"), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	script, err := filepath.Abs(filepath.Join("testdata", "python_runtime_test.py"))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(python, "-I", script, dir)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("python runtime tests failed: %v\n%s", err, out)
	}
}
