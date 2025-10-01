package data

import "time"

type Task struct {
	Id       int
	Title    string
	Complete bool
	DueDate  time.Time
}
