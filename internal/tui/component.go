package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

type Component interface {
	Render(model Model) string
}

type ErrorComponent struct{}

func (e ErrorComponent) Render(model Model) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("1")).
		Render(fmt.Sprintf("Error: %v\n", model.Err()))
}
