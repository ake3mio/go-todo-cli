package tui

type Command string

const (
	NoneTask  Command = "-"
	AddTask   Command = "add"
	ListTasks Command = ""
)
