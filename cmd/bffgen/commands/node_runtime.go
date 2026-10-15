// File: node_runtime.go
// Purpose: Runtime support files that generated Node.js routes depend on

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/RichGod93/bffgen/internal/scaffolding"
	"github.com/RichGod93/bffgen/internal/templates"
	"github.com/RichGod93/bffgen/internal/utils"
)

// nodeRuntimeVersion is the version of the runtime files that generated routes expect.
// Each runtime template carries a "bffgen:runtime <n>" marker. Bump the number in this
// constant and in the templates whenever newly generated routes need a changed runtime.
const nodeRuntimeVersion = 1

var nodeRuntimeMarker = regexp.MustCompile(`bffgen:runtime (\d+)`)

// nodeRuntimeFiles are written by init and kept current by generate
var nodeRuntimeFiles = []struct {
	template string
	path     string
}{
	{"upstream.js.tmpl", filepath.Join("src", "utils", "upstream.js")},
}

// nodeRuntimeAction describes what ensureNodeRuntimeFiles did (or would do) to one file
type nodeRuntimeAction struct {
	Path   string
	Action string // "created" or "updated"
	Backup string // previous copy, for "updated"
}

// ensureNodeRuntimeFiles makes sure the runtime files in projectDir exist and match
// nodeRuntimeVersion. Missing files are created. A file with an older (or no) runtime marker
// is replaced, and the previous copy is kept as <file>.bak. Files that already carry the
// current marker are left alone, including any local edits. With write=false nothing is
// written and the returned actions describe what would happen.
func ensureNodeRuntimeFiles(projectDir, framework string, write bool) ([]nodeRuntimeAction, error) {
	langType := scaffolding.LanguageNodeExpress
	if framework == "fastify" {
		langType = scaffolding.LanguageNodeFastify
	}
	loader := templates.NewTemplateLoader(langType)

	var actions []nodeRuntimeAction
	for _, file := range nodeRuntimeFiles {
		path := filepath.Join(projectDir, file.path)

		existing, err := os.ReadFile(path)
		switch {
		case os.IsNotExist(err):
			actions = append(actions, nodeRuntimeAction{Path: file.path, Action: "created"})
		case err != nil:
			return actions, err
		case runtimeVersionOf(existing) == nodeRuntimeVersion:
			continue
		default:
			actions = append(actions, nodeRuntimeAction{Path: file.path, Action: "updated", Backup: file.path + ".bak"})
		}

		if !write {
			continue
		}

		content, err := loader.RenderTemplate(framework, file.template, &templates.TemplateData{Framework: framework})
		if err != nil {
			return actions, fmt.Errorf("failed to render %s: %w", file.template, err)
		}
		if existing != nil {
			if err := os.WriteFile(path+".bak", existing, utils.ProjectFilePerm); err != nil {
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

// runtimeVersionOf returns the bffgen:runtime version in a file, or 0 if it has none
func runtimeVersionOf(content []byte) int {
	m := nodeRuntimeMarker.FindSubmatch(content)
	if m == nil {
		return 0
	}
	v, _ := strconv.Atoi(string(m[1]))
	return v
}
