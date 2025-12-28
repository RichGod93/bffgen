package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/RichGod93/bffgen/internal/templates"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type templateItem struct {
	template *templates.Template
}

func (i templateItem) FilterValue() string { return i.template.Name }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(templateItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.template.Name)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("▸ " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type TemplateSelectorModel struct {
	list     list.Model
	choice   *templates.Template
	quitting bool
}

func NewTemplateSelector(templateList []*templates.Template) TemplateSelectorModel {
	items := make([]list.Item, len(templateList))
	for i, tmpl := range templateList {
		items[i] = templateItem{template: tmpl}
	}

	const defaultWidth = 80
	const listHeight = 14

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "🎨 Select a Template"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return TemplateSelectorModel{list: l}
}

func (m TemplateSelectorModel) Init() tea.Cmd {
	return nil
}

func (m TemplateSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			i, ok := m.list.SelectedItem().(templateItem)
			if ok {
				m.choice = i.template
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m TemplateSelectorModel) View() string {
	if m.choice != nil {
		return quitTextStyle.Render(fmt.Sprintf("✅ Selected: %s", m.choice.Name))
	}
	if m.quitting {
		return quitTextStyle.Render("Cancelled.")
	}
	return "\n" + m.list.View()
}

// GetSelectedTemplate returns the selected template or nil if cancelled
func (m TemplateSelectorModel) GetSelectedTemplate() *templates.Template {
	return m.choice
}

// RunTemplateSelector runs the interactive template selector
func RunTemplateSelector(templateList []*templates.Template) (*templates.Template, error) {
	if len(templateList) == 0 {
		return nil, fmt.Errorf("no templates available")
	}

	p := tea.NewProgram(NewTemplateSelector(templateList))

	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("error running template selector: %w", err)
	}

	if model, ok := finalModel.(TemplateSelectorModel); ok {
		return model.GetSelectedTemplate(), nil
	}

	return nil, fmt.Errorf("unexpected model type")
}
