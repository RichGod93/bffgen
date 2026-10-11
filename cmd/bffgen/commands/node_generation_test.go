package commands

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/RichGod93/bffgen/internal/scaffolding"
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

// fetch() throws "Request with GET/HEAD method cannot have body", so generated
// GET/HEAD routes must never forward the request body.
func TestGenerateNodeJSRouteFile_NoBodyOnGet(t *testing.T) {
	for _, framework := range []string{"express", "fastify"} {
		t.Run(framework, func(t *testing.T) {
			t.Chdir(t.TempDir())

			backend := map[string]interface{}{
				"name":    "users",
				"baseUrl": "http://localhost:4001/api",
				"endpoints": []interface{}{
					map[string]interface{}{"path": "/users/:id", "method": "GET", "exposeAs": "/api/users/:id"},
					map[string]interface{}{"path": "/users", "method": "POST", "exposeAs": "/api/users"},
				},
			}
			if err := generateNodeJSRouteFile("users", backend, framework); err != nil {
				t.Fatal(err)
			}

			content, err := os.ReadFile(filepath.Join("src", "routes", "users.js"))
			if err != nil {
				t.Fatal(err)
			}

			// Split per route registration; only the POST route may forward a body
			routes := regexp.MustCompile(`(?m)^\s*(?:router|fastify)\.(get|post)\(`).Split(string(content), -1)[1:]
			if len(routes) != 2 {
				t.Fatalf("expected 2 routes, got %d", len(routes))
			}
			if strings.Contains(routes[0], "body: JSON.stringify") {
				t.Error("GET route forwards a request body")
			}
			if !strings.Contains(routes[1], "body: JSON.stringify") {
				t.Error("POST route does not forward the request body")
			}
		})
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
