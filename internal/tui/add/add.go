package add

import (
	"context"

	"github.com/ake3mio/go-todo-cli/internal/tui"
)

func NewAdd() *tui.Runner {
	return tui.NewRunner(context.Background(), createModel())
}
