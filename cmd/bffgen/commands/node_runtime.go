// File: node_runtime.go
// Purpose: Runtime support files that generated Node.js and Python routes depend on

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/RichGod93/bffgen/v3/internal/scaffolding"
	"github.com/RichGod93/bffgen/v3/internal/templates"
	"github.com/RichGod93/bffgen/v3/internal/utils"
)

// Each runtime template carries a "bffgen:runtime <n>" marker. Bump a file's version here and
// in its template whenever newly generated routes need a changed copy of that file; only files
// whose version changed are replaced in existing projects.
var runtimeMarker = regexp.MustCompile(`bffgen:runtime (\d+)`)

// runtimeFile maps a template to the project file it is written to
type runtimeFile struct {
	template string
	path     string
	version  int // runtime version generated routes expect
	// userIfUnmarked: bffgen never wrote this file without a marker (it was added in 3.8), so an
	// existing file without one is the user's own and is never replaced
	userIfUnmarked bool
}

// nodeRuntimeFiles are written by init and kept current by generate
var nodeRuntimeFiles = []runtimeFile{
	{template: "upstream.js.tmpl", version: 2, path: filepath.Join("src", "utils", "upstream.js")},
	{template: "service-base.js.tmpl", version: 1, path: filepath.Join("src", "services", "httpClient.js")},
	{template: "circuit-breaker.js.tmpl", version: 1, path: filepath.Join("src", "utils", "circuit-breaker.js")},
	{template: "authorize.js.tmpl", version: 1, path: filepath.Join("src", "middleware", "authorize.js"), userIfUnmarked: true},
}

// pythonRuntimeFiles are written by init and kept current by generate (templates under python/common)
var pythonRuntimeFiles = []runtimeFile{
	{template: "upstream.py.tmpl", version: 2, path: filepath.Join("utils", "upstream.py")},
	{template: "http_client.py.tmpl", version: 1, path: filepath.Join("services", "http_client.py")},
	{template: "circuit_breaker.py.tmpl", version: 1, path: filepath.Join("utils", "circuit_breaker.py")},
	{template: "authorize.py.tmpl", version: 1, path: filepath.Join("utils", "authorize.py"), userIfUnmarked: true},
}

// runtimeAction describes what ensureRuntimeFiles did (or would do) to one file
type runtimeAction struct {
	Path   string
	Action string // "created", "updated", or "conflict" (a file of the user's is in the way; left alone)
	Backup string // previous copy, for "updated"
}

// runtimeChangesPending reports whether ensureRuntimeFiles would write anything (for --check)
func runtimeChangesPending(actions []runtimeAction) bool {
	for _, a := range actions {
		if a.Action != "conflict" {
			return true
		}
	}
	return false
}

// runtimeConflict returns the conflicting file for a runtime template, or ""
func runtimeConflict(actions []runtimeAction, path string) string {
	for _, a := range actions {
		if a.Action == "conflict" && a.Path == path {
			return a.Path
		}
	}
	return ""
}

// ensureNodeRuntimeFiles keeps the Node.js runtime files in projectDir current (see ensureRuntimeFiles)
func ensureNodeRuntimeFiles(projectDir, framework string, write bool) ([]runtimeAction, error) {
	langType := scaffolding.LanguageNodeExpress
	if framework == "fastify" {
		langType = scaffolding.LanguageNodeFastify
	}
	loader := templates.NewTemplateLoader(langType)
	render := func(name string) (string, error) {
		return loader.RenderTemplate(framework, name, &templates.TemplateData{Framework: framework})
	}
	return ensureRuntimeFiles(projectDir, nodeRuntimeFiles, render, write)
}

// ensurePythonRuntimeFiles keeps the Python runtime files in projectDir current (see ensureRuntimeFiles)
func ensurePythonRuntimeFiles(projectDir string, write bool) ([]runtimeAction, error) {
	render := func(name string) (string, error) {
		content, err := templates.TemplateFS.ReadFile("python/common/" + name)
		return string(content), err
	}
	return ensureRuntimeFiles(projectDir, pythonRuntimeFiles, render, write)
}

// ensureRuntimeFiles makes sure the runtime files in projectDir exist and match their version.
// Missing files are created. A file with an older (or no) runtime marker is replaced, and the
// previous copy is kept as <file>.bak. Files that already carry the current marker are left
// alone, including any local edits. With write=false nothing is written and the returned
// actions describe what would happen; earlier backups are never overwritten (see backupPath).
func ensureRuntimeFiles(projectDir string, files []runtimeFile, render func(string) (string, error), write bool) ([]runtimeAction, error) {
	var actions []runtimeAction
	for _, file := range files {
		path := filepath.Join(projectDir, file.path)

		existing, err := os.ReadFile(path)
		var action runtimeAction
		switch {
		case os.IsNotExist(err):
			action = runtimeAction{Path: file.path, Action: "created"}
		case err != nil:
			return actions, err
		case runtimeVersionOf(existing) == file.version:
			continue
		case runtimeVersionOf(existing) == 0 && file.userIfUnmarked:
			actions = append(actions, runtimeAction{Path: file.path, Action: "conflict"})
			continue
		default:
			action = runtimeAction{Path: file.path, Action: "updated", Backup: backupPath(projectDir, file.path)}
		}
		actions = append(actions, action)

		if !write {
			continue
		}

		content, err := render(file.template)
		if err != nil {
			return actions, fmt.Errorf("failed to render %s: %w", file.template, err)
		}
		if existing != nil {
			if err := os.WriteFile(filepath.Join(projectDir, action.Backup), existing, utils.ProjectFilePerm); err != nil {
				return actions, err
			}
		}
		if err := os.MkdirAll(filepath.Dir(path), utils.ProjectDirPerm); err != nil {
			return actions, err
		}
		if err := os.WriteFile(path, []byte(content), utils.ProjectFilePerm); err != nil {
			return actions, err
		}
	}

	return actions, nil
}

// printRuntimeActions reports what ensureRuntimeFiles did, or would do when write is false
func printRuntimeActions(actions []runtimeAction, write bool) {
	for _, a := range actions {
		switch {
		case a.Action == "conflict":
			fmt.Printf("⚠️  %s wasn't written by bffgen, so it was kept. Endpoints with \"roles\" need bffgen's version:\n", a.Path)
			fmt.Println("   rename yours and run 'bffgen generate' again")
		case !write:
			fmt.Printf("🔍 Would %s %s\n", map[string]string{"created": "create", "updated": "update"}[a.Action], a.Path)
		case a.Backup != "":
			fmt.Printf("♻️  Updated %s to the current bffgen runtime (previous copy: %s)\n", a.Path, a.Backup)
		default:
			fmt.Printf("✅ Created %s\n", a.Path)
		}
	}
}

// runtimeVersionOf returns the bffgen:runtime version in a file, or 0 if it has none
func runtimeVersionOf(content []byte) int {
	m := runtimeMarker.FindSubmatch(content)
	if m == nil {
		return 0
	}
	v, _ := strconv.Atoi(string(m[1]))
	return v
}
