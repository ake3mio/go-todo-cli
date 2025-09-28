package tui

import (
	"context"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

type Runner struct {
	program *tea.Program
	once    sync.Once
	doneCh  chan struct{}
	runErr  error
	model   Model
}

func (r *Runner) Close() {
	r.once.Do(func() {
		r.program.Send(CloseCmd())
	})
}

func (r *Runner) Wait() error {
	<-r.doneCh
	return r.runErr
}

func NewRunner(ctx context.Context, model Model) *Runner {
	runner := &Runner{
		model:   model,
		doneCh:  make(chan struct{}),
		program: tea.NewProgram(model, tea.WithReportFocus(), tea.WithAltScreen()),
	}

	go func() {
		_, err := runner.program.Run()
		runner.runErr = err
		close(runner.doneCh)
	}()

	go func() {
		select {
		case <-ctx.Done():
			runner.Close()
		case <-runner.doneCh:
		}
	}()

	return runner
}
