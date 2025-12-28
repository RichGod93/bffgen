package tui

import (
	"fmt"

	"github.com/RichGod93/bffgen/internal/scaffolding"
	"github.com/RichGod93/bffgen/internal/types"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Screen represents different TUI screens
type Screen int

const (
	ScreenLanguage Screen = iota
	ScreenFramework
	ScreenCORS
	ScreenArchitecture
	ScreenServices
	ScreenRouteConfig
	ScreenConfirm
	ScreenDone
)

// Model represents the TUI state
type Model struct {
	// Navigation
	currentScreen Screen
	cursor        int
	err           error

	// Project configuration
	projectName  string
	langType     scaffolding.LanguageType
	framework    string
	corsOrigins  []string
	architecture string // "1", "2", or "3"
	services     []types.BackendService
	routeOption  string

	// UI State
	textInput     textinput.Model
	validationMsg string
	width         int
	height        int

	// Results
	completed bool
	cancelled bool
}

// NewModel creates a new TUI model
func NewModel(projectName string) Model {
	ti := textinput.New()
	ti.Placeholder = "localhost:3000"
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 40

	return Model{
		currentScreen: ScreenLanguage,
		cursor:        0,
		projectName:   projectName,
		textInput:     ti,
		corsOrigins:   []string{},
		services:      []types.BackendService{},
		width:         80,
		height:        24,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages and updates model state
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case error:
		m.err = msg
		return m, nil
	}

	// Handle text input updates
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// handleKeyPress processes keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global shortcuts
	switch msg.String() {
	case "ctrl+c", "q":
		m.cancelled = true
		return m, tea.Quit

	case "esc":
		if m.currentScreen > ScreenLanguage {
			m.currentScreen--
			m.resetCursor()
			return m, nil
		}
	}

	// Screen-specific navigation
	switch m.currentScreen {
	case ScreenLanguage:
		return m.handleLanguageScreen(msg)
	case ScreenFramework:
		return m.handleFrameworkScreen(msg)
	case ScreenCORS:
		return m.handleCORSScreen(msg)
	case ScreenArchitecture:
		return m.handleArchitectureScreen(msg)
	case ScreenServices:
		return m.handleServicesScreen(msg)
	case ScreenRouteConfig:
		return m.handleRouteConfigScreen(msg)
	case ScreenConfirm:
		return m.handleConfirmScreen(msg)
	case ScreenDone:
		return m, tea.Quit
	}

	return m, nil
}

// handleLanguageScreen handles language selection
func (m Model) handleLanguageScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	languages := scaffolding.GetSupportedLanguages()

	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(languages)-1 {
			m.cursor++
		}
	case "enter":
		selectedLang := languages[m.cursor]
		m.langType = selectedLang.Type
		m.framework = selectedLang.Framework

		// If Go, need framework selection
		if m.langType == scaffolding.LanguageGo {
			m.currentScreen = ScreenFramework
			m.resetCursor()
		} else {
			m.currentScreen = ScreenCORS
			m.resetCursor()
		}
		return m, nil
	}

	return m, nil
}

// handleFrameworkScreen handles Go framework selection
func (m Model) handleFrameworkScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	frameworks := []string{"chi", "echo", "fiber"}

	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(frameworks)-1 {
			m.cursor++
		}
	case "enter":
		m.framework = frameworks[m.cursor]
		m.currentScreen = ScreenCORS
		m.resetCursor()
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, textinput.Blink
	}

	return m, nil
}

// handleCORSScreen handles CORS configuration
func (m Model) handleCORSScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		value := m.textInput.Value()
		if value == "" {
			value = "localhost:3000"
		}

		// Parse comma-separated origins
		origins := parseOrigins(value)
		m.corsOrigins = origins
		m.currentScreen = ScreenArchitecture
		m.resetCursor()
		return m, nil
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// handleArchitectureScreen handles backend architecture selection
func (m Model) handleArchitectureScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	architectures := []string{"Microservices", "Monolithic", "Hybrid"}

	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(architectures)-1 {
			m.cursor++
		}
	case "enter":
		m.architecture = fmt.Sprintf("%d", m.cursor+1)
		m.currentScreen = ScreenServices
		m.resetCursor()
		return m, nil
	}

	return m, nil
}

// handleServicesScreen handles service configuration
func (m Model) handleServicesScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Simplified - skip for now and generate defaults
	switch msg.String() {
	case "enter":
		// Generate default services based on architecture
		m.services = m.generateDefaultServices()
		m.currentScreen = ScreenRouteConfig
		m.resetCursor()
		return m, nil
	}

	return m, nil
}

// handleRouteConfigScreen handles route configuration option
func (m Model) handleRouteConfigScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	options := []string{"Define manually", "Use a template", "Skip for now"}

	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(options)-1 {
			m.cursor++
		}
	case "enter":
		m.routeOption = fmt.Sprintf("%d", m.cursor+1)
		m.currentScreen = ScreenConfirm
		m.resetCursor()
		return m, nil
	}

	return m, nil
}

// handleConfirmScreen handles final confirmation
func (m Model) handleConfirmScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		m.completed = true
		m.currentScreen = ScreenDone
		return m, tea.Quit
	case "n", "N":
		m.cancelled = true
		return m, tea.Quit
	}

	return m, nil
}

// Helper functions

func (m *Model) resetCursor() {
	m.cursor = 0
}

// GetResults returns the collected configuration
func (m Model) GetResults() (scaffolding.LanguageType, string, []string, string, []types.BackendService, string, bool) {
	return m.langType, m.framework, m.corsOrigins, m.architecture, m.services, m.routeOption, m.completed && !m.cancelled
}

// generateDefaultServices creates default services based on architecture
func (m Model) generateDefaultServices() []types.BackendService {
	switch m.architecture {
	case "1": // Microservices
		return []types.BackendService{
			{Name: "users", BaseURL: "http://localhost:4001/api", Port: 4001},
			{Name: "products", BaseURL: "http://localhost:4002/api", Port: 4002},
		}
	case "2": // Monolithic
		baseURL := "http://localhost:3000/api"
		return []types.BackendService{
			{Name: "users", BaseURL: baseURL, Port: 3000},
			{Name: "products", BaseURL: baseURL, Port: 3000},
			{Name: "orders", BaseURL: baseURL, Port: 3000},
		}
	case "3": // Hybrid
		return []types.BackendService{
			{Name: "users", BaseURL: "http://localhost:3000/api/users", Port: 3000},
			{Name: "products", BaseURL: "http://localhost:3000/api/products", Port: 3000},
		}
	default:
		return []types.BackendService{}
	}
}
