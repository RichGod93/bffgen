// File: aggregates.go
// Purpose: Config-driven aggregation routes ("aggregates" in bffgen.config.json and
// bffgen.config.py.json) for the Node.js and Python generators

package commands

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/RichGod93/bffgen/v3/internal/scaffolding"
	"github.com/RichGod93/bffgen/v3/internal/templates"
)

// aggregatesFile is the base name of the generated aggregates route file (src/routes/aggregates.js,
// routers/aggregates_router.py); no backend may use it
const aggregatesFile = "aggregates"

// aggregateSpec is one entry of "aggregates": a GET route that calls several backends in
// parallel and answers with their combined responses
//
//	{
//	  "name": "dashboard",                       optional, used in comments and handler names
//	  "exposeAs": "/api/dashboard/:id",          path the BFF exposes ("path" also works in Python configs)
//	  "requiresAuth": true, "roles": ["user"],   as for endpoints
//	  "sources": [
//	    {"backend": "users", "path": "/users/:id", "as": "user", "required": true},
//	    {"backend": "orders", "path": "/orders?userId=:id", "as": "orders"}
//	  ]
//	}
type aggregateSpec struct {
	Name         string
	Path         string // exposed path, :param style
	RequiresAuth bool
	Roles        []string
	Sources      []aggregateSource
}

// aggregateSource is one backend call of an aggregate
type aggregateSource struct {
	Backend  string
	Path     string // backend path, :param style; may end in a query string
	As       string // key of this source's data in the response
	Required bool   // a failure fails the whole aggregate instead of leaving the data null
}

