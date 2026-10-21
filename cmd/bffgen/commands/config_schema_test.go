package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xeipuuv/gojsonschema"
)

// validateAgainstConfigSchema validates a JSON document against schemas/bffgen-config.json
func validateAgainstConfigSchema(t *testing.T, doc string) []string {
	t.Helper()
	schemaPath, err := filepath.Abs(filepath.Join("..", "..", "..", "schemas", "bffgen-config.json"))
	if err != nil {
		t.Fatal(err)
	}
	result, err := gojsonschema.Validate(gojsonschema.NewReferenceLoader("file://"+filepath.ToSlash(schemaPath)), gojsonschema.NewStringLoader(doc))
	if err != nil {
		t.Fatal(err)
	}
	var errs []string
	for _, e := range result.Errors() {
		errs = append(errs, e.String())
	}
	return errs
}

func TestConfigSchema_AcceptsShippedConfigs(t *testing.T) {
	docs := map[string]string{}
	for _, example := range []string{"express-example.json", "fastify-example.json", "python-example.json"} {
		data, err := os.ReadFile(filepath.Join("..", "..", "..", "examples", example))
		if err != nil {
			t.Fatal(err)
		}
		docs[example] = string(data)
	}
	// What init writes, and configs with the backends the tests use
	for _, framework := range []string{"express", "fastify"} {
		init := readFile(t, filepath.Join(nodeInitProject(t, framework), "bffgen.config.json"))
		if !strings.Contains(init, `"$schema": "https://raw.githubusercontent.com/RichGod93/bffgen/master/schemas/bffgen-config.json"`) {
			t.Errorf("init %s: config doesn't reference the schema", framework)
		}
		docs["init "+framework] = init
		dir := nodeTestProject(t, framework)
		addRoles(t, dir)
		docs["test project "+framework] = readFile(t, filepath.Join(dir, "bffgen.config.json"))
	}
	pyDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(pyDir, "app"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := inDir(t, pyDir, false, false, false, func() error {
		return createPythonBFFGenConfig("app", ProjectOptions{AsyncEndpoints: true})
	}); err != nil {
		t.Fatal(err)
	}
	docs["init python"] = readFile(t, filepath.Join(pyDir, "app", pythonConfigFile))
	docs["python with roles"] = `{"backends": [{"name": "users", "baseUrl": "http://localhost:4001", "timeout": "5s",
		"circuitBreaker": {"recoveryTimeout": 30000},
		"endpoints": [{"method": "DELETE", "path": "/api/users/{id}", "upstreamPath": "/users/{id}", "roles": ["admin", "support"]}]}]}`

	for name, doc := range docs {
		if errs := validateAgainstConfigSchema(t, doc); len(errs) > 0 {
			t.Errorf("%s doesn't match the schema:\n%s", name, strings.Join(errs, "\n"))
		}
	}
}

func TestConfigSchema_RejectsInvalidConfigs(t *testing.T) {
	endpoint := func(extra string) string {
		return `{"backends": [{"name": "a", "baseUrl": "http://h", "endpoints": [{"method": "GET", "path": "/x"` + extra + `}]}]}`
	}
	for name, doc := range map[string]string{
		"roles not a list":   endpoint(`, "roles": "admin"`),
		"empty roles":        endpoint(`, "roles": []`),
		"role with a space":  endpoint(`, "roles": ["super admin"]`),
		"unsupported method": `{"backends": [{"name": "a", "endpoints": [{"method": "TRACE", "path": "/x"}]}]}`,
		"no leading slash":   `{"backends": [{"name": "a", "endpoints": [{"method": "GET", "path": "x"}]}]}`,
		"negative retries":   `{"backends": [{"name": "a", "retries": -1, "endpoints": []}]}`,
		"bad duration":       `{"backends": [{"name": "a", "timeout": "soon", "endpoints": []}]}`,
		"no backend name":    `{"backends": [{"endpoints": []}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			if errs := validateAgainstConfigSchema(t, doc); len(errs) == 0 {
				t.Error("schema accepted an invalid config")
			}
		})
	}
}
