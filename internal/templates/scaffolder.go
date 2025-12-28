package templates

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Scaffolder handles template scaffolding
type Scaffolder struct {
	template  *Template
	variables map[string]string
	data      interface{} // Optional complex data for text/template execution
}

// NewScaffolder creates a new template scaffolder
func NewScaffolder(template *Template, variables map[string]string) *Scaffolder {
	return &Scaffolder{
		template:  template,
		variables: variables,
		data:      variables, // Default data is variables map
	}
}

// SetData sets the Complex data for template execution
func (s *Scaffolder) SetData(data interface{}) {
	s.data = data
}

// OverlayFiles overlays template files on top of an existing project
func (s *Scaffolder) OverlayFiles(projectName, outputDir string) error {
	// Add PROJECT_NAME to variables if not present
	if _, ok := s.variables["PROJECT_NAME"]; !ok {
		s.variables["PROJECT_NAME"] = projectName
	}

	projectDir := filepath.Join(outputDir, projectName)
	templateSrcDir := filepath.Join(s.template.Path, "src")

	// Check if template src directory exists
	if _, err := os.Stat(templateSrcDir); err == nil {
		// Copy template files to project's src directory
		projectSrcDir := filepath.Join(projectDir, "src")
		if err := s.copyAndSubstitute(templateSrcDir, projectSrcDir); err != nil {
			return fmt.Errorf("failed to overlay template src files: %w", err)
		}
	}

	// Also overlay root files (package.json, .env, etc)
	files, err := os.ReadDir(s.template.Path)
	if err == nil {
		for _, file := range files {
			if file.Name() == "template.yaml" || file.Name() == "src" {
				continue
			}
			// Copy root file/dir
			srcPath := filepath.Join(s.template.Path, file.Name())
			dstPath := filepath.Join(projectDir, file.Name())
			if file.IsDir() {
				if err := s.copyAndSubstitute(srcPath, dstPath); err != nil {
					return fmt.Errorf("failed to overlay root directory %s: %w", file.Name(), err)
				}
			} else {
				info, _ := file.Info()
				if err := s.copyFileWithSubstitution(srcPath, dstPath, info.Mode()); err != nil {
					return fmt.Errorf("failed to overlay root file %s: %w", file.Name(), err)
				}
			}
		}
	}

	return nil
}

// Scaffold creates a project from the template (legacy method)
func (s *Scaffolder) Scaffold(projectName, outputDir string) error {
	// Add PROJECT_NAME to variables if not present
	if _, ok := s.variables["PROJECT_NAME"]; !ok {
		s.variables["PROJECT_NAME"] = projectName
	}

	// Create project directory
	projectDir := filepath.Join(outputDir, projectName)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Copy template files
	templateSrcDir := filepath.Join(s.template.Path, "src")
	if err := s.copyAndSubstitute(templateSrcDir, projectDir); err != nil {
		return fmt.Errorf("failed to copy template files: %w", err)
	}

	return nil
}

// copyAndSubstitute recursively copies files with variable substitution
func (s *Scaffolder) copyAndSubstitute(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy and substitute file
		return s.copyFileWithSubstitution(path, dstPath, info.Mode())
	})
}

// copyFileWithSubstitution copies a file and performs variable substitution
func (s *Scaffolder) copyFileWithSubstitution(src, dst string, mode os.FileMode) error {
	// Read source file
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Skip binary files
	if s.isBinaryFile(src) {
		// Just copy binary files without substitution
		return s.copyFileDirect(src, dst, mode)
	}

	// Perform variable substitution using text/template
	// First, try simple string replacement for {{VAR}} style (legacy/simple)
	sContent := string(content)
	replaced := s.replaceVariables(sContent)

	// Then, execute as Go template for {{ .Var }} style
	finalContent := replaced
	if strings.Contains(replaced, "{{") {
		// Define helper functions
		funcMap := template.FuncMap{
			"ToPascalCase": func(s string) string {
				words := strings.FieldsFunc(s, func(r rune) bool {
					return r == '-' || r == '_' || r == ' '
				})
				for i, word := range words {
					if len(word) > 0 {
						words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
					}
				}
				return strings.Join(words, "")
			},
			"ToCamelCase": func(s string) string {
				words := strings.FieldsFunc(s, func(r rune) bool {
					return r == '-' || r == '_' || r == ' '
				})
				for i, word := range words {
					if len(word) > 0 {
						words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
					}
				}
				pascal := strings.Join(words, "")
				if len(pascal) > 0 {
					return strings.ToLower(string(pascal[0])) + pascal[1:]
				}
				return pascal
			},
			"ToUpper": strings.ToUpper,
		}

		tmpl, err := template.New(filepath.Base(src)).Funcs(funcMap).Parse(replaced)
		if err != nil {
			// If parsing fails (e.g. conflict with other braces), fall back to simple replacement result
			// Log warning?
			finalContent = replaced
		} else {
			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, s.data); err == nil {
				finalContent = buf.String()
			}
			// If execution fails, we might just keep the replaced content
		}
	}

	// Write to destination
	return os.WriteFile(dst, []byte(finalContent), mode)
}

// copyFileDirect copies a file without modification
func (s *Scaffolder) copyFileDirect(src, dst string, mode os.FileMode) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// replaceVariables replaces {{VAR_NAME}} with actual values (Legacy support)
func (s *Scaffolder) replaceVariables(content string) string {
	result := content

	for name, value := range s.variables {
		// Replace {{VAR_NAME}}
		placeholder := fmt.Sprintf("{{%s}}", name)
		result = strings.ReplaceAll(result, placeholder, value)

		// Also support {{ .VAR_NAME }} if user wrote it that way but we are using string replace
		// (Though Go template execution below handles this better)
	}

	return result
}

// isBinaryFile checks if a file is binary
func (s *Scaffolder) isBinaryFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	binaryExts := []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".zip", ".tar", ".gz", ".exe", ".bin"}

	for _, binaryExt := range binaryExts {
		if ext == binaryExt {
			return true
		}
	}

	return false
}