var (
	// A source path may carry a query string, e.g. /orders?userId=:id
	aggregateSourcePath = regexp.MustCompile(`^/[A-Za-z0-9\-._~!$&()*+,;=:@%/]*(\?[A-Za-z0-9\-._~!$&()*+,;=:@%/?]*)?$`)
	aggregateKey        = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]*$`)
)

// routePattern reduces a path to its shape, so /users/:id and /users/{userId} compare equal
func routePattern(path string) string {
	return colonParam.ReplaceAllString(normalizeRouteParams(path), ":")
}

// parseAggregates validates the "aggregates" section of a Node.js (python=false) or Python
// config. Sources must name a configured backend, and an aggregate's path must not clash with
// another aggregate or with a GET endpoint.
func parseAggregates(config map[string]interface{}, python bool) ([]aggregateSpec, error) {
	raw, ok := config["aggregates"]
	if !ok || raw == nil {
		return nil, nil
	}
	list, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf(`"aggregates" must be a list`)
	}

	backends := map[string]bool{}
	routes := map[string]string{} // route pattern -> what defines it
	configured, _ := config["backends"].([]interface{})
	for _, b := range configured {
		backend, _ := b.(map[string]interface{})
		name, _ := backend["name"].(string)
		if name == "" {
			continue
		}
		backends[name] = true
		endpoints, _ := backend["endpoints"].([]interface{})
		for _, e := range endpoints {
			endpoint, _ := e.(map[string]interface{})
			method, _ := endpoint["method"].(string)
			exposed, _ := endpoint["exposeAs"].(string)
			if exposed == "" && python {
				exposed, _ = endpoint["path"].(string)
			}
			if method = strings.ToUpper(method); exposed != "" && (method == "GET" || method == "HEAD") {
				routes[routePattern(exposed)] = fmt.Sprintf("the %s endpoint %s %s", name, method, exposed)
			}
		}
	}

	var specs []aggregateSpec
	for i, a := range list {
		entry, ok := a.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("aggregate %d: must be an object", i)
		}
		spec, err := parseAggregate(entry, python, backends)
		label := fmt.Sprintf("aggregate %d", i)
		if spec.Name != "" {
			label = fmt.Sprintf("aggregate %q", spec.Name)
		} else if spec.Path != "" {
			label = "aggregate " + spec.Path
		}
		if err != nil {
			return nil, fmt.Errorf("%s: %w", label, err)
		}
		pattern := routePattern(spec.Path)
		if other, ok := routes[pattern]; ok {
			return nil, fmt.Errorf("%s: GET %s is already defined by %s", label, spec.Path, other)
		}
		routes[pattern] = label
		specs = append(specs, spec)
	}
	return specs, nil
}

// parseAggregate validates one aggregate; the returned spec carries Name and Path even on error
func parseAggregate(entry map[string]interface{}, python bool, backends map[string]bool) (aggregateSpec, error) {
	var spec aggregateSpec
	spec.Name, _ = entry["name"].(string)
	spec.Path, _ = entry["exposeAs"].(string)
	if spec.Path == "" && python {
		spec.Path, _ = entry["path"].(string)
	}
	if spec.Name != "" && !nodeServiceName.MatchString(spec.Name) {
		return spec, fmt.Errorf("name %q must contain only letters, digits, '-' and '_'", spec.Name)
	}
	if spec.Path == "" {
		return spec, fmt.Errorf(`"exposeAs" is required`)
	}
	spec.Path = normalizeRouteParams(spec.Path)
	if !safeRoutePath.MatchString(spec.Path) {
		return spec, fmt.Errorf("exposeAs %q must start with / and contain only URL path characters", spec.Path)
	}
	if method, ok := entry["method"].(string); ok && strings.ToUpper(method) != "GET" {
		return spec, fmt.Errorf("aggregates are GET routes; remove \"method\": %q", method)
	}
	exposed := map[string]bool{}
	for _, param := range routeParams(spec.Path) {
		if python && pythonReservedNames[param] {
			return spec, fmt.Errorf("path parameter %q is a reserved name in Python; rename it", param)
		}
		exposed[param] = true
	}

	roles, err := endpointRoles(entry)
	if err != nil {
		return spec, err
	}
	requiresAuth, _ := entry["requiresAuth"].(bool)
	spec.Roles, spec.RequiresAuth = roles, requiresAuth || len(roles) > 0

	sources, ok := entry["sources"].([]interface{})
	if !ok || len(sources) == 0 {
		return spec, fmt.Errorf(`"sources" must be a non-empty list of {"backend", "path", "as"}`)
	}
	keys := map[string]bool{}
	for j, s := range sources {
		source, ok := s.(map[string]interface{})
		if !ok {
			return spec, fmt.Errorf("source %d: must be an object", j)
		}
		var src aggregateSource
		src.Backend, _ = source["backend"].(string)
		src.Path, _ = source["path"].(string)
		src.As, _ = source["as"].(string)
		if src.As == "" {
			src.As = src.Backend
		}
		if required, ok := source["required"]; ok {
			if src.Required, ok = required.(bool); !ok {
				return spec, fmt.Errorf("source %d: \"required\" must be true or false", j)
			}
		}

		switch {
		case src.Backend == "":
			return spec, fmt.Errorf("source %d: \"backend\" is required", j)
		case !backends[src.Backend]:
			return spec, fmt.Errorf("source %d: backend %q is not configured in \"backends\"", j, src.Backend)
		case !aggregateKey.MatchString(src.As):
			return spec, fmt.Errorf("source %d: \"as\" %q must start with a letter or '_' and contain only letters, digits, '-' and '_'", j, src.As)
		case keys[src.As]:
			return spec, fmt.Errorf("source %d: \"as\" %q is used twice; give each source its own \"as\"", j, src.As)
		}
		keys[src.As] = true

		src.Path = normalizeRouteParams(src.Path)
		if !aggregateSourcePath.MatchString(src.Path) {
			return spec, fmt.Errorf("source %q: path %q must start with / and contain only URL characters", src.As, src.Path)
		}
		for _, param := range routeParams(src.Path) {
			if !exposed[param] {
				return spec, fmt.Errorf("source %q: path %s uses :%s, which exposeAs %s does not define", src.As, src.Path, param, spec.Path)
			}
		}
		spec.Sources = append(spec.Sources, src)
	}
	return spec, nil
}

// aggregateBackends returns the backends the aggregates call, in order of first use
func aggregateBackends(specs []aggregateSpec) []string {
	var names []string
	seen := map[string]bool{}
	for _, spec := range specs {
		for _, src := range spec.Sources {
			if !seen[src.Backend] {
				seen[src.Backend] = true
				names = append(names, src.Backend)
			}
		}
	}
	return names
}

// findBackend returns the backend with the given name from a config
func findBackend(config map[string]interface{}, name string) map[string]interface{} {
	backends, _ := config["backends"].([]interface{})
	for _, b := range backends {
		if backend, ok := b.(map[string]interface{}); ok && backend["name"] == name {
			return backend
		}
	}
	return nil
}

// nodeAggregatesData is the template data for src/routes/aggregates.js
type nodeAggregatesData struct {
	AnyAuth    bool
	AnyRoles   bool
	Backends   []nodeAggregateBackend
	Aggregates []nodeAggregateRoute
}

// nodeAggregateBackend is one backend the aggregates call; Var is its upstream's JS variable
type nodeAggregateBackend struct {
	Var string
	templates.RouteTemplateData
}

type nodeAggregateRoute struct {
	Name         string
	Path         string
	RequiresAuth bool
	Roles        []string
	RolesJS      string
	Sources      []nodeAggregateSource
}

type nodeAggregateSource struct {
	As       string
	Var      string
	Path     string // matches aggregateSourcePath (no quotes or backslashes), so it is written unescaped
	Required bool
}

// renderNodeAggregates renders src/routes/aggregates.js
func renderNodeAggregates(config map[string]interface{}, specs []aggregateSpec, framework string) (string, error) {
	data := &nodeAggregatesData{}
	vars := map[string]string{}
	for _, name := range aggregateBackends(specs) {
		backend := findBackend(config, name)
		baseURL, _ := backend["baseUrl"].(string)
		if baseURL == "" {
			return "", fmt.Errorf("backend %s: baseUrl is required", name)
		}
		b := nodeAggregateBackend{
			Var: nodeIdentifier(name) + "Backend",
			RouteTemplateData: templates.RouteTemplateData{
				ServiceName: name,
				EnvKey:      nodeEnvKey(name),
				BaseURL:     baseURL,
			},
		}
		if err := applyResilienceConfig(&b.RouteTemplateData, backend); err != nil {
			return "", fmt.Errorf("backend %s: %w", name, err)
		}
		vars[name] = b.Var
		data.Backends = append(data.Backends, b)
	}

	for _, spec := range specs {
		route := nodeAggregateRoute{
			Name:         spec.Name,
			Path:         spec.Path,
			RequiresAuth: spec.RequiresAuth,
			Roles:        spec.Roles,
			RolesJS:      jsStringArray(spec.Roles),
		}
		for _, src := range spec.Sources {
			route.Sources = append(route.Sources, nodeAggregateSource{As: src.As, Var: vars[src.Backend], Path: src.Path, Required: src.Required})
		}
		data.AnyAuth = data.AnyAuth || spec.RequiresAuth
		data.AnyRoles = data.AnyRoles || len(spec.Roles) > 0
		data.Aggregates = append(data.Aggregates, route)
	}

	langType := scaffolding.LanguageNodeExpress
	if framework == "fastify" {
		langType = scaffolding.LanguageNodeFastify
	}
	content, err := templates.NewTemplateLoader(langType).LoadTemplate(framework, "aggregates-template.js.tmpl")
	if err != nil {
		return "", err
	}
	tmpl, err := template.New("aggregates-template.js.tmpl").Parse(content)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render aggregates: %w", err)
	}
	return buf.String(), nil
}

// pythonAggregatesData is the template data for routers/aggregates_router.py
type pythonAggregatesData struct {
	AnyAuth    bool // an aggregate requires authentication but no roles
	AnyRoles   bool
	Backends   []pythonAggregateBackend
	Aggregates []pythonAggregateRoute
}

// pythonAggregateBackend is one backend the aggregates call; Var is its upstream's variable
type pythonAggregateBackend struct {
	Var string
	*pythonBackendData
}

type pythonAggregateRoute struct {
	Name         string
	Path         string // {param} style
	Params       []string
	HandlerName  string
	RequiresAuth bool
	Roles        []string
	RolesArgs    string
	Sources      []nodeAggregateSource // Path in {param} style
}

// renderPythonAggregates renders routers/aggregates_router.py
func renderPythonAggregates(config map[string]interface{}, specs []aggregateSpec) (string, error) {
	data := &pythonAggregatesData{}
	vars := map[string]string{}
	for _, name := range aggregateBackends(specs) {
		backend := findBackend(config, name)
		baseURL, _ := backend["baseUrl"].(string)
		if baseURL == "" {
			baseURL, _ = backend["url"].(string)
		}
		if baseURL == "" {
			return "", fmt.Errorf("backend %s: baseUrl is required", name)
		}
		module := pythonModuleName(name)
		b := pythonAggregateBackend{
			Var:               module + "_backend",
			pythonBackendData: &pythonBackendData{ServiceName: name, Module: module, EnvKey: nodeEnvKey(name), BaseURL: baseURL},
		}
		if err := applyPythonResilienceConfig(b.pythonBackendData, backend); err != nil {
			return "", fmt.Errorf("backend %s: %w", name, err)
		}
		vars[name] = b.Var
		data.Backends = append(data.Backends, b)
	}

	handlers := map[string]bool{}
	for _, spec := range specs {
		base := spec.Name
		if base == "" {
			base = strings.TrimPrefix(nodeHandlerName("GET", spec.Path), "get")
		}
		if base == "" {
			base = "root"
		}
		handler := "aggregate_" + toSnakeCase(base)
		for h, n := handler, 2; handlers[handler]; n++ {
			handler = fmt.Sprintf("%s_%d", h, n)
		}
		handlers[handler] = true

		route := pythonAggregateRoute{
			Name:         spec.Name,
			Path:         colonToBraceParam.ReplaceAllString(spec.Path, "{$1}"),
			Params:       routeParams(spec.Path),
			HandlerName:  handler,
			RequiresAuth: spec.RequiresAuth,
			Roles:        spec.Roles,
			RolesArgs:    pythonStringArgs(spec.Roles),
		}
		for _, src := range spec.Sources {
			route.Sources = append(route.Sources, nodeAggregateSource{
				As: src.As, Var: vars[src.Backend], Path: colonToBraceParam.ReplaceAllString(src.Path, "{$1}"), Required: src.Required,
			})
		}
		data.AnyAuth = data.AnyAuth || spec.RequiresAuth && len(spec.Roles) == 0
		data.AnyRoles = data.AnyRoles || len(spec.Roles) > 0
		data.Aggregates = append(data.Aggregates, route)
	}
	return renderPythonTemplate("python/fastapi/aggregates_router.py.tmpl", data)
}
