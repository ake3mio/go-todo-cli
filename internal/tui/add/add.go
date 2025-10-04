package add

import (
	"context"

	"github.com/ake3mio/go-todo-cli/internal/persistence"
	"github.com/ake3mio/go-todo-cli/internal/tui"
)

func NewAdd(repository persistence.TodoRepository) *tui.Runner {
	var model tui.Model = createModel(repository)
	return tui.NewRunner(context.Background(), model)
}
