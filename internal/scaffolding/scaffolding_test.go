package scaffolding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodeMarker(t *testing.T) {
	marker := DefaultMarkers()

	if marker.Begin != "// bffgen:begin" {
		t.Errorf("Expected begin marker '// bffgen:begin', got '%s'", marker.Begin)
	}

	if marker.End != "// bffgen:end" {
		t.Errorf("Expected end marker '// bffgen:end', got '%s'", marker.End)
	}
}

func TestCustomMarkers(t *testing.T) {
	marker := CustomMarkers("routes")

	expectedBegin := "// bffgen:begin:routes"
	expectedEnd := "// bffgen:end:routes"

	if marker.Begin != expectedBegin {
		t.Errorf("Expected begin marker '%s', got '%s'", expectedBegin, marker.Begin)
	}

	if marker.End != expectedEnd {
		t.Errorf("Expected end marker '%s', got '%s'", expectedEnd, marker.End)
	}
}

func TestFindSections(t *testing.T) {
	content := `package main

import "fmt"

// bffgen:begin
func generatedFunction() {
	fmt.Println("This is generated code")
}
// bffgen:end

func main() {
	fmt.Println("Hello, World!")
}
`

	marker := DefaultMarkers()
	sections, err := FindSections(content, marker)
	if err != nil {
		t.Fatalf("Failed to find sections: %v", err)
	}

	if len(sections) != 1 {
		t.Errorf("Expected 1 section, got %d", len(sections))
	}

	section := sections[0]
	if section.BeginLine != 5 {
		t.Errorf("Expected begin line 5, got %d", section.BeginLine)
	}

	if section.EndLine != 9 {
		t.Errorf("Expected end line 9, got %d", section.EndLine)
	}

	expectedContent := `func generatedFunction() {
	fmt.Println("This is generated code")
}`

	if strings.TrimSpace(section.Content) != expectedContent {
		t.Errorf("Expected content '%s', got '%s'", expectedContent, section.Content)
	}
}

func TestReplaceSection(t *testing.T) {
	content := `package main

import "fmt"

// bffgen:begin
func oldFunction() {
	fmt.Println("Old code")
}
// bffgen:end

func main() {
	fmt.Println("Hello, World!")
}
`

	marker := DefaultMarkers()
	sections, err := FindSections(content, marker)
	if err != nil {
		t.Fatalf("Failed to find sections: %v", err)
	}

	newContent := `func newFunction() {
	fmt.Println("New code")
}`

	updated, err := ReplaceSection(content, sections[0], newContent)
	if err != nil {
		t.Fatalf("Failed to replace section: %v", err)
	}

	expected := `package main

import "fmt"

// bffgen:begin
func newFunction() {
	fmt.Println("New code")
}
// bffgen:end

func main() {
	fmt.Println("Hello, World!")
}
`

	if updated != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, updated)
	}
}

