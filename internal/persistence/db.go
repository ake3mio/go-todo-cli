package persistence

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"time"

	"github.com/ake3mio/go-todo-cli/internal/data"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

//go:embed schema.sql
var schema string

func dbFilePath() string {
	if e := os.Getenv("TODO_DB"); e != "" {
		return e
	}
	return "todo.sqlite"
}
func newDB() *sql.DB {
	dsn := fmt.Sprintf("file:%s?mode=rwc&_busy_timeout=5000&_journal_mode=WAL", dbFilePath())
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(1)
	if err = db.Ping(); err != nil {
		panic(err)
	}

	if _, err := db.ExecContext(context.TODO(), schema); err != nil {
		panic(err)
	}
	return db
}

type TodoRepository interface {
	SaveTask(task string, dueDate time.Time) error
	GetTasks() ([]data.Task, error)
	UpdateTask(task data.Task) error
	UpdateTasks(tasks []data.Task) error
	DeleteTaskById(id int) error
	Close() error
}

type SqlLiteTodoRepository struct {
	db *sql.DB
}

func (t *SqlLiteTodoRepository) SaveTask(title string, dueDate time.Time) (err error) {
	ctx := context.TODO()
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.ExecContext(ctx, `INSERT INTO tasks (title, due_date) VALUES (?, ?)`, title, dueDate.UTC().Unix())
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

func (t *SqlLiteTodoRepository) GetTasks() ([]data.Task, error) {
	ctx := context.TODO()
	rows, err := t.db.QueryContext(ctx, `SELECT id, title, complete, due_date FROM tasks ORDER BY due_date`)
	var tasks []data.Task
	if err != nil {
		return tasks, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var title string
		var complete bool
		var dueDate time.Time
		if err := rows.Scan(&id, &title, &complete, &dueDate); err != nil {
			return tasks, err
		}
		task := data.Task{
			Id:       id,
			Title:    title,
			Complete: complete,
			DueDate:  dueDate,
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (t *SqlLiteTodoRepository) UpdateTask(task data.Task) error {
	ctx := context.TODO()
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.ExecContext(ctx, `UPDATE tasks SET title = ?, complete = ?, due_date = ? WHERE id=?`, task.Title, task.Complete, task.DueDate.UTC().Unix(), task.Id)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}
func (t *SqlLiteTodoRepository) UpdateTasks(tasks []data.Task) error {
	ctx := context.TODO()
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	for _, task := range tasks {
		_, err = tx.ExecContext(ctx, `UPDATE tasks SET title = ?, complete = ?, due_date = ? WHERE id=?`, task.Title, task.Complete, task.DueDate.UTC().Unix(), task.Id)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	return err
}
func (t *SqlLiteTodoRepository) DeleteTaskById(id int) error {
	ctx := context.TODO()
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.ExecContext(ctx, `DELETE FROM tasks WHERE id=?`, id)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

func (t *SqlLiteTodoRepository) Close() error {
	return t.db.Close()
}

func NewTodoRepository() TodoRepository {
	var repository TodoRepository = &SqlLiteTodoRepository{db: newDB()}
	return repository
}
