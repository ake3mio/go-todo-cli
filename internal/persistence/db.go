package persistence

import (
	"context"
	"database/sql"
	_ "embed"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schema string

func newDB() *sql.DB {
	db, err := sql.Open("sqlite3", "file:todo.sqlite?cache=shared")
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
			_ = tx.Rollback() // best effort
		}
	}()

	_, err = tx.ExecContext(ctx, `INSERT INTO tasks (title, due_date) VALUES (?, ?)`, title, dueDate.UTC().Unix())
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

func NewTodoRepository() *TodoRepository {
	var repository TodoRepository = &SqlLiteTodoRepository{db: newDB()}
	return &repository
}
