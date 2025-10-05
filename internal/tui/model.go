package tui

import tea "github.com/charmbracelet/bubbletea"

var (
	QuitKeys = []string{"q", "esc", "ctrl+c"}
)

type Model interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() string
	Cleanup()
	Err() error
	Next() Command
}
