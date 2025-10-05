package add

import (
	"errors"
	"testing"
	"time"

	"github.com/ake3mio/go-todo-cli/internal/data"
	"github.com/ake3mio/go-todo-cli/internal/persistence"
	"github.com/ake3mio/go-todo-cli/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/stretchr/testify/assert"
)

type TestTodoRepository struct {
	Closed int
	Saved  []struct {
		Task string
		Due  time.Time
	}
}

func (t *TestTodoRepository) SaveTask(task string, dueDate time.Time) error {
	t.Saved = append(t.Saved, struct {
		Task string
		Due  time.Time
	}{task, dueDate})
	return nil
}
func (t *TestTodoRepository) GetTasks() ([]data.Task, error)      { return []data.Task{}, nil }
func (t *TestTodoRepository) UpdateTask(task data.Task) error     { return nil }
func (t *TestTodoRepository) UpdateTasks(tasks []data.Task) error { return nil }
func (t *TestTodoRepository) DeleteTaskById(id int) error         { return nil }
func (t *TestTodoRepository) Close() error                        { t.Closed++; return nil }

func TestModel_InitialState(t *testing.T) {
	repo := &TestTodoRepository{}
	var repository persistence.TodoRepository = repo

	m := createModel(repository)
	assert.Equal(t, "Add Task", m.message)
	assert.Equal(t, "", m.taskName)

	today := time.Now().Format(time.DateOnly)
	assert.Equal(t, today, m.dueDate, "dueDate should default to today")
}

func TestModel_Init_ReturnsCmd(t *testing.T) {
	repo := &TestTodoRepository{}
	m := createModel(repo)
	cmd := m.Init()
	assert.NotNil(t, cmd, "Init should return a non-nil tea.Cmd from the form")
}

func TestModel_Update_Default_NoOp(t *testing.T) {
	repo := &TestTodoRepository{}
	m := createModel(repo)

	next, cmd := m.Update(struct{}{})
	assert.Same(t, m, next, "model pointer should be unchanged")
	assert.Nil(t, cmd)
}

func TestModel_Update_Error_SetsErr(t *testing.T) {
	repo := &TestTodoRepository{}
	m := createModel(repo)

	want := errors.New("boom")
	next, cmd := m.Update(want)
	got := next.(*model)

	assert.Nil(t, cmd)
	assert.EqualError(t, got.err, "boom")
}

func TestModel_Update_KeyCtrlL_NavigatesAndCleansUp(t *testing.T) {
	repo := &TestTodoRepository{}
	m := createModel(repo)

	next, cmd := m.Update(key("ctrl+l"))
	got := next.(*model)

	assert.Equal(t, tui.ListTasks, got.next)
	assert.NotNil(t, cmd)
	_ = cmd()
	assert.Equal(t, 1, repo.Closed, "Cleanup should call repository.Close() exactly once")
}

func TestModel_Update_QuitKeys_UseQuitHelper(t *testing.T) {
	repo := &TestTodoRepository{}
	m := createModel(repo)

	for _, k := range tui.QuitKeys {
		next, cmd := m.Update(key(k))
		assert.NotNil(t, cmd, "quit key %q should return a command", k)
		assert.Same(t, m, next)
	}
}

func TestModel_Update_FormCompleted_Saves_ThenQuits(t *testing.T) {
	repo := &TestTodoRepository{}
	m := createModel(repo)

	m.taskName = "Write tests"
	m.dueDate = time.Now().Format(time.DateOnly)

	m.form.State = huh.StateCompleted

	next, cmd := m.Update(struct{}{})
	got := next.(*model)

	assert.Equal(t, tui.ListTasks, got.next)
	assert.NotNil(t, cmd)

	_ = cmd()
	assert.Equal(t, 1, repo.Closed)
	if assert.Len(t, repo.Saved, 1) {
		assert.Equal(t, "Write tests", repo.Saved[0].Task)
	}
}

type FailingRepo struct {
	TestTodoRepository
	called bool
}

func (f *FailingRepo) SaveTask(task string, due time.Time) error {
	if !f.called {
		f.called = true
		return errors.New("save failed")
	}
	return nil
}

func TestModel_Update_FormCompleted_SaveError_StaysOnForm(t *testing.T) {
	override := &FailingRepo{}
	var _ persistence.TodoRepository = override

	m := createModel(override)
	m.taskName = "x"
	m.dueDate = time.Now().Format(time.DateOnly)
	m.form.State = huh.StateCompleted

	next, cmd := m.Update(struct{}{})
	got := next.(*model)

	assert.Nil(t, cmd, "on save error we should not quit")
	assert.EqualError(t, got.err, "save failed")
	assert.Equal(t, tui.NoneTask, got.next)
}

func TestModel_Update_FormAborted_QuitsAndCleansUp(t *testing.T) {
	repo := &TestTodoRepository{}
	m := createModel(repo)
	m.form.State = huh.StateAborted

	next, cmd := m.Update(struct{}{})
	assert.NotNil(t, cmd)
	_ = cmd()
	assert.Equal(t, 1, repo.Closed)
	_ = next
}

func TestModel_View_RendersTitles(t *testing.T) {
	repo := &TestTodoRepository{}
	m := createModel(repo)

	if init := m.Init(); init != nil {
		_ = init()
	}

	out := m.View()
	assert.Contains(t, out, "Task name")
	for _, r := range "Write tests" {
		update, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})

		m = update.(*model)
	}
	update, _ := m.Update(m.form.NextGroup())
	m = update.(*model)
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

func key(s string) tea.KeyMsg {
	switch s {
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case "ctrl+d":
		return tea.KeyMsg{Type: tea.KeyCtrlD}
	case "ctrl+l":
		return tea.KeyMsg{Type: tea.KeyCtrlL}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "q":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}
