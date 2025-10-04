package list

import (
	"errors"
	"testing"
	"time"

	"github.com/ake3mio/go-todo-cli/internal/data"
	"github.com/ake3mio/go-todo-cli/internal/persistence"
	"github.com/ake3mio/go-todo-cli/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

type fakeRepo struct {
	tasks            []data.Task
	updateTaskCalls  []data.Task
	updateTasksCalls int
	deletes          []int
}

func (r *fakeRepo) Close() error                                  { return nil }
func (r *fakeRepo) SaveTask(task string, dueDate time.Time) error { return nil }

func (r *fakeRepo) GetTasks() ([]data.Task, error) {
	cp := make([]data.Task, len(r.tasks))
	copy(cp, r.tasks)
	return cp, nil
}

func (r *fakeRepo) UpdateTask(t data.Task) error {
	for i := range r.tasks {
		if r.tasks[i].Id == t.Id {
			r.tasks[i] = t
			break
		}
	}
	r.updateTaskCalls = append(r.updateTaskCalls, t)
	return nil
}

func (r *fakeRepo) UpdateTasks(ts []data.Task) error {
	cp := make([]data.Task, len(ts))
	copy(cp, ts)
	r.tasks = cp
	r.updateTasksCalls++
	return nil
}

func (r *fakeRepo) DeleteTaskById(id int) error {
	newSlice := r.tasks[:0]
	for _, t := range r.tasks {
		if t.Id != id {
			newSlice = append(newSlice, t)
		}
	}
	r.tasks = newSlice
	r.deletes = append(r.deletes, id)
	return nil
}

func newFakeRepo() (persistence.TodoRepository, *fakeRepo) {
	repo := &fakeRepo{
		tasks: []data.Task{
			{Id: 1, Title: "A", Complete: false, DueDate: time.Now()},
			{Id: 2, Title: "B", Complete: true, DueDate: time.Now()},
		},
	}
	var todoRepo persistence.TodoRepository = repo
	return todoRepo, repo
}

func sendKey(m tea.Model, key string) (tea.Model, tea.Cmd) {
	switch key {
	case "ctrl+h":
		return m.Update(tea.KeyMsg{Type: tea.KeyCtrlH})
	case "delete":
		return m.Update(tea.KeyMsg{Type: tea.KeyDelete})
	case "backspace":
		return m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	default:
		return m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	}
}

func drain(cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	_ = cmd()
}

func TestModel_InitialState(t *testing.T) {
	tr, _ := newFakeRepo()
	m := createModel(tr)

	assert.NotNil(t, m.form)
	assert.NotNil(t, m.ms)

	assert.Equal(t, map[int]bool{
		1: false,
		2: true,
	}, m.lastSelected)

	assert.ElementsMatch(t, []string{"2"}, m.selectedIDs)
}

func TestModel_ToggleHideCompletedWithCtrlH(t *testing.T) {
	tr, _ := newFakeRepo()
	m := createModel(tr)

	upd, cmd := sendKey(m, "ctrl+h")
	drain(cmd)
	got := upd.(*model)

	assert.True(t, got.hideCompleted, "hideCompleted should be true after toggle")

	assert.NotContains(t, got.selectedIDs, "2")
}

func TestModel_DeleteHovered_RemovesFirstItem(t *testing.T) {
	tr, fr := newFakeRepo()
	m := createModel(tr)

	upd, cmd := sendKey(m, "delete")
	drain(cmd)
	got := upd.(*model)

	assert.Contains(t, fr.deletes, 1)

	for _, tsk := range got.tasks {
		assert.NotEqual(t, 1, tsk.Id)
	}
}

func TestModel_Reconcile_ToggleSelection_PersistsImmediately(t *testing.T) {
	tr, fr := newFakeRepo()
	m := createModel(tr)

	m.selectedIDs = append(m.selectedIDs, "1")
	m.lastSelected[1] = false

	upd, cmd := m.Update(struct{}{})
	assert.NotNil(t, upd)
	assert.Nil(t, cmd)

	assert.Len(t, fr.updateTaskCalls, 1)
	assert.Equal(t, 1, fr.updateTaskCalls[0].Id)
	assert.True(t, fr.updateTaskCalls[0].Complete)

	assert.True(t, m.lastSelected[1])
}

func TestModel_ErrorMsg_BubblesIntoErr(t *testing.T) {
	tr, _ := newFakeRepo()
	m := createModel(tr)

	e := errors.New("boom")
	upd, cmd := m.Update(e)
	got := upd.(*model)

	assert.Nil(t, cmd)
	assert.Equal(t, e, got.err)
}

func TestModel_QuitKeys_Quit(t *testing.T) {
	tr, _ := newFakeRepo()
	m := createModel(tr)

	for _, key := range tui.QuitKeys {
		_, cmd := sendKey(m, key)
		assert.NotNil(t, cmd)

		assert.Equal(t, tea.Quit(), cmd())

		m = createModel(tr)
	}
}

func TestModel_NoOpMsg_NoChange(t *testing.T) {
	tr, _ := newFakeRepo()
	m := createModel(tr)

	upd, cmd := m.Update(struct{}{})
	assert.Same(t, m, upd)
	assert.Nil(t, cmd)

}
