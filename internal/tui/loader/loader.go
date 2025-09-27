package loader

import (
	"context"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

type Loader struct {
	program *tea.Program
	once    sync.Once
	doneCh  chan struct{}
	runErr  error
}

type doneMsg struct{}

// CloseCmd lets external code quit the loader via Program.Send(CloseCmd()).
func CloseCmd() tea.Msg { return doneMsg{} }

// Close requests the loader to quit (safe to call multiple times).
func (l *Loader) Close() {
	l.once.Do(func() {
		l.program.Send(CloseCmd())
	})
}

// Wait blocks until the loader exits and returns any run error.
func (l *Loader) Wait() error {
	<-l.doneCh
	return l.runErr
}

func RunLoader() *Loader {
	m := createModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	ldr := &Loader{
		program: p,
		doneCh:  make(chan struct{}),
	}

	go func() {
		_, err := ldr.program.Run()
		ldr.runErr = err
		close(ldr.doneCh)
	}()
	return ldr
}

func NewLoader(ctx context.Context) *Loader {
	ldr := RunLoader()

	go func() {
		select {
		case <-ctx.Done():
			ldr.Close()
			err := ldr.Wait()
			if err != nil {
				panic(err)
			}

		case <-ldr.doneCh:
			if ldr.runErr != nil {
				panic(ldr.runErr)
			}
		}
	}()

	return ldr
}
