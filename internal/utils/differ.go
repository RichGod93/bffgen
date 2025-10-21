package utils

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/mitchellh/colorstring"
)

// DiffLine represents a single line in a diff
type DiffLine struct {
	Type    DiffLineType
	LineNum int
	Content string
}

// DiffLineType represents the type of diff line
type DiffLineType int

const (
	DiffLineContext DiffLineType = iota
	DiffLineAdded
	DiffLineRemoved
	DiffLineModified
)

// ColorizedDiff generates a colorized git-style diff
type ColorizedDiff struct {
	FilePath     string
	OldContent   string
	NewContent   string
	ContextLines int // Number of context lines to show
}

// NewColorizedDiff creates a new colorized diff
func NewColorizedDiff(filePath, oldContent, newContent string) *ColorizedDiff {
	return &ColorizedDiff{
		FilePath:     filePath,
		OldContent:   oldContent,
		NewContent:   newContent,
		ContextLines: 3,
	}
}

// Generate generates the colorized diff output
func (cd *ColorizedDiff) Generate() string {
	oldLines := strings.Split(cd.OldContent, "\n")
	newLines := strings.Split(cd.NewContent, "\n")

	var buf bytes.Buffer

	// File header
	buf.WriteString(colorstring.Color(fmt.Sprintf("[bold]diff --git a/%s b/%s[reset]\n", cd.FilePath, cd.FilePath)))

	if cd.OldContent == "" {
		// New file
		buf.WriteString(colorstring.Color("[green]--- /dev/null[reset]\n"))
		buf.WriteString(colorstring.Color(fmt.Sprintf("[green]+++ b/%s[reset]\n", cd.FilePath)))
		buf.WriteString(colorstring.Color("[cyan]@@ -0,0 +1," + fmt.Sprintf("%d", len(newLines)) + " @@[reset]\n"))

		for _, line := range newLines {
			buf.WriteString(colorstring.Color(fmt.Sprintf("[green]+%s[reset]\n", line)))
		}
	} else if cd.NewContent == "" {
		// Deleted file
		buf.WriteString(colorstring.Color(fmt.Sprintf("[red]--- a/%s[reset]\n", cd.FilePath)))
		buf.WriteString(colorstring.Color("[red]+++ /dev/null[reset]\n"))
		buf.WriteString(colorstring.Color("[cyan]@@ -1," + fmt.Sprintf("%d", len(oldLines)) + " +0,0 @@[reset]\n"))

		for _, line := range oldLines {
			buf.WriteString(colorstring.Color(fmt.Sprintf("[red]-%s[reset]\n", line)))
		}
	} else {
		// Modified file
		buf.WriteString(colorstring.Color(fmt.Sprintf("[red]--- a/%s[reset]\n", cd.FilePath)))
		buf.WriteString(colorstring.Color(fmt.Sprintf("[green]+++ b/%s[reset]\n", cd.FilePath)))

		// Compute line-by-line diff
		diffLines := cd.computeLineDiff(oldLines, newLines)

		// Group diff into hunks
		hunks := cd.groupIntoHunks(diffLines)

		// Render each hunk
		for _, hunk := range hunks {
			cd.renderHunk(&buf, hunk, oldLines, newLines)
		}
	}

	return buf.String()
}

// computeLineDiff computes line-by-line differences
func (cd *ColorizedDiff) computeLineDiff(oldLines, newLines []string) []DiffLine {
	var diffLines []DiffLine

	// Simple LCS-based diff algorithm
	oldMap := make(map[string][]int)
	for i, line := range oldLines {
		oldMap[line] = append(oldMap[line], i)
	}

	used := make(map[int]bool)
	matches := make(map[int]int)

	// Find matching lines
	for newIdx, newLine := range newLines {
		if oldIndexes, ok := oldMap[newLine]; ok {
			for _, oldIdx := range oldIndexes {
				if !used[oldIdx] {
					matches[newIdx] = oldIdx
					used[oldIdx] = true
					break
				}
			}
		}
	}

	// Build diff lines
	oldIdx, newIdx := 0, 0
	for oldIdx < len(oldLines) || newIdx < len(newLines) {
		if oldIdx < len(oldLines) && newIdx < len(newLines) {
			if matchedOld, ok := matches[newIdx]; ok && matchedOld == oldIdx {
				// Lines match (context)
				diffLines = append(diffLines, DiffLine{
					Type:    DiffLineContext,
					LineNum: newIdx + 1,
					Content: newLines[newIdx],
				})
				oldIdx++
				newIdx++
			} else if !used[oldIdx] {
				// Line removed
				diffLines = append(diffLines, DiffLine{
					Type:    DiffLineRemoved,
					LineNum: oldIdx + 1,
					Content: oldLines[oldIdx],
				})
				oldIdx++
			} else {
				// Line added
				diffLines = append(diffLines, DiffLine{
					Type:    DiffLineAdded,
					LineNum: newIdx + 1,
					Content: newLines[newIdx],
				})
				newIdx++
			}
		} else if oldIdx < len(oldLines) {
			// Remaining lines removed
			diffLines = append(diffLines, DiffLine{
				Type:    DiffLineRemoved,
				LineNum: oldIdx + 1,
				Content: oldLines[oldIdx],
			})
			oldIdx++
		} else {
			// Remaining lines added
			diffLines = append(diffLines, DiffLine{
				Type:    DiffLineAdded,
				LineNum: newIdx + 1,
				Content: newLines[newIdx],
			})
			newIdx++
		}
	}

	return diffLines
}

