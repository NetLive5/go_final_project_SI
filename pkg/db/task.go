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

// Tasks returns up to limit tasks ordered by their date.
func Tasks(limit int) ([]*Task, error) {
	return queryTasks(`SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date, id LIMIT ?`, limit)
}

// TasksByText searches task titles and comments for a substring.
func TasksByText(limit int, search string) ([]*Task, error) {
	pattern := "%" + search + "%"
	return queryTasks(
		`SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date, id LIMIT ?`,
		pattern,
		pattern,
		limit,
	)
}

// TasksByDate returns tasks scheduled on date in YYYYMMDD format.
func TasksByDate(limit int, date string) ([]*Task, error) {
	return queryTasks(
		`SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? ORDER BY date, id LIMIT ?`,
		date,
		limit,
	)
}

func queryTasks(query string, args ...any) ([]*Task, error) {
	if database == nil {
		return nil, fmt.Errorf("database is not initialized")
	}

	rows, err := database.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]*Task, 0)
	for rows.Next() {
		task := new(Task)
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read tasks: %w", err)
	}
	return tasks, nil
}
