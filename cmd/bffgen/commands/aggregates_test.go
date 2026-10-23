package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// aggregateConfig is a Node.js config with backends users, orders and products and the given aggregates (JSON)
func aggregateConfig(t *testing.T, aggregates string) map[string]interface{} {
	t.Helper()
	var config map[string]interface{}
	err := json.Unmarshal([]byte(`{
		"project": {"name": "app", "framework": "express"},
		"backends": [
			{"name": "users", "baseUrl": "http://localhost:4001/api", "timeout": 2000, "endpoints": [
				{"method": "GET", "path": "/users/:id", "exposeAs": "/api/users/:id"}]},
			{"name": "orders", "baseUrl": "http://localhost:4002/api", "endpoints": [
				{"method": "POST", "path": "/orders", "exposeAs": "/api/dashboard/:id"}]},
			{"name": "products", "baseUrl": "http://localhost:4003/api"}
		],
		"aggregates": `+aggregates+`}`), &config)
	if err != nil {
		t.Fatal(err)
	}
	return config
}

const dashboardAggregate = `[{"name": "dashboard", "exposeAs": "/api/dashboard/{id}", "roles": ["user"], "sources": [
	{"backend": "users", "path": "/users/{id}", "as": "user", "required": true},
	{"backend": "orders", "path": "/orders?userId=:id&sort=-date"},
	{"backend": "products", "path": "/products", "as": "recommendations"}]}]`

func TestParseAggregates(t *testing.T) {
	specs, err := parseAggregates(aggregateConfig(t, dashboardAggregate), false)
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 1 {
		t.Fatalf("got %d aggregates, want 1", len(specs))
	}
	a := specs[0]
	if a.Path != "/api/dashboard/:id" || !a.RequiresAuth || len(a.Roles) != 1 {
		t.Errorf("aggregate = %+v", a)
	}
	want := []aggregateSource{
		{Backend: "users", Path: "/users/:id", As: "user", Required: true},
		{Backend: "orders", Path: "/orders?userId=:id&sort=-date", As: "orders"},
		{Backend: "products", Path: "/products", As: "recommendations"},
	}
	for i, src := range a.Sources {
		if src != want[i] {
			t.Errorf("source %d = %+v, want %+v", i, src, want[i])
		}
	}
	if got := aggregateBackends(specs); strings.Join(got, ",") != "users,orders,products" {
		t.Errorf("aggregateBackends = %v", got)
	}
}

func TestParseAggregates_Rejects(t *testing.T) {
	src := `{"backend": "users", "path": "/users"}`
	tests := map[string]struct{ aggregates, want string }{
		"not a list":          {`{}`, "must be a list"},
		"no exposeAs":         {`[{"sources": [` + src + `]}]`, "exposeAs"},
		"no sources":          {`[{"exposeAs": "/api/a", "sources": []}]`, "sources"},
		"unknown backend":     {`[{"exposeAs": "/api/a", "sources": [{"backend": "nope", "path": "/x"}]}]`, "not configured"},
		"missing backend":     {`[{"exposeAs": "/api/a", "sources": [{"path": "/x"}]}]`, `"backend" is required`},
		"duplicate as":        {`[{"exposeAs": "/api/a", "sources": [` + src + `, ` + src + `]}]`, "used twice"},
		"invalid as":          {`[{"exposeAs": "/api/a", "sources": [{"backend": "users", "path": "/x", "as": "a b"}]}]`, `"as"`},
		"undefined param":     {`[{"exposeAs": "/api/a", "sources": [{"backend": "users", "path": "/users/:id"}]}]`, "does not define"},
		"quote in path":       {`[{"exposeAs": "/api/a", "sources": [{"backend": "users", "path": "/users'"}]}]`, "URL characters"},
		"quote in exposeAs":   {`[{"exposeAs": "/api/a'", "sources": [` + src + `]}]`, "URL path characters"},
		"query in exposeAs":   {`[{"exposeAs": "/api/a?x=1", "sources": [` + src + `]}]`, "URL path characters"},
		"not GET":             {`[{"exposeAs": "/api/a", "method": "POST", "sources": [` + src + `]}]`, "GET routes"},
		"required not bool":   {`[{"exposeAs": "/api/a", "sources": [{"backend": "users", "path": "/x", "required": "yes"}]}]`, "true or false"},
		"invalid roles":       {`[{"exposeAs": "/api/a", "roles": "admin", "sources": [` + src + `]}]`, "roles"},
		"invalid name":        {`[{"name": "../x", "exposeAs": "/api/a", "sources": [` + src + `]}]`, "name"},
		"clashes w/ GET":      {`[{"exposeAs": "/api/users/:userId", "sources": [` + src + `]}]`, "the users endpoint GET /api/users/:id"},
		"duplicate aggregate": {`[{"exposeAs": "/api/a/:x", "sources": [` + src + `]}, {"name": "b", "exposeAs": "/api/a/{y}", "sources": [` + src + `]}]`, `aggregate "b": GET /api/a/:y is already defined`},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := parseAggregates(aggregateConfig(t, tt.aggregates), false)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("got %v, want an error containing %q", err, tt.want)
			}
		})
	}

	// A POST endpoint on the same path doesn't clash with a GET aggregate
	if _, err := parseAggregates(aggregateConfig(t, dashboardAggregate), false); err != nil {
		t.Errorf("aggregate next to a POST endpoint rejected: %v", err)
	}
	// Python path parameters are function arguments, so reserved names are rejected
	if _, err := parseAggregates(aggregateConfig(t, `[{"exposeAs": "/api/a/{request}", "sources": [`+src+`]}]`), true); err == nil {
		t.Error("reserved Python parameter name accepted")
	}
}

