// File: generated_files.go
// Purpose: Regeneration-safe writes for the files `bffgen generate` owns (Node.js and Python)

package commands

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/RichGod93/bffgen/v3/internal/utils"
)

// generatedManifestPath records the SHA-256 of every file generate wrote. Commit it: it is how the
// next run knows which files are unchanged since generation (safe to regenerate) and which you
// edited (kept). It contains no timestamps, so the same config gives a byte-identical file.
var generatedManifestPath = filepath.Join(".bffgen", "generated.json")

const generatedManifestVersion = 1

type generatedManifest struct {
	Version int               `json:"version"`
	Files   map[string]string `json:"files"` // slash-separated path -> "sha256:<hex>"

	existed bool // loaded from disk (false for projects made before bffgen 3.7)
}

// Actions planGeneratedFiles can decide on for one file
const (
	genCreate    = "create"    // file doesn't exist
	genUpdate    = "update"    // unchanged since bffgen wrote it; new content differs
	genUnchanged = "unchanged" // already has the new content
	genKeep      = "keep"      // edited since generation (or not written by bffgen); left alone
	genReplace   = "replace"   // edited, but --force: replaced, old copy kept as <file>.bak
	genStale     = "stale"     // generated earlier, no longer produced by the config; left alone
)

type generatedAction struct {
	Path     string // slash-separated, relative to the project
	Action   string
	Backup   string // genReplace: where the edited version is saved
	Outdated bool   // genKeep: the config now generates different content than the kept file was made from
}

