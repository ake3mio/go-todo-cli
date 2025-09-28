package loader

import (
	"fmt"
	"strings"

	"github.com/ake3mio/go-todo-cli/internal/tui"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	spinner spinner.Model
	message string

	quitting bool
	err      error

	width  int
	height int
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		return tui.Quit(msg.String(), m)
	case tui.DoneMsg:
		return m, tea.Quit
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case error:
		m.err = msg
		return m, nil

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m model) View() string {
	if m.err != nil {
		component := tui.ErrorComponent{}
		return component.Render(m)
	}

	spn := m.spinner.View()
	body := fmt.Sprintf("%s %s", spn, m.message)

	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Render(fmt.Sprintf("Press %s to quit.",
			lipgloss.NewStyle().Bold(true).Render(strings.Join(tui.QuitKeys, "/"))))

	content := body + "\n\n" + hint
	if m.quitting {
		content += "\n"
	}

	// Center the content in the viewport if we know the size.
	if m.width > 0 && m.height > 0 {
		box := lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(m.width).
			Height(m.height).
			Render(content)
		return box
	}
	return content + "\n"
}

func (m model) Quit() tea.Model {
	m.quitting = true
	return m
}

func (m model) Err() error {
	return m.err
}

func createModel() model {
	s := spinner.New()
	s.Spinner = spinner.Globe
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		spinner: s,
		message: "Loading tasks...",
	}
}