// DiffHunk represents a chunk of related changes
type DiffHunk struct {
	OldStart int
	OldCount int
	NewStart int
	NewCount int
	Lines    []DiffLine
}

// groupIntoHunks groups diff lines into hunks
func (cd *ColorizedDiff) groupIntoHunks(diffLines []DiffLine) []DiffHunk {
	if len(diffLines) == 0 {
		return nil
	}

	var hunks []DiffHunk
	var currentHunk *DiffHunk

	for i, line := range diffLines {
		// Start new hunk if we're at a change or near previous changes
		if line.Type != DiffLineContext || currentHunk == nil {
			if currentHunk == nil || (i > 0 && diffLines[i-1].Type == DiffLineContext && cd.ContextLines > 0) {
				if currentHunk != nil {
					hunks = append(hunks, *currentHunk)
				}
				currentHunk = &DiffHunk{
					OldStart: line.LineNum,
					NewStart: line.LineNum,
					Lines:    []DiffLine{},
				}
			}
		}

		if currentHunk != nil {
			currentHunk.Lines = append(currentHunk.Lines, line)
			if line.Type == DiffLineRemoved {
				currentHunk.OldCount++
			} else if line.Type == DiffLineAdded {
				currentHunk.NewCount++
			} else {
				currentHunk.OldCount++
				currentHunk.NewCount++
			}
		}
	}

	if currentHunk != nil {
		hunks = append(hunks, *currentHunk)
	}

	return hunks
}

// renderHunk renders a single hunk with colorization
func (cd *ColorizedDiff) renderHunk(buf *bytes.Buffer, hunk DiffHunk, oldLines, newLines []string) {
	// Hunk header
	buf.WriteString(colorstring.Color(fmt.Sprintf(
		"[cyan]@@ -%d,%d +%d,%d @@[reset]\n",
		hunk.OldStart, hunk.OldCount,
		hunk.NewStart, hunk.NewCount,
	)))

	// Render lines
	for _, line := range hunk.Lines {
		switch line.Type {
		case DiffLineAdded:
			buf.WriteString(colorstring.Color(fmt.Sprintf("[green]+%s[reset]\n", line.Content)))
		case DiffLineRemoved:
			buf.WriteString(colorstring.Color(fmt.Sprintf("[red]-%s[reset]\n", line.Content)))
		case DiffLineContext:
			buf.WriteString(fmt.Sprintf(" %s\n", line.Content))
		}
	}
}

// GenerateSummary generates a summary of changes
func (cd *ColorizedDiff) GenerateSummary() string {
	oldLines := strings.Split(cd.OldContent, "\n")
	newLines := strings.Split(cd.NewContent, "\n")

	added := 0
	removed := 0

	// Count actual differences
	oldSet := make(map[string]bool)
	for _, line := range oldLines {
		oldSet[line] = true
	}

	newSet := make(map[string]bool)
	for _, line := range newLines {
		newSet[line] = true
		if !oldSet[line] {
			added++
		}
	}

	for _, line := range oldLines {
		if !newSet[line] {
			removed++
		}
	}

	if added == 0 && removed == 0 {
		return "No changes"
	}

	var parts []string
	if added > 0 {
		parts = append(parts, colorstring.Color(fmt.Sprintf("[green]+%d[reset]", added)))
	}
	if removed > 0 {
		parts = append(parts, colorstring.Color(fmt.Sprintf("[red]-%d[reset]", removed)))
	}

	return strings.Join(parts, " ")
}

// MultiFileDiff handles diffs for multiple files
type MultiFileDiff struct {
	Files map[string]*ColorizedDiff
}

// NewMultiFileDiff creates a new multi-file diff
func NewMultiFileDiff() *MultiFileDiff {
	return &MultiFileDiff{
		Files: make(map[string]*ColorizedDiff),
	}
}

// AddFile adds a file to the diff
func (mfd *MultiFileDiff) AddFile(path, oldContent, newContent string) {
	mfd.Files[path] = NewColorizedDiff(path, oldContent, newContent)
}

// Generate generates the full multi-file diff
func (mfd *MultiFileDiff) Generate() string {
	var buf bytes.Buffer

	for _, diff := range mfd.Files {
		buf.WriteString(diff.Generate())
		buf.WriteString("\n")
	}

	return buf.String()
}

// GenerateSummary generates a summary for all files
func (mfd *MultiFileDiff) GenerateSummary() string {
	var buf bytes.Buffer

	totalAdded := 0
	totalRemoved := 0
	filesChanged := 0

	for filePath, diff := range mfd.Files {
		oldLines := strings.Split(diff.OldContent, "\n")
		newLines := strings.Split(diff.NewContent, "\n")

		if len(oldLines) != len(newLines) || diff.OldContent != diff.NewContent {
			filesChanged++

			// Count changes
			for _, line := range newLines {
				found := false
				for _, oldLine := range oldLines {
					if line == oldLine {
						found = true
						break
					}
				}
				if !found && line != "" {
					totalAdded++
				}
			}

			for _, line := range oldLines {
				found := false
				for _, newLine := range newLines {
					if line == newLine {
						found = true
						break
					}
				}
				if !found && line != "" {
					totalRemoved++
				}
			}

			buf.WriteString(fmt.Sprintf("  %s: %s\n", filePath, diff.GenerateSummary()))
		}
	}

	summary := colorstring.Color(fmt.Sprintf(
		"\n[bold]%d file(s) changed[reset]: [green]+%d[reset] [red]-%d[reset] lines\n",
		filesChanged, totalAdded, totalRemoved,
	))

	return summary + buf.String()
}
