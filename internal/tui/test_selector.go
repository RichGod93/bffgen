package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TestTypeItem struct {
	title string
	desc  string
	value string
}

func (i TestTypeItem) FilterValue() string { return i.title }
func (i TestTypeItem) Title() string       { return i.title }
func (i TestTypeItem) Description() string { return i.desc }

type testDelegate struct{}

func (d testDelegate) Height() int                             { return 2 }
func (d testDelegate) Spacing() int                            { return 1 }
func (d testDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d testDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(TestTypeItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.title)
	desc := i.desc

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("▸ " + strings.Join(s, " "))
		}
	}

	fmt.Fprintf(w, "%s\n", fn(str))
	fmt.Fprintf(w, "   %s", lipgloss.NewStyle().Faint(true).Render(desc))
}

type TestSelectorModel struct {
	list     list.Model
	choice   string
	quitting bool
}

func NewTestSelector() TestSelectorModel {
	items := []list.Item{
		TestTypeItem{
			title: "All Tests",
			desc:  "Generate integration, unit, E2E, and contract tests",
			value: "all",
		},
		TestTypeItem{
			title: "Integration Tests",
			desc:  "API endpoint integration tests with HTTP clients",
			value: "integration",
		},
		TestTypeItem{
			title: "Unit Tests",
			desc:  "Controller and service layer unit tests with mocks",
			value: "unit",
		},
		TestTypeItem{
			title: "E2E Tests",
			desc:  "End-to-end user flow tests with Playwright/Cypress",
			value: "e2e",
		},
		TestTypeItem{
			title: "Contract Tests",
			desc:  "Provider/consumer contract tests with Pact",
			value: "contract",
		},
	}

	const defaultWidth = 80
	const listHeight = 18

	l := list.New(items, testDelegate{}, defaultWidth, listHeight)
	l.Title = "🧪 Select Test Type"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return TestSelectorModel{list: l}
}

func (m TestSelectorModel) Init() tea.Cmd {
	return nil
}

func (m TestSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(TestTypeItem)
			if ok {
				m.choice = i.value
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m TestSelectorModel) View() string {
	if m.choice != "" {
		return quitTextStyle.Render(fmt.Sprintf("✅ Selected: %s tests", m.choice))
	}
	if m.quitting {
		return quitTextStyle.Render("Cancelled.")
	}
	return "\n" + m.list.View()
}

func (m TestSelectorModel) GetChoice() string {
	return m.choice
}

// RunTestSelector runs the interactive test type selector
func RunTestSelector() (string, error) {
	p := tea.NewProgram(NewTestSelector())

	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("error running test selector: %w", err)
	}

	if model, ok := finalModel.(TestSelectorModel); ok {
		return model.GetChoice(), nil
	}

	return "", fmt.Errorf("unexpected model type")
}
