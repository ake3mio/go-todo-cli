package add

import (
	"fmt"
	"time"

	"github.com/ake3mio/go-todo-cli/internal/persistence"
	"github.com/ake3mio/go-todo-cli/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

type model struct {
	form       *huh.Form
	repository *persistence.TodoRepository
	message    string
	taskName   string
	dueDate    string
	quitting   bool
	err        error
}

func (m model) Init() tea.Cmd {
	return m.form.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	fm, cmd := m.form.Update(msg)
	if f, ok := fm.(*huh.Form); ok {
		m.form = f
	}

	if k, ok := msg.(tea.KeyMsg); ok {
		m, c := tui.Quit(k.String(), m)
		if c != nil {
			return m, c
		}
	}

	if _, ok := msg.(tui.DoneMsg); ok {
		return m, tea.Quit
	}

	if err, ok := msg.(error); ok {
		m.err = err
		return m, nil
	}

	if m.form.State == huh.StateCompleted {
		m.taskName = m.form.GetString("taskName")
		m.dueDate = m.form.GetString("dueDate")
		parse, err := time.Parse(time.DateOnly, m.dueDate)
		if err != nil {
			m.err = err
		}
		err = (*m.repository).SaveTask(m.taskName, parse)
		if err != nil {
			m.err = err
		}
		return m.Quit(), tea.Quit
	}

	if m.form.State == huh.StateAborted {
		return m.Quit(), tea.Quit
	}

	return m, cmd
}

func (m model) View() string {
	if m.err != nil {
		component := tui.ErrorComponent{}
		return component.Render(m)
	}

	return m.form.View()
}

func (m model) Quit() tea.Model {
	m.quitting = true
	return m
}

func (m model) Err() error {
	return m.err
}

func createModel(repository *persistence.TodoRepository) model {
	f := createNewTaskForm()
	return model{
		form:       f,
		repository: repository,
		message:    "Add Task",
		taskName:   "",
		dueDate:    "",
	}
}

func createNewTaskForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("taskName").
				Title("/////////////// Task name /////////////////"),
		),
		huh.NewGroup(
			huh.NewInput().
				Key("dueDate").
				Title("////////// Due date (YYYY-MM-DD) //////////").
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