func TestRenderNodeAggregates(t *testing.T) {
	config := aggregateConfig(t, dashboardAggregate)
	specs, err := parseAggregates(config, false)
	if err != nil {
		t.Fatal(err)
	}
	for framework, wants := range map[string][]string{
		"express": {
			"const { authenticate } = require('../middleware/auth');",
			"router.get('/api/dashboard/:id', authenticate, authorize(['user']), (req, res, next) =>",
			"]).catch(next));",
		},
		"fastify": {
			"onRequest: [authenticate],",
			"preHandler: [authorize(['user'])],",
			"fastify.get('/api/dashboard/:id', {",
			"module.exports = aggregateRoutes;",
		},
	} {
		t.Run(framework, func(t *testing.T) {
			out, err := renderNodeAggregates(config, specs, framework)
			if err != nil {
				t.Fatal(err)
			}
			wants = append(wants,
				"const { createUpstream, aggregate } = require('../utils/upstream');",
				"const usersBackend = createUpstream('users', {",
				"timeout: 2000,",
				// products has no endpoints, so it gets its upstream here only
				"const productsBackend = createUpstream('products', {",
				"{ as: 'user', upstream: usersBackend, path: '/users/:id', required: true },",
				// validated paths are written as is, without = escapes
				"{ as: 'orders', upstream: ordersBackend, path: '/orders?userId=:id&sort=-date', required: false },",
				"{ as: 'recommendations', upstream: productsBackend, path: '/products', required: false },",
			)
			for _, want := range wants {
				if !strings.Contains(out, want) {
					t.Errorf("missing %q in\n%s", want, out)
				}
			}
		})
	}
}

func TestNodeGenerate_Aggregates(t *testing.T) {
	for _, framework := range []string{"express", "fastify"} {
		t.Run(framework, func(t *testing.T) {
			dir := nodeTestProject(t, framework)
			if err := inDir(t, dir, false, false, false, generateNodeJS); err != nil {
				t.Fatal(err)
			}
			routes := readFile(t, filepath.Join(dir, "src", "routes", "aggregates.js"))
			if !strings.Contains(routes, "'/api/dashboard/:id'") {
				t.Errorf("aggregates.js lacks the dashboard route:\n%s", routes)
			}
			index := readFile(t, filepath.Join(dir, "src", "index.js"))
			if !strings.Contains(index, "require('./routes/aggregates')") {
				t.Error("aggregates routes not registered in src/index.js")
			}
			if !strings.Contains(readFile(t, filepath.Join(dir, "src", "utils", "upstream.js")), "async function aggregate(") {
				t.Error("upstream.js runtime lacks aggregate()")
			}
		})
	}

	// A backend named "aggregates" would share the route file
	dir := nodeTestProject(t, "express")
	config := strings.Replace(readFile(t, filepath.Join(dir, "bffgen.config.json")), `"orders"`, `"aggregates"`, -1)
	writeFile(t, filepath.Join(dir, "bffgen.config.json"), config)
	if err := inDir(t, dir, false, false, false, generateNodeJS); err == nil || !strings.Contains(err.Error(), "reserved") {
		t.Errorf("backend named aggregates accepted: %v", err)
	}
}

