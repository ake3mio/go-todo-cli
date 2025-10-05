package cmd

import (
	"github.com/ake3mio/go-todo-cli/internal/persistence"
	"github.com/ake3mio/go-todo-cli/internal/tui"
	"github.com/ake3mio/go-todo-cli/internal/tui/add"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:        string(tui.AddTask),
	Aliases:    nil,
	SuggestFor: nil,
	Short:      "Add a task to do",
	Long: `
Type the task name and due date, then press Enter to save it.
Once the task is added, you’ll automatically return to the task list view.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repository := persistence.NewTodoRepository()
		runner := add.NewAdd(repository)
		return runner.Run(rootCmd)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
