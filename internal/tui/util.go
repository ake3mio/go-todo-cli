package tui

import tea "github.com/charmbracelet/bubbletea"

func Quit(key string, m Model) (tea.Model, tea.Cmd) {
	for _, k := range QuitKeys {
		if key == k {
			return m.Quit(), tea.Quit
		}
	}
	return m, nil
}

// CloseCmd lets external code quit the loader via program.Send(CloseCmd()).
func CloseCmd() tea.Msg { return DoneMsg{} }
