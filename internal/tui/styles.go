package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Color palette
	primaryColor   = lipgloss.AdaptiveColor{Light: "#5B21B6", Dark: "#A78BFA"}
	secondaryColor = lipgloss.AdaptiveColor{Light: "#0891B2", Dark: "#22D3EE"}
	successColor   = lipgloss.AdaptiveColor{Light: "#059669", Dark: "#10B981"}
	errorColor     = lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#EF4444"}
	warningColor   = lipgloss.AdaptiveColor{Light: "#D97706", Dark: "#F59E0B"}
	mutedColor     = lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9CA3AF"}
	subtleColor    = lipgloss.AdaptiveColor{Light: "#F3F4F6", Dark: "#1F2937"}

	// Base styles
	BaseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor).
			Padding(1, 2)

	// Header styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Italic(true)

	// Selection styles
	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true).
				PaddingLeft(2).
				Render("▶ ")

	UnselectedItemStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				PaddingLeft(4)

	// Status styles
	SuccessStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(warningColor)

	// Input styles
	FocusedInputStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Padding(0, 1)

	BlurredInputStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(mutedColor).
				Padding(0, 1)

	// Help text styles
	HelpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true).
			MarginTop(1)

	// Box styles
	PreviewBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(secondaryColor).
			Padding(1, 2).
			MarginTop(1)

	ValidationBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(errorColor).
				Padding(0, 1)

	// List styles
	ListHeaderStyle = lipgloss.NewStyle().
			Background(primaryColor).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 1).
			MarginBottom(1)

	ListItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	// Progress styles
	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(primaryColor)

	ProgressEmptyStyle = lipgloss.NewStyle().
				Foreground(mutedColor)
)

// RenderHeader renders a styled header with title and subtitle
func RenderHeader(title, subtitle string) string {
	header := TitleStyle.Render(title)
	if subtitle != "" {
		header += "\n" + SubtitleStyle.Render(subtitle)
	}
	return header + "\n\n"
}

// RenderSuccess renders a success message
func RenderSuccess(message string) string {
	return SuccessStyle.Render("✅ " + message)
}

// RenderError renders an error message
func RenderError(message string) string {
	return ErrorStyle.Render("❌ " + message)
}

// RenderWarning renders a warning message
func RenderWarning(message string) string {
	return WarningStyle.Render("⚠️  " + message)
}

// RenderHelp renders help text with keyboard shortcuts
func RenderHelp(shortcuts ...string) string {
	var help string
	for i, shortcut := range shortcuts {
		if i > 0 {
			help += " • "
		}
		help += shortcut
	}
	return HelpStyle.Render(help)
}

// RenderProgressBar renders a progress bar
func RenderProgressBar(current, total int, width int) string {
	if total == 0 {
		return ""
	}

	percentage := float64(current) / float64(total)
	filled := int(percentage * float64(width))
	empty := width - filled

	bar := ""
	for i := 0; i < filled; i++ {
		bar += ProgressBarStyle.Render("█")
	}
	for i := 0; i < empty; i++ {
		bar += ProgressEmptyStyle.Render("░")
	}

	label := lipgloss.NewStyle().
		Foreground(mutedColor).
		Render(fmt.Sprintf(" %d/%d", current, total))

	return bar + label
}
