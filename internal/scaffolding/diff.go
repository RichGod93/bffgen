package scaffolding

import (
	"bytes"
	"fmt"
	"strings"
)

// Diff represents a difference between two pieces of content
type Diff struct {
	Type    DiffType
	Line    int
	Content string
}

// DiffType represents the type of difference
type DiffType int

const (
	DiffAdded DiffType = iota
	DiffRemoved
	DiffModified
)

func (dt DiffType) String() string {
	switch dt {
	case DiffAdded:
		return "+"
	case DiffRemoved:
		return "-"
	case DiffModified:
		return "~"
	default:
		return "?"
	}
}

// DiffResult represents the result of a diff operation
type DiffResult struct {
	HasChanges bool
	Diffs      []Diff
	Summary    string
}

// ComputeDiff computes the differences between two pieces of content
func ComputeDiff(oldContent, newContent string) *DiffResult {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")
	
	var diffs []Diff
	hasChanges := false

	// Simple line-by-line diff
	maxLines := len(oldLines)
	if len(newLines) > maxLines {
		maxLines = len(newLines)
	}

	for i := 0; i < maxLines; i++ {
		var oldLine, newLine string
		if i < len(oldLines) {
			oldLine = oldLines[i]
		}
		if i < len(newLines) {
			newLine = newLines[i]
		}

		if oldLine != newLine {
			hasChanges = true
			
			if i >= len(oldLines) {
				// Line added
				diffs = append(diffs, Diff{
					Type:    DiffAdded,
					Line:    i + 1,
					Content: newLine,
				})
			} else if i >= len(newLines) {
				// Line removed
				diffs = append(diffs, Diff{
					Type:    DiffRemoved,
					Line:    i + 1,
					Content: oldLine,
				})
			} else {
				// Line modified
				diffs = append(diffs, Diff{
					Type:    DiffModified,
					Line:    i + 1,
					Content: fmt.Sprintf("%s -> %s", oldLine, newLine),
				})
			}
		}
	}

	summary := fmt.Sprintf("%d changes", len(diffs))
	if len(diffs) > 0 {
		added := 0
		removed := 0
		modified := 0
		
		for _, diff := range diffs {
			switch diff.Type {
			case DiffAdded:
				added++
			case DiffRemoved:
				removed++
			case DiffModified:
				modified++
			}
		}
		
		var parts []string
		if added > 0 {
			parts = append(parts, fmt.Sprintf("%d added", added))
		}
		if removed > 0 {
			parts = append(parts, fmt.Sprintf("%d removed", removed))
		}
		if modified > 0 {
			parts = append(parts, fmt.Sprintf("%d modified", modified))
		}
		
		summary = strings.Join(parts, ", ")
	}

	return &DiffResult{
		HasChanges: hasChanges,
		Diffs:      diffs,
		Summary:    summary,
	}
}

// FormatDiff formats the diff result as a string
func (dr *DiffResult) FormatDiff() string {
	if !dr.HasChanges {
		return "No changes detected"
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("Changes: %s\n", dr.Summary))
	buf.WriteString(strings.Repeat("-", 50) + "\n")

	for _, diff := range dr.Diffs {
		buf.WriteString(fmt.Sprintf("%s %d: %s\n", diff.Type.String(), diff.Line, diff.Content))
	}

	return buf.String()
}

// ThreeWayMerge performs a 3-way merge between base, local, and remote content
type ThreeWayMerge struct {
	Base   string // Original content
	Local  string // User's modifications
	Remote string // New generated content
}

// MergeResult represents the result of a 3-way merge
type MergeResult struct {
	Content    string
	Conflicts  []Conflict
	HasConflicts bool
	Summary    string
}

// Conflict represents a merge conflict
type Conflict struct {
	Line     int
	Base     string
	Local    string
	Remote   string
	Resolved string
}

// PerformMerge performs a 3-way merge
func (twm *ThreeWayMerge) PerformMerge() *MergeResult {
	// For now, implement a simple strategy:
	// 1. Use remote content as base
	// 2. Preserve user modifications outside of generated sections
	// 3. Report conflicts for overlapping changes

	baseLines := strings.Split(twm.Base, "\n")
	localLines := strings.Split(twm.Local, "\n")
	remoteLines := strings.Split(twm.Remote, "\n")

	var resultLines []string
	var conflicts []Conflict

	// Simple line-by-line merge
	maxLines := len(baseLines)
	if len(localLines) > maxLines {
		maxLines = len(localLines)
	}
	if len(remoteLines) > maxLines {
		maxLines = len(remoteLines)
	}

	for i := 0; i < maxLines; i++ {
		var baseLine, localLine, remoteLine string
		
		if i < len(baseLines) {
			baseLine = baseLines[i]
		}
		if i < len(localLines) {
			localLine = localLines[i]
		}
		if i < len(remoteLines) {
			remoteLine = remoteLines[i]
		}

		// Check for conflicts
		if localLine != baseLine && remoteLine != baseLine && localLine != remoteLine {
			// Conflict detected
			conflict := Conflict{
				Line:   i + 1,
				Base:   baseLine,
				Local:  localLine,
				Remote: remoteLine,
			}
			conflicts = append(conflicts, conflict)
			
			// For now, prefer remote (generated) content
			resultLines = append(resultLines, remoteLine)
		} else if localLine != baseLine {
			// Local modification, preserve it
			resultLines = append(resultLines, localLine)
		} else {
			// Use remote content
			resultLines = append(resultLines, remoteLine)
		}
	}

	summary := fmt.Sprintf("Merged %d lines", len(resultLines))
	if len(conflicts) > 0 {
		summary += fmt.Sprintf(", %d conflicts", len(conflicts))
	}

	return &MergeResult{
		Content:      strings.Join(resultLines, "\n"),
		Conflicts:    conflicts,
		HasConflicts: len(conflicts) > 0,
		Summary:      summary,
	}
}

// FormatConflicts formats the conflicts as a string
func (mr *MergeResult) FormatConflicts() string {
	if !mr.HasConflicts {
		return "No conflicts"
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("Merge Conflicts (%d):\n", len(mr.Conflicts)))
	buf.WriteString(strings.Repeat("=", 50) + "\n")

	for _, conflict := range mr.Conflicts {
		buf.WriteString(fmt.Sprintf("Line %d:\n", conflict.Line))
		buf.WriteString(fmt.Sprintf("  Base:   %s\n", conflict.Base))
		buf.WriteString(fmt.Sprintf("  Local:  %s\n", conflict.Local))
		buf.WriteString(fmt.Sprintf("  Remote: %s\n", conflict.Remote))
		buf.WriteString("\n")
	}

	return buf.String()
}

// ResolveConflict resolves a conflict by choosing a resolution
func (mr *MergeResult) ResolveConflict(line int, resolution string) {
	for i, conflict := range mr.Conflicts {
		if conflict.Line == line {
			mr.Conflicts[i].Resolved = resolution
			break
		}
	}
}

// ApplyResolutions applies all resolved conflicts to the content
func (mr *MergeResult) ApplyResolutions() string {
	if !mr.HasConflicts {
		return mr.Content
	}

	lines := strings.Split(mr.Content, "\n")
	
	for _, conflict := range mr.Conflicts {
		if conflict.Resolved != "" && conflict.Line <= len(lines) {
			lines[conflict.Line-1] = conflict.Resolved
		}
	}

	return strings.Join(lines, "\n")
}
