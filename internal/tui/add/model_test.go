package add

import (
	"errors"
	"testing"
	"time"

	"github.com/ake3mio/go-todo-cli/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/stretchr/testify/assert"
)

func TestModel_InitialState(t *testing.T) {
	m := createModel()
	assert.Equal(t, "Add Task", m.message)
	assert.Equal(t, "", m.taskName)
	assert.Equal(t, "", m.dueDate)
}

func TestModel_Update_KeyExit(t *testing.T) {
	m := createModel()

	for _, key := range tui.QuitKeys {
		msg, cmd := sendKeyMsg(key, m)

		assert.NotEqual(t, m, msg)
		assert.Equal(t, tea.Quit(), cmd())
		assert.True(t, msg.(model).quitting)
		m = createModel()
	}

	update, cmd := sendKeyMsg("a", m)
	assert.Equal(t, m, update)
	assert.Nil(t, cmd)
	assert.False(t, update.(model).quitting)
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

func TestModel_Init_ReturnsCmd(t *testing.T) {
	m := createModel()
	cmd := m.Init()
	assert.NotNil(t, cmd, "Init should return a non-nil tea.Cmd from the form")
}

func TestModel_Update_FormCompleted_Quits(t *testing.T) {
	m := createModel()

	m.form.State = huh.StateCompleted

	next, cmd := m.Update(struct{}{})
	assert.Equal(t, tea.Quit(), cmd())
	got := next.(model)
	assert.True(t, got.quitting, "model should be in quitting state")

	assert.Equal(t, "", got.taskName)
	assert.Equal(t, "", got.dueDate)
}

func TestModel_Update_FormAborted_Quits(t *testing.T) {
	m := createModel()
	m.form.State = huh.StateAborted

	next, cmd := m.Update(struct{}{})
	assert.Equal(t, tea.Quit(), cmd())
	assert.True(t, next.(model).quitting)
}

func TestModel_Quit_ValueSemantics(t *testing.T) {
	m := createModel()
	assert.False(t, m.quitting)

	m2 := m.Quit().(model)
	assert.False(t, m.quitting, "original model should remain unchanged (value receiver)")
	assert.True(t, m2.quitting, "returned model should have quitting=true")
}

func TestModel_View_RendersFormTitles(t *testing.T) {
	var m model = createModel()
	if cmd := m.Init(); cmd != nil {
		if msg := cmd(); msg != nil {
			nm, _ := m.Update(msg)
			m = nm.(model)
		}
	}
	out := m.View()
	assert.Contains(t, out, "Task name")
	for _, r := range "Write tests" {
		update, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = update.(model)
	}

	update, _ := m.Update(m.form.NextGroup())
	m = update.(model)

	out = m.View()
	assert.Contains(t, out, "Due date (YYYY-MM-DD)")
}

func TestIsDateBeforeToday(t *testing.T) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := todayStart.AddDate(0, 0, -1)
	todayNoon := todayStart.Add(12 * time.Hour)
	tomorrow := todayStart.AddDate(0, 0, 1)

	assert.True(t, isDateBeforeToday(yesterday))

	assert.False(t, isDateBeforeToday(todayNoon))

	assert.False(t, isDateBeforeToday(todayStart))

	assert.False(t, isDateBeforeToday(tomorrow))
}

func TestModel_Update_ErrorSetsErr(t *testing.T) {
	m := createModel()
	want := errors.New("boom")
	next, cmd := m.Update(want)
	got := next.(model)

	assert.Nil(t, cmd)
	assert.EqualError(t, got.err, "boom")
}

func sendKeyMsg(key string, m model) (tea.Model, tea.Cmd) {
	keyMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune(key),
	}
	update, cmd := m.Update(keyMsg)
	return update, cmd
}
