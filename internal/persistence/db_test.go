package persistence

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_SaveTask(t *testing.T) {
	os.Remove("todo.sqlite")

	repository := NewTodoRepository()
	task := "Test"
	dueDate := time.Date(2025, time.September, 28, 0, 0, 0, 0, time.Local)
	err := (*repository).SaveTask(task, dueDate)

	assert.Nil(t, err)

	todoRepository := (*repository).(*SqlLiteTodoRepository)

	rows, err := todoRepository.db.Query("SELECT title, due_date FROM tasks;")

	assert.Nil(t, err)

	for rows.Next() {
		var rowTask string
		var rowDueDate time.Time
		_ = rows.Scan(&rowTask, &rowDueDate)
		assert.Equal(t, task, rowTask)
		assert.Equal(t, dueDate, rowDueDate.Local())
	}
}
