package list

import (
	"fmt"
	"strconv"
	"time"

	"github.com/ake3mio/go-todo-cli/internal/data"
	"github.com/ake3mio/go-todo-cli/internal/persistence"
	"github.com/ake3mio/go-todo-cli/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

type model struct {
	form                  *huh.Form
	tasks                 []data.Task
	selectedIDs           []string
	lastSelected          map[int]bool
	hideCompleted         bool
	suppressNextReconcile bool
	repository            *persistence.TodoRepository
	err                   error
	quitting              bool
	ms                    *huh.MultiSelect[string]
}

func (m *model) Init() tea.Cmd {
	return m.form.Init()
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	fm, cmd := m.form.Update(msg)
	if f, ok := fm.(*huh.Form); ok {
		m.form = f
	}

	if k, ok := msg.(tea.KeyMsg); ok {
		if k.String() == "ctrl+h" {
			m.hideCompleted = !m.hideCompleted
			m.form = createNewTaskListForm(m)
			m.suppressNextReconcile = true
			return m, tea.Batch(
				m.form.Init(),
				tea.Tick(0, func(time.Time) tea.Msg { return tea.KeyMsg{Type: tea.KeyHome} }),
			)
		}

		if k.String() == "delete" || k.String() == "backspace" {
			if id, ok := m.ms.Hovered(); ok {
				m.deleteTaskById(id)
				m.form = createNewTaskListForm(m)
				return m, tea.Batch(
					m.form.Init(),
					tea.Tick(0, func(time.Time) tea.Msg { return tea.KeyMsg{Type: tea.KeyHome} }),
				)
			}
			return m, nil
		}
		newM, quitCmd := tui.Quit(k.String(), m)
		if quitCmd != nil {
			_ = m.saveAll()
			return newM, quitCmd
		}
	}

	if !m.suppressNextReconcile {
		if err := m.applyAndSaveToggles(); err != nil {
			m.err = err
		}
	} else {
		m.suppressNextReconcile = false
	}

	if _, ok := msg.(tui.DoneMsg); ok {
		_ = m.saveAll()
		return m, tea.Quit
	}

	if err, ok := msg.(error); ok {
		m.err = err
		return m, nil
	}

	if m.form.State == huh.StateCompleted || m.form.State == huh.StateAborted {
		_ = m.saveAll()
		return m.Quit(), tea.Quit
	}

	return m, cmd
}

func (m *model) View() string {
	if m.err != nil {
		component := tui.ErrorComponent{}
		return component.Render(m)
	}
	return m.form.View()
}

func (m *model) Quit() tea.Model {
	m.quitting = true
	return m
}

func (m *model) Err() error { return m.err }

func createModel(repo *persistence.TodoRepository) *model {
	tasks, err := (*repo).GetTasks()
	if err != nil {
		panic(err)
	}
	m := &model{
		repository:   repo,
		tasks:        tasks,
		selectedIDs:  []string{},
		lastSelected: map[int]bool{},
	}
	m.form = createNewTaskListForm(m)

	return m
}

func createNewTaskListForm(m *model) *huh.Form {
	opts := make([]huh.Option[string], 0, len(m.tasks))
	m.lastSelected = make(map[int]bool)
	m.selectedIDs = m.selectedIDs[:0]
	for _, task := range m.tasks {
		if m.hideCompleted && task.Complete {
			continue
		}
		m.lastSelected[task.Id] = task.Complete
		label := fmt.Sprintf("%s ~ due %s", task.Title, task.DueDate.Format(time.DateOnly))
		idStr := strconv.Itoa(task.Id)
		opt := huh.NewOption(label, idStr)
		opts = append(opts, opt)
		if task.Complete {
			m.selectedIDs = append(m.selectedIDs, idStr)
		}
	}
	ms := huh.NewMultiSelect[string]().
		Title("Tasks").
		Options(opts...).
		Value(&m.selectedIDs).
		Limit(10)
	m.ms = ms
	return huh.NewForm(
		huh.NewGroup(
			ms,
		),
	)
}

func (m *model) applyAndSaveToggles() error {

	curr := make(map[int]bool, len(m.selectedIDs))
	for _, s := range m.selectedIDs {
		if id, err := strconv.Atoi(s); err == nil {
			curr[id] = true
		}
	}

	var firstErr error
	for i := range m.tasks {
		id := m.tasks[i].Id
		shouldBe := curr[id]
		was := m.lastSelected[id]
		if shouldBe != was {

			m.tasks[i].Complete = shouldBe

			if err := (*m.repository).UpdateTask(m.tasks[i]); err != nil && firstErr == nil {
				firstErr = err
			}

			m.lastSelected[id] = shouldBe
		}
	}
	return firstErr
}

func (m *model) saveAll() error {

	_ = m.applyAndSaveToggles()
	return (*m.repository).UpdateTasks(m.tasks)
}

func (m *model) deleteTaskById(id string) {
	if val, err := strconv.Atoi(id); err == nil {
		out := m.tasks[:0]

		for _, task := range m.tasks {
			if task.Id != val {
				out = append(out, task)
			}
		}
		m.tasks = out

		newSelected := m.selectedIDs[:0]
		for _, selected := range m.selectedIDs {
			if selected != id {
				newSelected = append(newSelected, selected)
			}
		}
		m.selectedIDs = newSelected
		delete(m.lastSelected, val)
		err = (*m.repository).DeleteTaskById(val)
		if err != nil {
			m.err = err
		}
	}
}
