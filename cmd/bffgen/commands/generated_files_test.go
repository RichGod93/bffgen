package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/RichGod93/bffgen/v3/internal/scaffolding"
)

func actionsByPath(actions []generatedAction) map[string]string {
	m := map[string]string{}
	for _, a := range actions {
		m[a.Path] = a.Action
	}
	return m
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestWriteGeneratedFiles(t *testing.T) {
	dir := t.TempDir()
	v1 := map[string]string{"src/a.js": "a1\n", "src/b.js": "b1\n", "src/c.js": "c1\n"}

	// First run creates everything and records it
	actions, err := writeGeneratedFiles(dir, v1, true, false)
	if err != nil {
		t.Fatal(err)
	}
	for p, a := range actionsByPath(actions) {
		if a != genCreate {
			t.Errorf("%s: got %s, want %s", p, a, genCreate)
		}
	}
	manifest := readFile(t, filepath.Join(dir, generatedManifestPath))

	// Same input again: nothing changes, manifest byte-identical (idempotent)
	actions, _ = writeGeneratedFiles(dir, v1, true, false)
	for p, a := range actionsByPath(actions) {
		if a != genUnchanged {
			t.Errorf("rerun %s: got %s, want %s", p, a, genUnchanged)
		}
	}
	if got := readFile(t, filepath.Join(dir, generatedManifestPath)); got != manifest {
		t.Errorf("manifest changed on an identical rerun:\n%s\nvs\n%s", got, manifest)
	}

	// The user edits b.js; the config changes all three files
	writeFile(t, filepath.Join(dir, "src/b.js"), "b1 + my custom code\n")
	v2 := map[string]string{"src/a.js": "a2\n", "src/b.js": "b2\n", "src/c.js": "c1\n"}

	// Dry run writes nothing
	actions, _ = writeGeneratedFiles(dir, v2, false, false)
	if got := actionsByPath(actions); got["src/a.js"] != genUpdate || got["src/b.js"] != genKeep {
		t.Errorf("dry run plan = %v", got)
	}
	if readFile(t, filepath.Join(dir, "src/a.js")) != "a1\n" {
		t.Error("dry run wrote a file")
	}

	actions, _ = writeGeneratedFiles(dir, v2, true, false)
	got := actionsByPath(actions)
	want := map[string]string{"src/a.js": genUpdate, "src/b.js": genKeep, "src/c.js": genUnchanged}
	for p, a := range want {
		if got[p] != a {
			t.Errorf("%s: got %s, want %s", p, got[p], a)
		}
	}
	if readFile(t, filepath.Join(dir, "src/a.js")) != "a2\n" {
		t.Error("unmodified file not regenerated")
	}
	if readFile(t, filepath.Join(dir, "src/b.js")) != "b1 + my custom code\n" {
		t.Error("custom code lost on regeneration")
	}

	// Still kept on the next run: the manifest keeps the hash of what bffgen wrote, not the edit
	actions, _ = writeGeneratedFiles(dir, v2, true, false)
	if actionsByPath(actions)["src/b.js"] != genKeep {
		t.Error("edited file no longer protected on the second run")
	}

	// --force replaces it and keeps the edit as .bak
	actions, _ = writeGeneratedFiles(dir, v2, true, true)
	if actionsByPath(actions)["src/b.js"] != genReplace {
		t.Errorf("force: got %v", actionsByPath(actions))
	}
	if readFile(t, filepath.Join(dir, "src/b.js")) != "b2\n" || readFile(t, filepath.Join(dir, "src/b.js.bak")) != "b1 + my custom code\n" {
		t.Error("force did not replace the file and back up the edit")
	}

	// A backend removed from the config: its file is reported, not deleted
	v3 := map[string]string{"src/a.js": "a2\n", "src/b.js": "b2\n"}
	actions, _ = writeGeneratedFiles(dir, v3, true, false)
	if actionsByPath(actions)["src/c.js"] != genStale {
		t.Errorf("removed file: got %v", actionsByPath(actions))
	}
	if _, err := os.Stat(filepath.Join(dir, "src/c.js")); err != nil {
		t.Error("stale file was deleted")
	}
}

func TestWriteGeneratedFiles_ExistingFilesWithoutManifest(t *testing.T) {
	dir := t.TempDir()
	// A project from before bffgen 3.7: files exist, no manifest
	writeFile(t, filepath.Join(dir, "same.js"), "x\n")
	writeFile(t, filepath.Join(dir, "different.js"), "maybe edited\n")
	writeFile(t, filepath.Join(dir, "placeholder.py"), "")

	files := map[string]string{"same.js": "x\n", "different.js": "new\n", "placeholder.py": "registry\n"}
	actions, err := writeGeneratedFiles(dir, files, true, false)
	if err != nil {
		t.Fatal(err)
	}
	got := actionsByPath(actions)
	want := map[string]string{"same.js": genUnchanged, "different.js": genKeep, "placeholder.py": genUpdate}
	for p, a := range want {
		if got[p] != a {
			t.Errorf("%s: got %s, want %s", p, got[p], a)
		}
	}
	if readFile(t, filepath.Join(dir, "different.js")) != "maybe edited\n" {
		t.Error("file of unknown origin overwritten without --force")
	}
}

func TestRecordGeneratedFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "routers/__init__.py"), "empty registry\n")
	if err := recordGeneratedFiles(dir, map[string]string{"routers/__init__.py": "empty registry\n"}); err != nil {
		t.Fatal(err)
	}
	actions, _ := writeGeneratedFiles(dir, map[string]string{"routers/__init__.py": "registry with routers\n"}, true, false)
	if actionsByPath(actions)["routers/__init__.py"] != genUpdate {
		t.Errorf("file written by init not regenerated: %v", actions)
	}
}

