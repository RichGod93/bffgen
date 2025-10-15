package utils

import (
	"strings"
	"testing"
)

func TestColorizedDiff(t *testing.T) {
	t.Run("NewFile", func(t *testing.T) {
		diff := NewColorizedDiff("test.txt", "", "line1\nline2\nline3")

		output := diff.Generate()
		if output == "" {
			t.Error("Diff output should not be empty")
		}

		if !strings.Contains(output, "test.txt") {
			t.Error("Diff should contain filename")
		}

		if !strings.Contains(output, "/dev/null") {
			t.Error("New file should show /dev/null as old")
		}
	})

	t.Run("DeletedFile", func(t *testing.T) {
		diff := NewColorizedDiff("test.txt", "line1\nline2", "")

		output := diff.Generate()
		if !strings.Contains(output, "/dev/null") {
			t.Error("Deleted file should show /dev/null as new")
		}
	})

	t.Run("ModifiedFile", func(t *testing.T) {
		old := "line1\nline2\nline3"
		new := "line1\nmodified\nline3"

		diff := NewColorizedDiff("test.txt", old, new)

		output := diff.Generate()
		if output == "" {
			t.Error("Diff output should not be empty")
		}
	})

	t.Run("NoChanges", func(t *testing.T) {
		content := "line1\nline2"
		diff := NewColorizedDiff("test.txt", content, content)

		summary := diff.GenerateSummary()
		if summary != "No changes" {
			t.Errorf("Expected 'No changes', got '%s'", summary)
		}
	})

	t.Run("GenerateSummary", func(t *testing.T) {
		old := "line1\nline2"
		new := "line1\nline2\nline3"

		diff := NewColorizedDiff("test.txt", old, new)

		summary := diff.GenerateSummary()
		if summary == "" {
			t.Error("Summary should not be empty")
		}

		// Should indicate addition
		if !strings.Contains(summary, "+") {
			t.Error("Summary should show additions")
		}
	})
}

func TestMultiFileDiff(t *testing.T) {
	t.Run("AddAndGenerate", func(t *testing.T) {
		mfd := NewMultiFileDiff()

		mfd.AddFile("file1.txt", "old1", "new1")
		mfd.AddFile("file2.txt", "old2", "new2")

		if len(mfd.Files) != 2 {
			t.Errorf("Expected 2 files, got %d", len(mfd.Files))
		}

		output := mfd.Generate()
		if output == "" {
			t.Error("Multi-file diff should not be empty")
		}
	})

	t.Run("GenerateSummary", func(t *testing.T) {
		mfd := NewMultiFileDiff()

		mfd.AddFile("file1.txt", "old", "new")
		mfd.AddFile("file2.txt", "same", "same")

		summary := mfd.GenerateSummary()
		if summary == "" {
			t.Error("Summary should not be empty")
		}

		// Should show file count
		if !strings.Contains(summary, "file(s)") {
			t.Error("Summary should mention files")
		}
	})
}
