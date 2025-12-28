package tui

import (
	"fmt"
	"strings"

	"github.com/RichGod93/bffgen/internal/scaffolding"
	"github.com/charmbracelet/lipgloss"
)

// View renders the current screen
func (m Model) View() string {
	if m.cancelled {
		return RenderWarning("Setup cancelled by user\n")
	}

	if m.currentScreen == ScreenDone {
		return RenderSuccess("Configuration complete! Creating project...\n")
	}

	var screen string

	switch m.currentScreen {
	case ScreenLanguage:
		screen = m.viewLanguageSelection()
	case ScreenFramework:
		screen = m.viewFrameworkSelection()
	case ScreenCORS:
		screen = m.viewCORSConfig()
	case ScreenArchitecture:
		screen = m.viewArchitectureSelection()
	case ScreenServices:
		screen = m.viewServicesConfig()
	case ScreenRouteConfig:
		screen = m.viewRouteConfig()
	case ScreenConfirm:
		screen = m.viewConfirmation()
	default:
		screen = "Unknown screen"
	}

	// Add preview panel and help
	preview := m.viewPreviewPanel()
	help := m.viewHelpText()

	// Layout: screen on left, preview on right if enough width
	if m.width >= 100 {
		left := lipgloss.NewStyle().Width(m.width / 2).Render(screen)
		right := lipgloss.NewStyle().Width(m.width / 2).Render(preview)
		content := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
		return content + "\n\n" + help
	}

	// Stack vertically for narrow terminals
	return screen + "\n\n" + preview + "\n\n" + help
}

// viewLanguageSelection renders language selection screen
func (m Model) viewLanguageSelection() string {
	var b strings.Builder
	b.WriteString(RenderHeader("Select Language/Runtime", "Choose your preferred programming language"))

	languages := scaffolding.GetSupportedLanguages()
	for i, lang := range languages {
		if i == m.cursor {
			b.WriteString(SelectedItemStyle + lang.Name + "\n")
		} else {
			b.WriteString(UnselectedItemStyle.Render(lang.Name) + "\n")
		}
	}

	return BaseStyle.Render(b.String())
}

// viewFrameworkSelection renders Go framework selection screen
func (m Model) viewFrameworkSelection() string {
	var b strings.Builder
	b.WriteString(RenderHeader("Select Go Framework", "Choose your preferred HTTP framework"))

	frameworks := []struct {
		name        string
		description string
	}{
		{"chi", "Lightweight, composable router"},
		{"echo", "High performance, minimalist framework"},
		{"fiber", "Express-inspired, fastest Go web framework"},
	}

	for i, fw := range frameworks {
		if i == m.cursor {
			b.WriteString(SelectedItemStyle + fw.name + "\n")
			b.WriteString(UnselectedItemStyle.Render("  "+fw.description) + "\n")
		} else {
			b.WriteString(UnselectedItemStyle.Render(fw.name) + "\n")
		}
	}

	return BaseStyle.Render(b.String())
}

// viewCORSConfig renders CORS configuration screen
func (m Model) viewCORSConfig() string {
	var b strings.Builder
	b.WriteString(RenderHeader("Configure CORS Origins", "Enter frontend URLs (comma-separated)"))

	b.WriteString("Frontend URLs:\n")
	b.WriteString(m.textInput.View() + "\n\n")

	if m.validationMsg != "" {
		b.WriteString(RenderError(m.validationMsg) + "\n")
	} else {
		b.WriteString(HelpStyle.Render("Example: localhost:3000, localhost:5173") + "\n")
	}

	return BaseStyle.Render(b.String())
}

// viewArchitectureSelection renders backend architecture selection screen
func (m Model) viewArchitectureSelection() string {
	var b strings.Builder
	b.WriteString(RenderHeader("Backend Architecture", "Choose how your backend is structured"))

	architectures := []struct {
		name        string
		description string
	}{
		{"Microservices", "Different services on different ports/URLs"},
		{"Monolithic", "Single backend URL for all services"},
		{"Hybrid", "Some services share ports with different paths"},
	}

	for i, arch := range architectures {
		if i == m.cursor {
			b.WriteString(SelectedItemStyle + arch.name + "\n")
			b.WriteString(UnselectedItemStyle.Render("  "+arch.description) + "\n")
		} else {
			b.WriteString(UnselectedItemStyle.Render(arch.name) + "\n")
		}
	}

	return BaseStyle.Render(b.String())
}

