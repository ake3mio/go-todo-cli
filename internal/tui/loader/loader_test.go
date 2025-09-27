package loader

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type testModel struct{}

func (testModel) Init() tea.Cmd { return nil }
func (testModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case doneMsg:
		return testModel{}, tea.Quit
	default:
		return testModel{}, nil
	}
}
func (testModel) View() string { return "" }

func newTestLoader(m tea.Model) *Loader {
	p := tea.NewProgram(
		m,
		tea.WithInput(bytes.NewBuffer(nil)),
		tea.WithOutput(io.Discard),
	)
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

func waitOrFail(t *testing.T, l *Loader, d time.Duration) error {
	t.Helper()
	done := make(chan error, 1)
	go func() {
		done <- l.Wait()
	}()
	select {
	case err := <-done:
		return err
	case <-time.After(d):
		t.Fatalf("timeout waiting for loader to finish")
		return nil
	}
}

func TestClose_Idempotent(t *testing.T) {
	ldr := newTestLoader(testModel{})

	ldr.Close()

	if err := waitOrFail(t, ldr, 2*time.Second); err != nil {
		t.Fatalf("unexpected error from Wait: %v", err)
	}
}

func TestWait_PropagatesRunErr(t *testing.T) {
	ldr := &Loader{
		doneCh: make(chan struct{}),
	}
	want := errors.New("boom")
	ldr.runErr = want
	close(ldr.doneCh)

	if got := ldr.Wait(); got != want {
		t.Fatalf("Wait() = %v, want %v", got, want)
	}
}

func TestHeadless_Close_ShutsDown(t *testing.T) {
	ldr := newTestLoader(testModel{})
	ldr.Close()

	if err := waitOrFail(t, ldr, 2*time.Second); err != nil {
		t.Fatalf("unexpected error from Wait: %v", err)
	}
}

func TestHeadless_NewLoaderLike_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ldr := newTestLoader(testModel{})

	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			ldr.Close()
			_ = ldr.Wait()
		case <-ldr.doneCh:
		}
		close(done)
	}()

	cancel()

	select {
	case <-done:
		if err := waitOrFail(t, ldr, 2*time.Second); err != nil {
			t.Fatalf("unexpected error from Wait: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for NewLoader-like goroutine")
	}
}
