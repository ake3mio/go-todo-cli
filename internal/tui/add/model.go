package add

import (
	"fmt"
	"sync"
	"time"

	"github.com/ake3mio/go-todo-cli/internal/persistence"
	"github.com/ake3mio/go-todo-cli/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	form       *huh.Form
	repository persistence.TodoRepository
	message    string
	taskName   string
	dueDate    string
	err        error
	next       tui.Command
	once       sync.Once
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
		if k.String() == "ctrl+l" {
			m.next = tui.ListTasks
			return m, m.cleanupAndQuit()
		}
		c := tui.Quit(k.String(), m.Cleanup)
		if c != nil {
			return m, c
		}
	}

	if err, ok := msg.(error); ok {
		m.err = err
		return m, nil
	}

	if m.form.State == huh.StateCompleted {
		parse, err := time.Parse(time.DateOnly, m.dueDate)
		if err != nil {
			m.err = err
			return m, nil
		}
		err = m.repository.SaveTask(m.taskName, parse)
		if err != nil {
			m.err = err
			return m, nil
		}
		m.next = tui.ListTasks
		return m, m.cleanupAndQuit()
	}

	if m.form.State == huh.StateAborted {
		return m, m.cleanupAndQuit()
	}

	return m, cmd
}

func (m *model) View() string {
	if m.err != nil {
		component := tui.ErrorComponent{}
		return component.Render(m)
	}

	return m.form.View() + lipgloss.NewStyle().
		Foreground(lipgloss.Color("3")).
		Padding(1).
		Render(`

Special Shortcuts:
ctrl + l - Go to the List View
q/ctrl + c/esc - Quit
`)
}

func (m *model) Cleanup() {
	m.once.Do(func() {
		err := m.repository.Close()
		if err != nil {
			m.err = err
		}

	})
}

func (m *model) cleanupAndQuit() tea.Cmd {
	m.Cleanup()
	return tea.Quit
}

func (m *model) Err() error {
	return m.err
}

func (m *model) Next() tui.Command {
	return m.next
}

func createModel(repository persistence.TodoRepository) *model {
	now := time.Now()
	m := &model{
		repository: repository,
		message:    "Add Task",
		taskName:   "",
		dueDate:    now.Format(time.DateOnly),
		next:       tui.NoneTask,
	}
	f := createNewTaskForm(m)
	m.form = f
	return m
}

func createNewTaskForm(f *model) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("taskName").
				Title("/////////////// Task name /////////////////").
				Value(&f.taskName).
				Validate(func(s string) error {
					if len(s) < 1 {
						return fmt.Errorf("task name cannot be empty")
					}
					return nil
				}),
		),
		huh.NewGroup(
			huh.NewInput().
				Key("dueDate").
				Title("////////// Due date (YYYY-MM-DD) //////////").
				Value(&f.dueDate).
				Validate(func(s string) error {
					inputTime, err := time.Parse(time.DateOnly, s)
					if err != nil {
						return err
					}

					if isDateBeforeToday(inputTime) {
						return fmt.Errorf("%s is in the past", s)
					}
					return nil
				}),
		),
	)
}

func isDateBeforeToday(date time.Time) bool {
	now := time.Now()
	nowAtStartOfDay := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0,
		0,
		0,
		0,
		now.Location(),
	)
	return date.Before(nowAtStartOfDay)
}
