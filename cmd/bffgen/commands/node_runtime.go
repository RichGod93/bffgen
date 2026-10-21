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

// Runtime versions that generated routes expect. Each runtime template carries a
// "bffgen:runtime <n>" marker. Bump the number here and in the templates whenever newly
// generated routes need a changed runtime.
const (
	nodeRuntimeVersion   = 1
	pythonRuntimeVersion = 1
)

var runtimeMarker = regexp.MustCompile(`bffgen:runtime (\d+)`)

// runtimeFile maps a template to the project file it is written to
type runtimeFile struct {
	template string
	path     string
}

// nodeRuntimeFiles are written by init and kept current by generate
var nodeRuntimeFiles = []runtimeFile{
	{"upstream.js.tmpl", filepath.Join("src", "utils", "upstream.js")},
	{"service-base.js.tmpl", filepath.Join("src", "services", "httpClient.js")},
	{"circuit-breaker.js.tmpl", filepath.Join("src", "utils", "circuit-breaker.js")},
	{"authorize.js.tmpl", filepath.Join("src", "middleware", "authorize.js")},
}

// pythonRuntimeFiles are written by init and kept current by generate (templates under python/common)
var pythonRuntimeFiles = []runtimeFile{
	{"upstream.py.tmpl", filepath.Join("utils", "upstream.py")},
	{"http_client.py.tmpl", filepath.Join("services", "http_client.py")},
	{"circuit_breaker.py.tmpl", filepath.Join("utils", "circuit_breaker.py")},
	{"authorize.py.tmpl", filepath.Join("utils", "authorize.py")},
}

// runtimeAction describes what ensureRuntimeFiles did (or would do) to one file
type runtimeAction struct {
	Path   string
	Action string // "created" or "updated"
	Backup string // previous copy, for "updated"
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
	return ensureRuntimeFiles(projectDir, nodeRuntimeFiles, nodeRuntimeVersion, render, write)
}

// ensurePythonRuntimeFiles keeps the Python runtime files in projectDir current (see ensureRuntimeFiles)
func ensurePythonRuntimeFiles(projectDir string, write bool) ([]runtimeAction, error) {
	render := func(name string) (string, error) {
		content, err := templates.TemplateFS.ReadFile("python/common/" + name)
		return string(content), err
	}
	return ensureRuntimeFiles(projectDir, pythonRuntimeFiles, pythonRuntimeVersion, render, write)
}

// ensureRuntimeFiles makes sure the runtime files in projectDir exist and match version.
// Missing files are created. A file with an older (or no) runtime marker is replaced, and the
// previous copy is kept as <file>.bak. Files that already carry the current marker are left
// alone, including any local edits. With write=false nothing is written and the returned
// actions describe what would happen; earlier backups are never overwritten (see backupPath).
func ensureRuntimeFiles(projectDir string, files []runtimeFile, version int, render func(string) (string, error), write bool) ([]runtimeAction, error) {
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
		case runtimeVersionOf(existing) == version:
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
