package tui

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
)

type testModel struct{}

func (testModel) Init() tea.Cmd                           { return nil }
func (testModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return testModel{}, nil }
func (testModel) View() string                            { return "" }

type quitOnInit struct{}

func (quitOnInit) Init() tea.Cmd                           { return tea.Quit }
func (quitOnInit) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return quitOnInit{}, nil }
func (quitOnInit) View() string                            { return "" }

func newHeadlessRunner(m tea.Model) *Runner {
	p := tea.NewProgram(
		m,
		tea.WithoutRenderer(),
		tea.WithInput(bytes.NewBuffer(nil)),
		tea.WithOutput(io.Discard),
	)
	r := &Runner{
		program: p,
		doneCh:  make(chan struct{}),
	}
	go func() {
		_, err := r.program.Run()
		r.Err = err
		close(r.doneCh)
	}()
	return r
}

func waitOrFail(t *testing.T, r *Runner, d time.Duration) error {
	t.Helper()
	select {
	case <-time.After(d):
		t.Fatalf("timeout waiting for runner to finish")
		return nil
	case <-r.Done():
		return r.Wait()
	}
}

func TestWait_PropagatesRunErr(t *testing.T) {
	r := &Runner{doneCh: make(chan struct{})}
	want := errors.New("boom")
	r.Err = want
	close(r.doneCh)

	got := r.Wait()
	require.Equal(t, want, got)
}

func TestHeadless_ContextCancel_StopsProgram(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := newHeadlessRunner(testModel{})

	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			r.Stop()
			_ = r.Wait()
		case <-r.Done():
		}
		close(done)
	}()

	cancel()

	select {
	case <-done:
		err := waitOrFail(t, r, 2*time.Second)
		require.NoError(t, err, "unexpected error from Wait after Stop")
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for cancel goroutine")
	}
}

func TestStop_IsIdempotent(t *testing.T) {
	r := newHeadlessRunner(testModel{})

	r.Stop()
	r.Stop()

	err := waitOrFail(t, r, 2*time.Second)
	require.NoError(t, err)
}

func TestProgram_ExitsCleanly_PropagatesNilErr(t *testing.T) {
	r := newHeadlessRunner(quitOnInit{})

	err := waitOrFail(t, r, 2*time.Second)
	require.NoError(t, err)
}

func TestWait_BlocksUntilDone(t *testing.T) {
	r := newHeadlessRunner(testModel{})

	go func() {
		time.Sleep(50 * time.Millisecond)
		r.Stop()
	}()

	start := time.Now()
	err := waitOrFail(t, r, 2*time.Second)
	elapsed := time.Since(start)

	require.NoError(t, err)
	require.GreaterOrEqual(t, elapsed, 50*time.Millisecond, "Wait should block until Stop/Done")
}
