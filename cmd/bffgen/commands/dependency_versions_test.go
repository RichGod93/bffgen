package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RichGod93/bffgen/v3/internal/scaffolding"
	"github.com/RichGod93/bffgen/v3/internal/templates"
	"github.com/RichGod93/bffgen/v3/internal/types"
)

// Generated package.json files must pin every dependency to the manifest version;
// an empty version makes npm install whatever is latest.
func TestNodeInit_PackageJsonUsesManifestVersions(t *testing.T) {
	cases := []struct {
		runtime  string
		langType scaffolding.LanguageType
		create   func(string, []types.BackendService, ProjectOptions) error
	}{
		{"nodejs-express", scaffolding.LanguageNodeExpress, createNodeExpressMainFileWithOptions},
		{"nodejs-fastify", scaffolding.LanguageNodeFastify, createNodeFastifyMainFileWithOptions},
	}

	for _, tc := range cases {
		t.Run(tc.runtime, func(t *testing.T) {
			projectName := filepath.Join(t.TempDir(), "app")
			if err := createProjectDirectories(projectName, tc.langType); err != nil {
				t.Fatal(err)
			}

			opts := ProjectOptions{SkipTests: true, SkipDocs: true, MiddlewareFlag: "none"}
			if err := tc.create(projectName, nil, opts); err != nil {
				t.Fatalf("init failed: %v", err)
			}

			content, err := os.ReadFile(filepath.Join(projectName, "package.json"))
			if err != nil {
				t.Fatal(err)
			}

			var pkg struct {
				Dependencies    map[string]string `json:"dependencies"`
				DevDependencies map[string]string `json:"devDependencies"`
			}
			if err := json.Unmarshal(content, &pkg); err != nil {
				t.Fatalf("package.json is not valid JSON: %v", err)
			}

			versions, err := templates.LoadVersions(tc.runtime)
			if err != nil {
				t.Fatal(err)
			}

			for _, deps := range []map[string]string{pkg.Dependencies, pkg.DevDependencies} {
				for name, version := range deps {
					if version == "" {
						t.Errorf("dependency %q has an empty version", name)
					} else if want := versions[name]; version != want {
						t.Errorf("dependency %q = %q, want manifest version %q", name, version, want)
					}
				}
			}
		})
	}
}

func TestRenderRequirementsTxt_UsesManifestVersions(t *testing.T) {
	content, err := renderRequirementsTxt()
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	text := string(content)
	if strings.Contains(text, "{{") {
		t.Fatalf("requirements.txt contains unrendered template syntax:\n%s", text)
	}
	if !strings.Contains(text, "fastapi>=") {
		t.Errorf("expected pinned fastapi requirement, got:\n%s", text)
	}
}

func TestGetVersion_UnknownPackageFailsRender(t *testing.T) {
	loader := templates.NewTemplateLoader("")
	data := &templates.TemplateData{Versions: map[string]string{}}

	if _, err := loader.RenderTemplate("express", "package.json.tmpl", data); err == nil {
		t.Fatal("expected rendering package.json without versions to fail, got nil error")
	}
}
