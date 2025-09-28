package loader

import (
	"errors"
	"testing"

	"github.com/ake3mio/go-todo-cli/internal/tui"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestModel_InitialState(t *testing.T) {
	m := createModel()
	assert.Equal(t, m.spinner.Spinner, spinner.Globe)
	assert.Equal(t, m.message, "Loading tasks...")
	assert.Equal(t, tui.QuitKeys, []string{"q", "esc", "ctrl+c"})
}

func TestModel_Update_KeyExit(t *testing.T) {
	m := createModel()

	for _, key := range tui.QuitKeys {
		update, cmd := sendKeyMsg(key, m)

		assert.NotEqual(t, m, update)
		assert.Equal(t, tea.Quit(), cmd())
		assert.True(t, update.(model).quitting)

		m = createModel()
	}

	update, cmd := sendKeyMsg("a", m)
	assert.Equal(t, m, update)
	assert.Nil(t, cmd)
	assert.False(t, update.(model).quitting)
}

func TestModel_Update_DoneMsg(t *testing.T) {
	m := createModel()

	update, cmd := m.Update(tui.DoneMsg{})

	assert.Equal(t, m, update)
	assert.Equal(t, tea.Quit(), cmd())
}

func TestModel_Update_Error(t *testing.T) {
	m := createModel()

	err := errors.New("error")
	update, cmd := m.Update(err)

	assert.NotEqual(t, m, update)
	assert.Equal(t, err, update.(model).err)
	assert.Nil(t, cmd)
}

func TestModel_Update_Default(t *testing.T) {
	m := createModel()

	update, cmd := m.Update(struct{}{})

	assert.Equal(t, m, update)
	assert.Nil(t, cmd)
}

func sendKeyMsg(key string, m model) (tea.Model, tea.Cmd) {
	keyMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune(key),
	}
	update, cmd := m.Update(keyMsg)
	return update, cmd
}