// nodeTestProject creates a minimal Express or Fastify project (no npm install) with two backends
func nodeTestProject(t *testing.T, framework string) string {
	t.Helper()
	parent := t.TempDir()
	dir := filepath.Join(parent, "app")
	create := createNodeExpressMainFileWithOptions
	if framework == "fastify" {
		create = createNodeFastifyMainFileWithOptions
	}
	// Like `bffgen init app`: run from the parent with a relative project name
	err := inDir(t, parent, false, false, false, func() error {
		if err := createProjectDirectories("app", scaffolding.LanguageNodeExpress); err != nil {
			return err
		}
		return create("app", nil, ProjectOptions{SkipTests: true, SkipDocs: true, MiddlewareFlag: "none"})
	})
	if err != nil {
		t.Fatal(err)
	}
	endpoint := func(method, path, exposeAs string) map[string]interface{} {
		return map[string]interface{}{"name": "e", "method": method, "path": path, "exposeAs": exposeAs, "requiresAuth": true}
	}
	config := map[string]interface{}{
		"project": map[string]interface{}{"name": "app", "framework": framework},
		"backends": []interface{}{
			map[string]interface{}{"name": "users", "baseUrl": "http://localhost:4001/api", "endpoints": []interface{}{
				endpoint("GET", "/users/:id", "/api/users/:id"), endpoint("POST", "/users", "/api/users"),
			}},
			map[string]interface{}{"name": "orders", "baseUrl": "http://localhost:4002/api", "endpoints": []interface{}{
				endpoint("GET", "/orders", "/api/orders"),
			}},
		},
	}
	data, _ := json.MarshalIndent(config, "", "  ")
	writeFile(t, filepath.Join(dir, "bffgen.config.json"), string(data))
	return dir
}

// inDir runs fn with the working directory set to dir and the generate flags set as given
func inDir(t *testing.T, dir string, dry, check, force bool, fn func() error) error {
	t.Helper()
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	oldDry, oldCheck, oldForce := dryRun, checkMode, forceMode
	dryRun, checkMode, forceMode = dry, check, force
	defer func() { dryRun, checkMode, forceMode = oldDry, oldCheck, oldForce }()
	return fn()
}

