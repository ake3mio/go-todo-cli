package loader

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	spinner  spinner.Model
	message  string
	quitKeys []string

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
		key := msg.String()
		for _, k := range m.quitKeys {
			if key == k {
				m.quitting = true
				return m, tea.Quit
			}
		}
		return m, nil
	case doneMsg:
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
		return lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(fmt.Sprintf("Error: %v\n", m.err))
	}

	spn := m.spinner.View()
	body := fmt.Sprintf("%s %s", spn, m.message)

	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Render(fmt.Sprintf("Press %s to quit.",
			lipgloss.NewStyle().Bold(true).Render(strings.Join(m.quitKeys, "/"))))

	content := body + "\n\n" + hint
	if m.quitting {
		content += "\n"
	}

	// Center the content in the viewport if we know size.
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

func createModel() model {
	s := spinner.New()
	s.Spinner = spinner.Globe
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		spinner:  s,
		message:  "Loading tasks...",
		quitKeys: []string{"q", "esc", "ctrl+c"},
	}
}
