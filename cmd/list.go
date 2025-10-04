package cmd

import (
	"github.com/ake3mio/go-todo-cli/internal/persistence"
	"github.com/ake3mio/go-todo-cli/internal/tui/list"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all todos",
	Run: func(cmd *cobra.Command, args []string) {
		repository := persistence.NewTodoRepository()
		newList := list.NewList(repository)
		defer newList.Close()
		newList.Wait()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