func TestInsertSection(t *testing.T) {
	content := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`

	marker := DefaultMarkers()
	newContent := `func generatedFunction() {
	fmt.Println("Generated code")
}`

	updated, err := InsertSection(content, marker, newContent, 3)
	if err != nil {
		t.Fatalf("Failed to insert section: %v", err)
	}

	expected := `package main

import "fmt"
// bffgen:begin
func generatedFunction() {
	fmt.Println("Generated code")
}
// bffgen:end

func main() {
	fmt.Println("Hello, World!")
}
`

	if updated != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, updated)
	}
}

func TestComputeDiff(t *testing.T) {
	oldContent := `line 1
line 2
line 3`

	newContent := `line 1
line 2 modified
line 3
line 4`

	diff := ComputeDiff(oldContent, newContent)

	if !diff.HasChanges {
		t.Error("Expected changes to be detected")
	}

	if len(diff.Diffs) != 2 {
		t.Errorf("Expected 2 diffs, got %d", len(diff.Diffs))
	}

	// Check first diff (modified line)
	if diff.Diffs[0].Type != DiffModified {
		t.Errorf("Expected first diff to be modified, got %v", diff.Diffs[0].Type)
	}

	// Check second diff (added line)
	if diff.Diffs[1].Type != DiffAdded {
		t.Errorf("Expected second diff to be added, got %v", diff.Diffs[1].Type)
	}
}

func TestThreeWayMerge(t *testing.T) {
	base := `line 1
line 2
line 3`

	local := `line 1
line 2 modified by user
line 3`

	remote := `line 1
line 2
line 3
line 4 added by generator`

	twm := &ThreeWayMerge{
		Base:   base,
		Local:  local,
		Remote: remote,
	}

	result := twm.PerformMerge()

	if result.HasConflicts {
		t.Error("Expected no conflicts")
	}

	expected := `line 1
line 2 modified by user
line 3
line 4 added by generator`

	if strings.TrimSpace(result.Content) != strings.TrimSpace(expected) {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, result.Content)
	}
}

func TestGenerator(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.go")

	generator := NewGenerator()
	// Verbose mode disabled in tests to avoid cluttering CI output
	generator.SetVerbose(false)

	// Test generating a new file
	newContent := `func generatedFunction() {
	fmt.Println("Generated code")
}`

	err := generator.GenerateFile(filePath, newContent)
	if err != nil {
		t.Fatalf("Failed to generate file: %v", err)
	}

	// Verify file was created
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	expected := `// bffgen:begin
func generatedFunction() {
	fmt.Println("Generated code")
}
// bffgen:end`

	if strings.TrimSpace(string(content)) != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, string(content))
	}

	// Test updating the file
	updatedContent := `func updatedFunction() {
	fmt.Println("Updated code")
}`

	err = generator.GenerateFile(filePath, updatedContent)
	if err != nil {
		t.Fatalf("Failed to update file: %v", err)
	}

	// Verify file was updated
	content, err = os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	expected = `// bffgen:begin
func updatedFunction() {
	fmt.Println("Updated code")
}
// bffgen:end`

	if strings.TrimSpace(string(content)) != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, string(content))
	}
}

func TestGeneratorDryRun(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.go")

	generator := NewGenerator()
	generator.SetDryRun(true)
	generator.SetVerbose(false) // Disabled to avoid cluttering CI output

	// Test dry run
	newContent := `func generatedFunction() {
	fmt.Println("Generated code")
}`

	err := generator.GenerateFile(filePath, newContent)
	if err != nil {
		t.Fatalf("Failed to generate file in dry run: %v", err)
	}

	// Verify file was NOT created
	if _, err := os.Stat(filePath); err == nil {
		t.Error("File should not exist in dry run mode")
	}
}

func TestGeneratorCheckMode(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.go")

	generator := NewGenerator()
	generator.SetCheckMode(true)
	generator.SetVerbose(false) // Disabled to avoid cluttering CI output

	// Test check mode
	newContent := `func generatedFunction() {
	fmt.Println("Generated code")
}`

	err := generator.GenerateFile(filePath, newContent)
	if err != nil {
		t.Fatalf("Failed to generate file in check mode: %v", err)
	}

	// Verify file was NOT created
	if _, err := os.Stat(filePath); err == nil {
		t.Error("File should not exist in check mode")
	}
}

func TestValidateMarkers(t *testing.T) {
	validContent := `package main

// bffgen:begin
func generated() {}
// bffgen:end
`

	invalidContent := `package main

// bffgen:begin
func generated() {}
// Missing end marker
`

	marker := DefaultMarkers()

	// Test valid content
	err := ValidateMarkers(validContent, marker)
	if err != nil {
		t.Errorf("Valid content should pass validation: %v", err)
	}

	// Test invalid content
	err = ValidateMarkers(invalidContent, marker)
	if err == nil {
		t.Error("Invalid content should fail validation")
	}
}

func TestGetMarkerSummary(t *testing.T) {
	content := `package main

// bffgen:begin
func first() {}
// bffgen:end

// bffgen:begin
func second() {}
// bffgen:end
`

	marker := DefaultMarkers()
	summary, err := GetMarkerSummary(content, marker)
	if err != nil {
		t.Fatalf("Failed to get marker summary: %v", err)
	}

	if len(summary) != 2 {
		t.Errorf("Expected 2 sections, got %d", len(summary))
	}
}
