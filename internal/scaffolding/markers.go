package scaffolding

import (
	"fmt"
	"regexp"
	"strings"
)

// CodeMarker represents a code fence marker for generated sections
type CodeMarker struct {
	Begin string
	End   string
}

// DefaultMarkers returns the default bffgen code markers
func DefaultMarkers() CodeMarker {
	return CodeMarker{
		Begin: "// bffgen:begin",
		End:   "// bffgen:end",
	}
}

// CustomMarkers creates custom markers with a specific identifier
func CustomMarkers(identifier string) CodeMarker {
	return CodeMarker{
		Begin: fmt.Sprintf("// bffgen:begin:%s", identifier),
		End:   fmt.Sprintf("// bffgen:end:%s", identifier),
	}
}

// MarkerPattern represents a compiled regex pattern for finding markers
type MarkerPattern struct {
	Begin *regexp.Regexp
	End   *regexp.Regexp
}

// CompilePatterns compiles regex patterns for the markers
func (m CodeMarker) CompilePatterns() (*MarkerPattern, error) {
	beginPattern, err := regexp.Compile(fmt.Sprintf(`^\s*%s\s*$`, regexp.QuoteMeta(m.Begin)))
	if err != nil {
		return nil, fmt.Errorf("failed to compile begin pattern: %w", err)
	}

	endPattern, err := regexp.Compile(fmt.Sprintf(`^\s*%s\s*$`, regexp.QuoteMeta(m.End)))
	if err != nil {
		return nil, fmt.Errorf("failed to compile end pattern: %w", err)
	}

	return &MarkerPattern{
		Begin: beginPattern,
		End:   endPattern,
	}, nil
}

// CodeSection represents a section of code between markers
type CodeSection struct {
	BeginLine int    // Line number where the begin marker is found
	EndLine   int    // Line number where the end marker is found
	Content   string // The content between the markers
	Marker    CodeMarker
}

// FindSections finds all code sections marked with the given markers
func FindSections(content string, marker CodeMarker) ([]CodeSection, error) {
	lines := strings.Split(content, "\n")
	pattern, err := marker.CompilePatterns()
	if err != nil {
		return nil, err
	}

	var sections []CodeSection
	var currentSection *CodeSection

	for i, line := range lines {
		// Check for begin marker
		if pattern.Begin.MatchString(line) {
			if currentSection != nil {
				return nil, fmt.Errorf("nested begin marker found at line %d", i+1)
			}
			currentSection = &CodeSection{
				BeginLine: i + 1,
				Marker:    marker,
			}
			continue
		}

		// Check for end marker
		if pattern.End.MatchString(line) {
			if currentSection == nil {
				return nil, fmt.Errorf("end marker without begin marker found at line %d", i+1)
			}
			currentSection.EndLine = i + 1
			sections = append(sections, *currentSection)
			currentSection = nil
			continue
		}

		// Add content to current section
		if currentSection != nil {
			if currentSection.Content != "" {
				currentSection.Content += "\n"
			}
			currentSection.Content += line
		}
	}

	// Check for unclosed section
	if currentSection != nil {
		return nil, fmt.Errorf("unclosed section starting at line %d", currentSection.BeginLine)
	}

	return sections, nil
}

// ReplaceSection replaces a code section with new content
func ReplaceSection(content string, section CodeSection, newContent string) (string, error) {
	lines := strings.Split(content, "\n")
	
	if section.BeginLine < 1 || section.EndLine > len(lines) {
		return "", fmt.Errorf("section boundaries out of range")
	}

	// Build new content
	var newLines []string
	
	// Add lines before the section
	newLines = append(newLines, lines[:section.BeginLine-1]...)
	
	// Add begin marker
	newLines = append(newLines, section.Marker.Begin)
	
	// Add new content
	if newContent != "" {
		newLines = append(newLines, strings.Split(newContent, "\n")...)
	}
	
	// Add end marker
	newLines = append(newLines, section.Marker.End)
	
	// Add lines after the section
	newLines = append(newLines, lines[section.EndLine:]...)

	return strings.Join(newLines, "\n"), nil
}

// InsertSection inserts a new code section at the specified location
func InsertSection(content string, marker CodeMarker, newContent string, insertAfterLine int) (string, error) {
	lines := strings.Split(content, "\n")
	
	if insertAfterLine < 0 || insertAfterLine > len(lines) {
		return "", fmt.Errorf("insert position out of range")
	}

	// Build new content
	var newLines []string
	
	// Add lines before insertion point
	newLines = append(newLines, lines[:insertAfterLine]...)
	
	// Add begin marker
	newLines = append(newLines, marker.Begin)
	
	// Add new content
	if newContent != "" {
		newLines = append(newLines, strings.Split(newContent, "\n")...)
	}
	
	// Add end marker
	newLines = append(newLines, marker.End)
	
	// Add lines after insertion point
	newLines = append(newLines, lines[insertAfterLine:]...)

	return strings.Join(newLines, "\n"), nil
}

// RemoveSection removes a code section and its markers
func RemoveSection(content string, section CodeSection) (string, error) {
	lines := strings.Split(content, "\n")
	
	if section.BeginLine < 1 || section.EndLine > len(lines) {
		return "", fmt.Errorf("section boundaries out of range")
	}

	// Build new content without the section
	var newLines []string
	
	// Add lines before the section
	newLines = append(newLines, lines[:section.BeginLine-1]...)
	
	// Add lines after the section
	newLines = append(newLines, lines[section.EndLine:]...)

	return strings.Join(newLines, "\n"), nil
}

// ValidateMarkers validates that all markers in content are properly paired
func ValidateMarkers(content string, marker CodeMarker) error {
	_, err := FindSections(content, marker)
	return err
}

// GetMarkerSummary returns a summary of all markers found in content
func GetMarkerSummary(content string, marker CodeMarker) ([]string, error) {
	sections, err := FindSections(content, marker)
	if err != nil {
		return nil, err
	}

	var summary []string
	for _, section := range sections {
		summary = append(summary, fmt.Sprintf("Section at lines %d-%d (%d lines)", 
			section.BeginLine, section.EndLine, section.EndLine-section.BeginLine-1))
	}

	return summary, nil
}
