package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	padding  = 2
	maxWidth = 80
)

var (
	currentPkgNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	doneStyle           = lipgloss.NewStyle().Margin(1, 2)
	checkMark           = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
)

type ProgressMsg struct {
	Step    int
	Total   int
	Message string
}

type DoneMsg struct{}

type ProgressModel struct {
	progress  progress.Model
	current   int
	total     int
	message   string
	done      bool
	steps     []string
	startTime time.Time
}

func NewProgressModel(total int, steps []string) ProgressModel {
	prog := progress.New(progress.WithDefaultGradient())
	return ProgressModel{
		progress:  prog,
		total:     total,
		steps:     steps,
		startTime: time.Now(),
	}
}

func (m ProgressModel) Init() tea.Cmd {
	return nil
}

func (m ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - padding*2 - 4
		if m.progress.Width > maxWidth {
			m.progress.Width = maxWidth
		}
		return m, nil

	case ProgressMsg:
		m.current = msg.Step
		m.message = msg.Message
		if m.current >= m.total {
			m.done = true
			return m, tea.Sequence(
				tea.Printf("%s Done! Completed in %s", checkMark, time.Since(m.startTime).Round(time.Millisecond)),
				tea.Quit,
			)
		}
		return m, nil

	case DoneMsg:
		m.done = true
		return m, tea.Quit

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	default:
		return m, nil
	}
}

func (m ProgressModel) View() string {
	if m.done {
		elapsed := time.Since(m.startTime).Round(time.Millisecond)
		return doneStyle.Render(fmt.Sprintf("%s All done! Completed in %s\n", checkMark, elapsed))
	}

	pad := strings.Repeat(" ", padding)
	percent := float64(m.current) / float64(m.total)

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(pad + currentPkgNameStyle.Render(m.message) + "\n\n")
	b.WriteString(pad + m.progress.ViewAs(percent) + "\n\n")
	b.WriteString(pad + fmt.Sprintf("Step %d/%d", m.current, m.total))

	// Show completed steps
	if len(m.steps) > 0 && m.current > 0 {
		b.WriteString("\n\n" + pad + "Completed:\n")
		for i := 0; i < m.current && i < len(m.steps); i++ {
			b.WriteString(pad + "  " + checkMark.String() + " " + m.steps[i] + "\n")
		}
	}

	return b.String()
}

// RunProgress runs a progress bar with the given steps
func RunProgress(steps []string, workFunc func(int, func(ProgressMsg))) error {
	p := tea.NewProgram(NewProgressModel(len(steps), steps))

	// Run the work in a goroutine
	go func() {
		reporter := func(msg ProgressMsg) {
			p.Send(msg)
		}

		for i, step := range steps {
			reporter(ProgressMsg{
				Step:    i,
				Total:   len(steps),
				Message: step,
			})
			workFunc(i, reporter)
		}

		reporter(ProgressMsg{
			Step:    len(steps),
			Total:   len(steps),
			Message: "Complete!",
		})
		p.Send(DoneMsg{})
	}()

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running progress: %w", err)
	}

	return nil
}
