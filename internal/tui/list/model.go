package list

import (
	"github.com/ake3mio/go-todo-cli/internal"
	"github.com/ake3mio/go-todo-cli/internal/persistence"
)

type model struct {
	tasks      []internal.Task
	repository *persistence.TodoRepository
	err        error
}
