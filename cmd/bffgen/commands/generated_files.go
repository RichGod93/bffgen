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
	Path   string // slash-separated, relative to the project
	Action string
}

func contentHash(content []byte) string {
	sum := sha256.Sum256(content)
	return "sha256:" + hex.EncodeToString(sum[:])
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
		var action string
		switch {
		case os.IsNotExist(err):
			action = genCreate
		case err != nil:
			return nil, err
		case bytes.Equal(existing, []byte(files[p])):
			action = genUnchanged
		case m.Files[p] == contentHash(existing), len(bytes.TrimSpace(existing)) == 0:
			// unchanged since bffgen wrote it, or an empty placeholder (e.g. from init)
			action = genUpdate
		case force:
			action = genReplace
		default:
			action = genKeep
		}
		actions = append(actions, generatedAction{Path: p, Action: action})
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
				if err := os.WriteFile(path+".bak", old, utils.ProjectFilePerm); err != nil {
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

// writeGeneratedFiles plans and (when write is set) applies regeneration-safe writes, printing
// what happened or would happen. It returns the plan.
func writeGeneratedFiles(projectDir string, files map[string]string, write, force bool) ([]generatedAction, error) {
	m, err := loadGeneratedManifest(projectDir)
	if err != nil {
		return nil, err
	}
	actions, err := planGeneratedFiles(projectDir, files, m, force)
	if err != nil {
		return nil, err
	}
	if write {
		if err := applyGeneratedFiles(projectDir, files, m, actions); err != nil {
			return actions, err
		}
	}
	printGeneratedActions(actions, write, !m.existed)
	return actions, nil
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
			fmt.Printf("%s %s (your version: %s.bak)\n", pick(write, "⚠️  Replaced", "🔍 Would replace"), a.Path, a.Path)
		case genUnchanged:
			LogVerboseCommand("Unchanged %s", a.Path)
		case genKeep:
			kept = append(kept, a.Path)
			fmt.Printf("✋ Kept %s: edited since bffgen generated it\n", a.Path)
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

// generatedChangesPending reports whether a plan would change any file (for --check)
func generatedChangesPending(actions []generatedAction) bool {
	for _, a := range actions {
		switch a.Action {
		case genCreate, genUpdate, genReplace:
			return true
		}
	}
	return false
}
