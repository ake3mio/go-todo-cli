package cmd

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/ake3mio/go-todo-cli/internal/persistence"
	"github.com/ake3mio/go-todo-cli/internal/tui/list"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "todo",
	Short: "Manage your tasks",
	Long: `
View and manage tasks.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repository := persistence.NewTodoRepository()
		runner := list.NewList(repository)
		return runner.Run(cmd)
	},
}

func Execute() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