// viewServicesConfig renders services configuration screen
func (m Model) viewServicesConfig() string {
	var b strings.Builder
	b.WriteString(RenderHeader("Backend Services", ""))

	archType := map[string]string{
		"1": "Microservices",
		"2": "Monolithic",
		"3": "Hybrid",
	}[m.architecture]

	b.WriteString(fmt.Sprintf("Using default services for %s architecture\n\n", archType))
	b.WriteString(HelpStyle.Render("Press Enter to continue with defaults\n"))
	b.WriteString(HelpStyle.Render("(You can customize services later with 'bffgen add-route')"))

	return BaseStyle.Render(b.String())
}

// viewRouteConfig renders route configuration option screen
func (m Model) viewRouteConfig() string {
	var b strings.Builder
	b.WriteString(RenderHeader("Configure Routes", "How would you like to set up routes?"))

	options := []struct {
		name        string
		description string
	}{
		{"Define manually", "Add routes interactively"},
		{"Use a template", "Choose from auth, ecommerce, or content templates"},
		{"Skip for now", "Set up routes later"},
	}

	for i, opt := range options {
		if i == m.cursor {
			b.WriteString(SelectedItemStyle + opt.name + "\n")
			b.WriteString(UnselectedItemStyle.Render("  "+opt.description) + "\n")
		} else {
			b.WriteString(UnselectedItemStyle.Render(opt.name) + "\n")
		}
	}

	return BaseStyle.Render(b.String())
}

// viewConfirmation renders final confirmation screen
func (m Model) viewConfirmation() string {
	var b strings.Builder
	b.WriteString(RenderHeader("Confirm Configuration", "Review your choices"))

	b.WriteString(fmt.Sprintf("Project: %s\n", SuccessStyle.Render(m.projectName)))
	b.WriteString(fmt.Sprintf("Language: %s (%s)\n", m.langType, m.framework))
	b.WriteString(fmt.Sprintf("CORS Origins: %s\n", strings.Join(m.corsOrigins, ", ")))

	archName := map[string]string{
		"1": "Microservices",
		"2": "Monolithic",
		"3": "Hybrid",
	}[m.architecture]
	b.WriteString(fmt.Sprintf("Architecture: %s\n", archName))
	b.WriteString(fmt.Sprintf("Services: %d configured\n\n", len(m.services)))

	b.WriteString(SuccessStyle.Render("Create project? (Y/n)"))

	return BaseStyle.Render(b.String())
}

// viewPreviewPanel renders live configuration preview
func (m Model) viewPreviewPanel() string {
	var b strings.Builder
	b.WriteString(ListHeaderStyle.Render(" Configuration Preview "))
	b.WriteString("\n\n")

	if m.projectName != "" {
		b.WriteString(fmt.Sprintf("📁 Project: %s\n", m.projectName))
	}

	if m.langType != "" {
		b.WriteString(fmt.Sprintf("💻 Language: %s\n", m.langType))
	}

	if m.framework != "" {
		b.WriteString(fmt.Sprintf("🔧 Framework: %s\n", m.framework))
	}

	if len(m.corsOrigins) > 0 {
		b.WriteString(fmt.Sprintf("🌐 CORS: %d origins\n", len(m.corsOrigins)))
	}

	if m.architecture != "" {
		archName := map[string]string{
			"1": "Microservices",
			"2": "Monolithic",
			"3": "Hybrid",
		}[m.architecture]
		b.WriteString(fmt.Sprintf("🏗️  Architecture: %s\n", archName))
	}

	if len(m.services) > 0 {
		b.WriteString(fmt.Sprintf("⚙️  Services: %d\n", len(m.services)))
	}

	// Progress indicator
	totalSteps := 7
	currentStep := int(m.currentScreen) + 1
	if currentStep > totalSteps {
		currentStep = totalSteps
	}
	b.WriteString("\n")
	b.WriteString(RenderProgressBar(currentStep, totalSteps, 20))

	return PreviewBoxStyle.Render(b.String())
}

// viewHelpText renders context-sensitive help text
func (m Model) viewHelpText() string {
	shortcuts := []string{
		"↑↓/jk: navigate",
		"enter: select",
		"esc: back",
		"ctrl+c/q: quit",
	}
	return RenderHelp(shortcuts...)
}
