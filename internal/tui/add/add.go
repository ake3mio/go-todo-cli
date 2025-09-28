package add

import (
	"context"

	"github.com/ake3mio/go-todo-cli/internal/persistence"
	"github.com/ake3mio/go-todo-cli/internal/tui"
)

func NewAdd(repository *persistence.TodoRepository) *tui.Runner {
	return tui.NewRunner(context.Background(), createModel(repository))
}
