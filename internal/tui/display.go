package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("9")).
			Padding(1, 2).
			Margin(1, 0)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("11")).
			Padding(1, 2).
			Margin(1, 0)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("10")).
			Padding(1, 2).
			Margin(1, 0)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("12")).
			Padding(1, 2).
			Margin(1, 0)
)

// DisplayError shows a formatted error message
func DisplayError(title string, err error) {
	var b strings.Builder
	b.WriteString("❌ " + title + "\n\n")
	b.WriteString("Error: " + err.Error())

	fmt.Println(errorStyle.Render(b.String()))
}

// DisplayErrorWithSuggestions shows an error with helpful suggestions
func DisplayErrorWithSuggestions(title string, err error, suggestions []string) {
	var b strings.Builder
	b.WriteString("❌ " + title + "\n\n")
	b.WriteString("Error: " + err.Error())

	if len(suggestions) > 0 {
		b.WriteString("\n\n💡 Suggestions:\n")
		for _, suggestion := range suggestions {
			b.WriteString("  • " + suggestion + "\n")
		}
	}

	fmt.Println(errorStyle.Render(b.String()))
}

// DisplayWarning shows a warning message
func DisplayWarning(title string, message string) {
	var b strings.Builder
	b.WriteString("⚠️  " + title + "\n\n")
	b.WriteString(message)

	fmt.Println(warningStyle.Render(b.String()))
}

// DisplaySuccess shows a success message
func DisplaySuccess(title string, message string) {
	var b strings.Builder
	b.WriteString("✅ " + title + "\n\n")
	b.WriteString(message)

	fmt.Println(successStyle.Render(b.String()))
}

// DisplayInfo shows an informational message
func DisplayInfo(title string, message string) {
	var b strings.Builder
	b.WriteString("ℹ️  " + title + "\n\n")
	b.WriteString(message)

	fmt.Println(infoStyle.Render(b.String()))
}

// DisplayValidationErrors shows multiple validation errors
func DisplayValidationErrors(errors []string) {
	var b strings.Builder
	b.WriteString("❌ Validation Failed\n\n")
	b.WriteString(fmt.Sprintf("Found %d error(s):\n\n", len(errors)))

	for i, err := range errors {
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, err))
	}

	fmt.Println(errorStyle.Render(b.String()))
}

// DisplayTable shows data in a formatted table
func DisplayTable(headers []string, rows [][]string) {
	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Create separator
	var separator strings.Builder
	for i, w := range widths {
		separator.WriteString(strings.Repeat("─", w+2))
		if i < len(widths)-1 {
			separator.WriteString("┼")
		}
	}

	// Print header
	var header strings.Builder
	for i, h := range headers {
		header.WriteString(" " + h + strings.Repeat(" ", widths[i]-len(h)+1))
		if i < len(headers)-1 {
			header.WriteString("│")
		}
	}

	fmt.Println(header.String())
	fmt.Println(separator.String())

	// Print rows
	for _, row := range rows {
		var rowStr strings.Builder
		for i, cell := range row {
			if i < len(widths) {
				rowStr.WriteString(" " + cell + strings.Repeat(" ", widths[i]-len(cell)+1))
				if i < len(row)-1 {
					rowStr.WriteString("│")
				}
			}
		}
		fmt.Println(rowStr.String())
	}
}
