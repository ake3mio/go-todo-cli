package list

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/ake3mio/go-todo-cli/internal/data"
	"github.com/ake3mio/go-todo-cli/internal/persistence"
	"github.com/ake3mio/go-todo-cli/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	form                  *huh.Form
	tasks                 []data.Task
	selectedIDs           []string
	lastSelected          map[int]bool
	hideCompleted         bool
	suppressNextReconcile bool
	repository            persistence.TodoRepository
	err                   error
	ms                    *huh.MultiSelect[string]
	next                  tui.Command
	once                  sync.Once
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
		switch k.String() {
		case "ctrl+h":
			m.hideCompleted = !m.hideCompleted
			m.suppressNextReconcile = true
			return m, m.updateWithNewForm()

		case "ctrl+a":
			m.next = tui.AddTask
			return m, m.cleanupAndQuit()

		case "delete", "backspace":
			if id, ok := m.ms.Hovered(); ok {

				if err := m.deleteTaskById(id); err != nil {
					m.err = err
					return m, nil
				}

				return m, m.updateWithNewForm()
			}
		}

		quitCmd := tui.Quit(k.String(), m.Cleanup)
		if quitCmd != nil {
			var _ = m.saveAll()
			return m, quitCmd
		}
	}

	if !m.suppressNextReconcile {
		var err error

		if err = m.applyAndSaveToggles(); err != nil {
			m.err = err
		}
	} else {
		m.suppressNextReconcile = false
	}

	if err, ok := msg.(error); ok {
		m.err = err
		return m, nil
	}

	if m.form.State == huh.StateCompleted || m.form.State == huh.StateAborted {
		var _ = m.saveAll()
		return m, m.cleanupAndQuit()
	}

	return m, cmd
}

func (m *model) View() string {
	if m.err != nil {
		component := tui.ErrorComponent{}
		return component.Render(m)
	}

	if len(m.tasks) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")).
			Padding(1).
			Render("Press ctrl+a to add a new task.")
	}

	return m.form.View() + lipgloss.NewStyle().
		Foreground(lipgloss.Color("3")).
		Padding(1).
		Render(`

Special Shortcuts:
ctrl + h - Toggle hiding completed tasks
ctrl + a - Add a new task
q/ctrl + c/esc - Quit
`)
}

func (m *model) Cleanup() {
	m.once.Do(func() {
		if err := m.repository.Close(); err != nil {
			m.err = err
		}
	})
}

func (m *model) cleanupAndQuit() tea.Cmd {
	m.Cleanup()
	return tea.Quit
}

func (m *model) updateWithNewForm() tea.Cmd {
	createNewTaskListForm(m)
	return tea.Sequence(
		m.form.Init(),
		tea.Tick(0, func(time.Time) tea.Msg { return tea.KeyMsg{Type: tea.KeyHome} }),
	)
}
func (m *model) Err() error { return m.err }

func (m *model) Next() tui.Command { return m.next }

func createModel(repo persistence.TodoRepository) *model {
	m := &model{
		repository:   repo,
		selectedIDs:  []string{},
		lastSelected: map[int]bool{},
		next:         tui.NoneTask,
	}
	createNewTaskListForm(m)
	return m
}

func createNewTaskListForm(m *model) {
	tasks, err := m.repository.GetTasks()
	if err != nil {
		panic(err)
	}
	m.tasks = tasks
	opts := make([]huh.Option[string], 0, len(m.tasks))
	m.lastSelected = make(map[int]bool)
	m.selectedIDs = make([]string, 0)

	for _, task := range m.tasks {
		if m.hideCompleted && task.Complete {
			continue
		}
		m.lastSelected[task.Id] = task.Complete
		label := fmt.Sprintf("%s ~ due %s", task.Title, task.DueDate.Format(time.DateOnly))
		idStr := strconv.Itoa(task.Id)
		opts = append(opts, huh.NewOption(idStr+" - "+label, idStr))
		if task.Complete {
			m.selectedIDs = append(m.selectedIDs, idStr)
		}
	}

	ms := huh.NewMultiSelect[string]().
		Title("Tasks").
		Options(opts...).
		Value(&m.selectedIDs).
		Filtering(false).
		Filterable(false)

	m.ms = ms

	form := huh.NewForm(huh.NewGroup(ms))
	m.form = form

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
			if err := m.repository.UpdateTask(m.tasks[i]); err != nil && firstErr == nil {
				firstErr = err
			}
			m.lastSelected[id] = shouldBe
		}
	}
	return firstErr
}

func (m *model) saveAll() error {
	_ = m.applyAndSaveToggles()
	return m.repository.UpdateTasks(m.tasks)
}

func (m *model) deleteTaskById(id string) error {
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

		if err = m.repository.DeleteTaskById(val); err != nil {
			m.err = err
			return err
		}
	}

	return nil
}