// contentHash hashes content with CRLF line endings normalised to LF, so a checkout with
// core.autocrlf=true doesn't make every generated file look edited
func contentHash(content []byte) string {
	sum := sha256.Sum256(normalizeEOL(content))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func normalizeEOL(content []byte) []byte {
	return bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
}

// backupPath returns where to save the previous copy of path (relative to projectDir):
// <path>.bak, or <path>.bak.1, <path>.bak.2, ... if earlier backups exist, so none is overwritten
func backupPath(projectDir, path string) string {
	backup := path + ".bak"
	for n := 1; ; n++ {
		if _, err := os.Lstat(filepath.Join(projectDir, filepath.FromSlash(backup))); os.IsNotExist(err) {
			return backup
		}
		backup = fmt.Sprintf("%s.bak.%d", path, n)
	}
}

func loadGeneratedManifest(projectDir string) (*generatedManifest, error) {
	m := &generatedManifest{Version: generatedManifestVersion, Files: map[string]string{}}
	data, err := os.ReadFile(filepath.Join(projectDir, generatedManifestPath))
	if os.IsNotExist(err) {
		return m, nil
	}
	if err != nil {
		return nil, err
	}
	m.existed = true
	if err := json.Unmarshal(data, m); err != nil {
		return nil, fmt.Errorf("%s is not valid JSON (delete it to start over): %w", generatedManifestPath, err)
	}
	if m.Files == nil {
		m.Files = map[string]string{}
	}
	return m, nil
}

// planGeneratedFiles decides, for every file in files (slash path -> content), whether writing it
// is safe. A file is only overwritten if it still has the content bffgen last wrote, unless force
// is set. Files the manifest lists that the config no longer produces are reported as stale.
func planGeneratedFiles(projectDir string, files map[string]string, m *generatedManifest, force bool) ([]generatedAction, error) {
	paths := make([]string, 0, len(files))
	for p := range files {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	var actions []generatedAction
	for _, p := range paths {
		existing, err := os.ReadFile(filepath.Join(projectDir, filepath.FromSlash(p)))
		a := generatedAction{Path: p}
		switch {
		case os.IsNotExist(err):
			a.Action = genCreate
		case err != nil:
			return nil, err
		case bytes.Equal(normalizeEOL(existing), []byte(files[p])):
			a.Action = genUnchanged
		case m.Files[p] == contentHash(existing):
			a.Action = genUpdate // unchanged since bffgen wrote it
		case m.Files[p] == "" && len(bytes.TrimSpace(existing)) == 0:
			a.Action = genUpdate // empty placeholder from a project made before bffgen 3.7
		case force:
			a.Action = genReplace
			a.Backup = backupPath(projectDir, p)
		default:
			a.Action = genKeep
			a.Outdated = m.Files[p] != contentHash([]byte(files[p]))
		}
		actions = append(actions, a)
	}

	var stale []string
	for p := range m.Files {
		if _, ok := files[p]; ok {
			continue
		}
		if _, err := os.Stat(filepath.Join(projectDir, filepath.FromSlash(p))); err == nil {
			stale = append(stale, p)
		}
	}
	sort.Strings(stale)
	for _, p := range stale {
		actions = append(actions, generatedAction{Path: p, Action: genStale})
	}
	return actions, nil
}

// applyGeneratedFiles carries out a plan and saves the updated manifest
func applyGeneratedFiles(projectDir string, files map[string]string, m *generatedManifest, actions []generatedAction) error {
	for _, a := range actions {
		path := filepath.Join(projectDir, filepath.FromSlash(a.Path))
		switch a.Action {
		case genCreate, genUpdate, genReplace:
			if a.Action == genReplace {
				old, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(projectDir, filepath.FromSlash(a.Backup)), old, utils.ProjectFilePerm); err != nil {
					return err
				}
			}
			if err := os.MkdirAll(filepath.Dir(path), utils.ProjectDirPerm); err != nil {
				return err
			}
			if err := os.WriteFile(path, []byte(files[a.Path]), utils.ProjectFilePerm); err != nil {
				return err
			}
			m.Files[a.Path] = contentHash([]byte(files[a.Path]))
		case genUnchanged:
			m.Files[a.Path] = contentHash([]byte(files[a.Path]))
		}
	}

	// Forget files that were generated once but are gone now
	for p := range m.Files {
		if _, err := os.Stat(filepath.Join(projectDir, filepath.FromSlash(p))); os.IsNotExist(err) {
			delete(m.Files, p)
		}
	}

	m.Version = generatedManifestVersion
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	manifest := filepath.Join(projectDir, generatedManifestPath)
	if err := os.MkdirAll(filepath.Dir(manifest), utils.ProjectDirPerm); err != nil {
		return err
	}
	return os.WriteFile(manifest, append(data, '\n'), utils.ProjectFilePerm)
}

// recordGeneratedFiles marks files that init wrote, and generate later owns, as generated
func recordGeneratedFiles(projectDir string, files map[string]string) error {
	m, err := loadGeneratedManifest(projectDir)
	if err != nil {
		return err
	}
	actions := make([]generatedAction, 0, len(files))
	for p := range files {
		actions = append(actions, generatedAction{Path: p, Action: genUnchanged})
	}
	return applyGeneratedFiles(projectDir, files, m, actions)
}

// generatedPlan is what writeGeneratedFiles will do. Planning reads but never writes, so
// generate can plan everything (and fail on a bad manifest or path) before writing anything.
type generatedPlan struct {
	files    map[string]string
	manifest *generatedManifest
	Actions  []generatedAction
}

func planGeneratedWrites(projectDir string, files map[string]string, force bool) (*generatedPlan, error) {
	m, err := loadGeneratedManifest(projectDir)
	if err != nil {
		return nil, err
	}
	actions, err := planGeneratedFiles(projectDir, files, m, force)
	if err != nil {
		return nil, err
	}
	return &generatedPlan{files: files, manifest: m, Actions: actions}, nil
}

func (p *generatedPlan) apply(projectDir string) error {
	return applyGeneratedFiles(projectDir, p.files, p.manifest, p.Actions)
}

func (p *generatedPlan) print(write bool) {
	printGeneratedActions(p.Actions, write, !p.manifest.existed)
}

// writeGeneratedFiles plans and (when write is set) applies regeneration-safe writes, printing
// what happened or would happen. It returns the plan.
func writeGeneratedFiles(projectDir string, files map[string]string, write, force bool) ([]generatedAction, error) {
	plan, err := planGeneratedWrites(projectDir, files, force)
	if err != nil {
		return nil, err
	}
	if write {
		if err := plan.apply(projectDir); err != nil {
			return plan.Actions, err
		}
	}
	plan.print(write)
	return plan.Actions, nil
}

// printGeneratedActions reports a plan; unchanged files are only listed with --verbose.
// firstRun is set for projects without a manifest yet (made before bffgen 3.7).
func printGeneratedActions(actions []generatedAction, write, firstRun bool) {
	var kept []string
	for _, a := range actions {
		switch a.Action {
		case genCreate:
			fmt.Printf("%s %s\n", pick(write, "✅ Created", "🔍 Would create"), a.Path)
		case genUpdate:
			fmt.Printf("%s %s\n", pick(write, "♻️  Updated", "🔍 Would update"), a.Path)
		case genReplace:
			fmt.Printf("%s %s (your version: %s)\n", pick(write, "⚠️  Replaced", "🔍 Would replace"), a.Path, a.Backup)
		case genUnchanged:
			LogVerboseCommand("Unchanged %s", a.Path)
		case genKeep:
			kept = append(kept, a.Path)
			if a.Outdated {
				fmt.Printf("✋ Kept %s: edited since bffgen generated it, and the config has changed since\n", a.Path)
			} else {
				fmt.Printf("✋ Kept %s: edited since bffgen generated it\n", a.Path)
			}
		case genStale:
			fmt.Printf("🗑️  %s is no longer generated by the config; delete it if you don't use it\n", a.Path)
		}
	}
	switch {
	case len(kept) > 0 && firstRun:
		fmt.Println("💡 This project has no " + filepath.ToSlash(generatedManifestPath) + " yet (it was made with bffgen < 3.7), so")
		fmt.Println("   files that differ from what bffgen generates now were kept in case you edited them.")
		fmt.Println("   If you didn't, run 'bffgen generate --force' once (old copies are saved as <file>.bak).")
	case len(kept) > 0:
		fmt.Println("💡 Kept files don't get config changes. To regenerate them, run 'bffgen generate --force'")
		fmt.Println("   (your versions are saved as <file>.bak), then copy your changes back.")
	}
}

func pick(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}

// generatedChangesPending reports whether a plan would change any file, or keeps an edited
// file that no longer matches the config (for --check). An edited file whose generated
// content is still current is up to date: your edits sit on top of the current output.
func generatedChangesPending(actions []generatedAction) bool {
	for _, a := range actions {
		switch {
		case a.Action == genCreate, a.Action == genUpdate, a.Action == genReplace:
			return true
		case a.Action == genKeep && a.Outdated:
			return true
		}
	}
	return false
}
