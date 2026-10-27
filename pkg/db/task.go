package db

import (
	"fmt"
)

// Task contains the fields stored in the scheduler table.
type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// AddTask inserts a task and returns its database identifier.
func AddTask(task *Task) (int64, error) {
	if task == nil {
		return 0, fmt.Errorf("task must not be nil")
	}
	if database == nil {
		return 0, fmt.Errorf("database is not initialized")
	}

	result, err := database.Exec(
		`INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`,
		task.Date,
		task.Title,
		task.Comment,
		task.Repeat,
	)
	if err != nil {
		return 0, fmt.Errorf("insert task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get inserted task id: %w", err)
	}
	return id, nil
}
