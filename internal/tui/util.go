package tui

import tea "github.com/charmbracelet/bubbletea"

func Quit(key string, cleanup func()) tea.Cmd {
	for _, k := range QuitKeys {
		if key == k {
			cleanup()
			return tea.Quit
		}
	}
	return nil
}
