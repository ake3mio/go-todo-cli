package loader

import (
	"context"

	"github.com/ake3mio/go-todo-cli/internal/tui"
)

func NewLoader(ctx context.Context) *tui.Runner {
	return tui.NewRunner(ctx, createModel())
}