func TestAggregatesInValidateAndDocs(t *testing.T) {
	config := aggregateConfig(t, dashboardAggregate)
	// products has no endpoints but is used by an aggregate, so it's fine
	if errs := validateNodeJSConfigStructure(config); len(errs) != 0 {
		t.Errorf("valid config with aggregates rejected: %v", errs)
	}
	config = aggregateConfig(t, `[{"exposeAs": "/api/a", "sources": [{"backend": "nope", "path": "/x"}]}]`)
	if errs := validateNodeJSConfigStructure(config); !strings.Contains(strings.Join(errs, "\n"), "not configured") {
		t.Errorf("unknown aggregate backend not reported: %v", errs)
	}

	config = aggregateConfig(t, dashboardAggregate)
	specs, _ := parseAggregates(config, false)
	spec := generateOpenAPISpec("app", config["backends"].([]interface{}))
	addAggregateDocs(spec, specs)
	get, ok := spec["paths"].(map[string]interface{})["/api/dashboard/:id"].(map[string]interface{})["get"].(map[string]interface{})
	if !ok {
		t.Fatal("aggregate not documented")
	}
	responses := get["responses"].(map[string]interface{})
	for _, code := range []string{"200", "401", "403", "502", "503", "504"} {
		if responses[code] == nil {
			t.Errorf("aggregate docs lack a %s response", code)
		}
	}
	// The POST endpoint on the same path is kept
	if spec["paths"].(map[string]interface{})["/api/dashboard/:id"].(map[string]interface{})["post"] == nil {
		t.Error("aggregate docs replaced the POST endpoint on the same path")
	}
}

func TestPythonAggregates(t *testing.T) {
	config := `{"backends": [
		{"name": "users", "baseUrl": "http://localhost:4001", "endpoints": [{"method": "GET", "path": "/api/users/{id}"}]},
		{"name": "order-service", "baseUrl": "http://localhost:4002"}],
		"aggregates": [
			{"name": "dashboard", "path": "/api/dashboard/:id", "requiresAuth": true, "sources": [
				{"backend": "users", "path": "/users/:id", "as": "user", "required": true},
				{"backend": "order-service", "path": "/orders?userId={id}", "as": "orders"}]},
			{"exposeAs": "/api/admin/summary", "roles": ["admin"], "sources": [{"backend": "users", "path": "/users"}]}]}`
	dir := pythonTestProject(t, config, "include_routers(app)\n")
	if errs, _, _ := validatePythonConfigStructure(mustJSON(t, config)); len(errs) != 0 {
		t.Errorf("valid config rejected by validate: %v", errs)
	}
	if err := generatePython(); err != nil {
		t.Fatal(err)
	}
	router := readFile(t, filepath.Join(dir, "routers", "aggregates_router.py"))
	for _, want := range []string{
		"from middleware.auth import require_authentication\nfrom utils.authorize import require_roles\n",
		"from utils.upstream import Source, aggregate, create_upstream",
		`order_service_backend = create_upstream(`,
		`"/api/dashboard/{id}",`,
		"async def aggregate_dashboard(request: Request, id: str) -> Response:",
		`Source("user", users_backend, "/users/{id}", required=True),`,
		`Source("orders", order_service_backend, "/orders?userId={id}"),`,
		`dependencies=[Depends(require_roles("admin"))],`,
		"async def aggregate_api_admin_summary(request: Request) -> Response:",
	} {
		if !strings.Contains(router, want) {
			t.Errorf("aggregates router missing %q\n%s", want, router)
		}
	}
	if !strings.Contains(readFile(t, filepath.Join(dir, "routers", "__init__.py")), "from routers.aggregates_router import router as aggregates_router") {
		t.Error("aggregates router not registered")
	}
	if !strings.Contains(readFile(t, filepath.Join(dir, "utils", "upstream.py")), "async def aggregate(") {
		t.Error("upstream.py runtime lacks aggregate()")
	}
}

func TestRuntimeTemplateVersions(t *testing.T) {
	check := func(files []runtimeFile, render func(string) (string, error)) {
		for _, f := range files {
			content, err := render(f.template)
			if err != nil {
				t.Fatal(err)
			}
			if got := runtimeVersionOf([]byte(content)); got != f.version {
				t.Errorf("%s has bffgen:runtime %d, but nodeRuntimeFiles/pythonRuntimeFiles say %d", f.template, got, f.version)
			}
		}
	}
	check(nodeRuntimeFiles, func(name string) (string, error) {
		b, err := os.ReadFile(filepath.Join("..", "..", "..", "internal", "templates", "node", "common", name))
		return string(b), err
	})
	check(pythonRuntimeFiles, func(name string) (string, error) {
		b, err := os.ReadFile(filepath.Join("..", "..", "..", "internal", "templates", "python", "common", name))
		return string(b), err
	})
}

func mustJSON(t *testing.T, s string) map[string]interface{} {
	t.Helper()
	var v map[string]interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatal(err)
	}
	return v
}
