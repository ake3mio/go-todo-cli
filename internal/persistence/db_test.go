package persistence

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/ake3mio/go-todo-cli/internal/data"
	"github.com/stretchr/testify/assert"
)

func mustNewRepo(t *testing.T) *TodoRepository {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("TODO_DB", filepath.Join(tmp, "todo.sqlite"))
	repo := NewTodoRepository()
	assert.NotNil(t, repo)
	return &repo
}

func cleanup(repo *TodoRepository) {
	db := (*repo).(*SqlLiteTodoRepository).db
	db.Exec("DELETE FROM tasks;")
	db.Exec("DELETE FROM sqlite_sequence WHERE name='tasks';")
	db.Close()
}

func Test_NewTodoRepository_And_GetTasks_Empty(t *testing.T) {
	repo := mustNewRepo(t)

	tasks, err := (*repo).GetTasks()
	assert.Nil(t, err)
	assert.Empty(t, tasks)
	t.Cleanup(func() {
		cleanup(repo)
	})
}

func Test_GetTasks_Returns_In_DueDate_Order(t *testing.T) {
	repo := mustNewRepo(t)

	d1 := time.Date(2025, time.September, 30, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2025, time.September, 28, 0, 0, 0, 0, time.UTC)
	d3 := time.Date(2025, time.October, 1, 0, 0, 0, 0, time.UTC)

	assert.Nil(t, (*repo).SaveTask("B", d1))
	assert.Nil(t, (*repo).SaveTask("A", d2))
	assert.Nil(t, (*repo).SaveTask("C", d3))

	got, err := (*repo).GetTasks()
	assert.Nil(t, err)
	if assert.Len(t, got, 3) {
		assert.Equal(t, "A", got[0].Title)
		assert.Equal(t, d2.Local().Truncate(time.Second), got[0].DueDate.Local().Truncate(time.Second))

		assert.Equal(t, "B", got[1].Title)
		assert.Equal(t, d1.Local().Truncate(time.Second), got[1].DueDate.Local().Truncate(time.Second))

		assert.Equal(t, "C", got[2].Title)
		assert.Equal(t, d3.Local().Truncate(time.Second), got[2].DueDate.Local().Truncate(time.Second))
	}

	t.Cleanup(func() {
		cleanup(repo)
	})
}

func Test_UpdateTask_Updates_Title_Complete_DueDate(t *testing.T) {
	repo := mustNewRepo(t)

	d := time.Date(2025, time.September, 28, 0, 0, 0, 0, time.UTC)
	assert.Nil(t, (*repo).SaveTask("Old", d))

	tasks, err := (*repo).GetTasks()
	assert.Nil(t, err)
	if assert.Len(t, tasks, 1) {
		task := tasks[0]
		task.Title = "New"
		task.Complete = true
		task.DueDate = time.Date(2025, time.October, 2, 0, 0, 0, 0, time.UTC)

		assert.Nil(t, (*repo).UpdateTask(task))

		after, err := (*repo).GetTasks()
		assert.Nil(t, err)
		if assert.Len(t, after, 1) {
			got := after[0]
			assert.Equal(t, task.Id, got.Id)
			assert.Equal(t, "New", got.Title)
			assert.True(t, got.Complete)
			assert.Equal(t, task.DueDate.Local().Truncate(time.Second), got.DueDate.Local().Truncate(time.Second))
		}
	}
	t.Cleanup(func() {
		cleanup(repo)
	})
}

func Test_UpdateTasks_Batch(t *testing.T) {
	repo := mustNewRepo(t)

	d1 := time.Date(2025, time.September, 28, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2025, time.September, 29, 0, 0, 0, 0, time.UTC)

	assert.Nil(t, (*repo).SaveTask("T1", d1))
	assert.Nil(t, (*repo).SaveTask("T2", d2))

	loaded, err := (*repo).GetTasks()
	assert.Nil(t, err)
	if assert.Len(t, loaded, 2) {
		loaded[0].Title = "T1-updated"
		loaded[0].Complete = true
		loaded[0].DueDate = d1.Add(24 * time.Hour)

		loaded[1].Title = "T2-updated"
		loaded[1].Complete = true
		loaded[1].DueDate = d2.Add(48 * time.Hour)

		assert.Nil(t, (*repo).UpdateTasks(loaded))

		after, err := (*repo).GetTasks()
		assert.Nil(t, err)
		if assert.Len(t, after, 2) {
			m := map[int]data.Task{after[0].Id: after[0], after[1].Id: after[1]}

			t1 := m[loaded[0].Id]
			assert.Equal(t, "T1-updated", t1.Title)
			assert.True(t, t1.Complete)
			assert.Equal(t, loaded[0].DueDate.Local().Truncate(time.Second), t1.DueDate.Local().Truncate(time.Second))

			t2 := m[loaded[1].Id]
			assert.Equal(t, "T2-updated", t2.Title)
			assert.True(t, t2.Complete)
			assert.Equal(t, loaded[1].DueDate.Local().Truncate(time.Second), t2.DueDate.Local().Truncate(time.Second))
		}
	}
	t.Cleanup(func() {
		cleanup(repo)
	})
}

func Test_DeleteTaskById_Removes_Row(t *testing.T) {
	repo := mustNewRepo(t)

	d := time.Date(2025, time.September, 28, 0, 0, 0, 0, time.UTC)
	assert.Nil(t, (*repo).SaveTask("ToDelete", d))
	assert.Nil(t, (*repo).SaveTask("ToKeep", d))

	all, err := (*repo).GetTasks()
	assert.Nil(t, err)
	if assert.Len(t, all, 2) {
		idToDelete := all[0].Id
		assert.Nil(t, (*repo).DeleteTaskById(idToDelete))

		after, err := (*repo).GetTasks()
		assert.Nil(t, err)
		if assert.Len(t, after, 1) {
			assert.NotEqual(t, idToDelete, after[0].Id)
			assert.Equal(t, "ToKeep", after[0].Title)
		}
	}

	t.Cleanup(func() {
		cleanup(repo)
	})
}
