package tui

import (
	"context"
	"fmt"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type Runner struct {
	program *tea.Program
	once    sync.Once
	doneCh  chan struct{}
	Err     error
	model   Model
}

func (r *Runner) Wait() error {
	<-r.doneCh
	return r.Err
}

func (r *Runner) Done() <-chan struct{} { return r.doneCh }

func (r *Runner) Next() Command {
	<-r.doneCh
	return r.model.Next()
}

func (r *Runner) Run(cmd *cobra.Command) error {
	if err := r.Wait(); err != nil {
		return err
	}
	if next := r.model.Next(); next != NoneTask {
		cmd.SetArgs([]string{string(next)})
		if _, err := cmd.ExecuteC(); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) Stop() {
	r.once.Do(func() { r.program.Send(tea.Quit()) })
}

func NewRunner(ctx context.Context, model Model) *Runner {
	r := &Runner{
		model:  model,
		doneCh: make(chan struct{}),
		program: tea.NewProgram(
			model,
			tea.WithReportFocus(),
			tea.WithAltScreen(),
		),
	}

	go func() {
		m, err := r.program.Run()
		if mm, ok := m.(Model); ok {
			r.model = mm
		} else if err == nil {
			err = fmt.Errorf("program returned model not implementing tui.Model")
		}

		if err != nil {
			r.Err = err
		} else if e := r.model.Err(); e != nil {
			r.Err = e
		}
		close(r.doneCh)
	}()

	go func() {
		select {
		case <-ctx.Done():
			r.Stop()
		case <-r.doneCh:
		}
	}()

	return r
}