// snapshot returns every file under dir (path -> content)
func snapshot(t *testing.T, dir string) map[string]string {
	t.Helper()
	files := map[string]string{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(dir, path)
		data, err := os.ReadFile(path)
		files[filepath.ToSlash(rel)] = string(data)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func diffSnapshots(a, b map[string]string) []string {
	var diffs []string
	for p, c := range a {
		if b[p] != c {
			diffs = append(diffs, p)
		}
	}
	for p := range b {
		if _, ok := a[p]; !ok {
			diffs = append(diffs, p)
		}
	}
	sort.Strings(diffs)
	return diffs
}

func TestNodeGenerate_RegenerationSafety(t *testing.T) {
	for _, framework := range []string{"express", "fastify"} {
		t.Run(framework, func(t *testing.T) {
			dir := nodeTestProject(t, framework)
			before := snapshot(t, dir)

			// --dry-run and --check write nothing; --check fails while generation is pending
			if err := inDir(t, dir, true, false, false, generateNodeJS); err != nil {
				t.Fatal(err)
			}
			if err := inDir(t, dir, false, true, false, generateNodeJS); err == nil {
				t.Error("--check succeeded on a project with pending changes")
			}
			if d := diffSnapshots(before, snapshot(t, dir)); len(d) > 0 {
				t.Fatalf("--dry-run/--check changed files: %v", d)
			}

			if err := inDir(t, dir, false, false, false, generateNodeJS); err != nil {
				t.Fatal(err)
			}
			generated := snapshot(t, dir)
			for _, p := range []string{"src/routes/users.js", "src/services/orders.service.js", "src/controllers/users.controller.js", ".bffgen/generated.json"} {
				if _, ok := generated[p]; !ok {
					t.Errorf("%s not generated", p)
				}
			}
			if !strings.Contains(generated["src/index.js"], "require('./routes/orders')") {
				t.Error("routes not registered in src/index.js")
			}

			// Idempotent: a second run changes nothing, and --check passes
			if err := inDir(t, dir, false, false, false, generateNodeJS); err != nil {
				t.Fatal(err)
			}
			if d := diffSnapshots(generated, snapshot(t, dir)); len(d) > 0 {
				t.Errorf("second run changed files: %v", d)
			}
			if err := inDir(t, dir, false, true, false, generateNodeJS); err != nil {
				t.Errorf("--check failed on an up-to-date project: %v", err)
			}

			// Custom code survives regeneration after a config change
			custom := generated["src/routes/users.js"] + "\n// my custom code\n"
			writeFile(t, filepath.Join(dir, "src/routes/users.js"), custom)
			indexCustom := generated["src/index.js"] + "\n// my custom index code\n"
			writeFile(t, filepath.Join(dir, "src/index.js"), indexCustom)
			config := strings.Replace(readFile(t, filepath.Join(dir, "bffgen.config.json")), `"/api/orders"`, `"/api/all-orders"`, 1)
			writeFile(t, filepath.Join(dir, "bffgen.config.json"), config)

			if err := inDir(t, dir, false, false, false, generateNodeJS); err != nil {
				t.Fatal(err)
			}
			if readFile(t, filepath.Join(dir, "src/routes/users.js")) != custom {
				t.Error("edited route file was overwritten")
			}
			if !strings.Contains(readFile(t, filepath.Join(dir, "src/index.js")), "// my custom index code") {
				t.Error("custom code in src/index.js was lost")
			}
			if !strings.Contains(readFile(t, filepath.Join(dir, "src/routes/orders.js")), "/api/all-orders") {
				t.Error("unedited route file did not pick up the config change")
			}
		})
	}
}

func TestNodeGenerate_Deterministic(t *testing.T) {
	for _, framework := range []string{"express", "fastify"} {
		t.Run(framework, func(t *testing.T) {
			a, b := nodeTestProject(t, framework), nodeTestProject(t, framework)
			for _, dir := range []string{a, b} {
				if err := inDir(t, dir, false, false, false, generateNodeJS); err != nil {
					t.Fatal(err)
				}
			}
			if d := diffSnapshots(snapshot(t, a), snapshot(t, b)); len(d) > 0 {
				t.Errorf("same config gave different output in: %v", d)
			}
		})
	}
}

func TestNodeGenerate_InvalidConfigWritesNothing(t *testing.T) {
	dir := nodeTestProject(t, "express")
	config := strings.Replace(readFile(t, filepath.Join(dir, "bffgen.config.json")), `"orders"`, `"../orders"`, 1)
	writeFile(t, filepath.Join(dir, "bffgen.config.json"), config)
	before := snapshot(t, dir)

	if err := inDir(t, dir, false, false, false, generateNodeJS); err == nil {
		t.Error("invalid backend name accepted")
	}
	if d := diffSnapshots(before, snapshot(t, dir)); len(d) > 0 {
		t.Errorf("invalid config still wrote: %v", d)
	}
}
