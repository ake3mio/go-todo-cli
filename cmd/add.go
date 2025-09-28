package cmd

import (
	"github.com/ake3mio/go-todo-cli/internal/tui/add"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a task to do",
	Run: func(cmd *cobra.Command, args []string) {
		newAdd := add.NewAdd()
		defer newAdd.Close()
		newAdd.Wait()
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
