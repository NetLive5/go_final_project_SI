package db

import (
	"database/sql"
	"fmt"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func AddTask(task *Task) (int64, error) {
	if task == nil {
		return 0, fmt.Errorf("задача не должна быть nil")
	}
	result, err := database.Exec(
		`INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`,
		task.Date,
		task.Title,
		task.Comment,
		task.Repeat,
	)
	if err != nil {
		return 0, fmt.Errorf("добавление задачи: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("получение идентификатора добавленной задачи: %w", err)
	}
	return id, nil
}

func GetTask(id string) (*Task, error) {
	if id == "" {
		return nil, fmt.Errorf("не указан идентификатор задачи")
	}
	task := new(Task)
	err := database.QueryRow(
		`SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`,
		id,
	).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("задача не найдена")
		}
		return nil, fmt.Errorf("получение задачи: %w", err)
	}
	return task, nil
}

func UpdateTask(task *Task) error {
	if task == nil {
		return fmt.Errorf("задача не должна быть nil")
	}
	if task.ID == "" {
		return fmt.Errorf("не указан идентификатор задачи")
	}
	result, err := database.Exec(
		`UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`,
		task.Date,
		task.Title,
		task.Comment,
		task.Repeat,
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("обновление задачи: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("проверка обновлённой задачи: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("задача не найдена")
	}
	return nil
}

func UpdateDate(date, id string) error {
	if id == "" {
		return fmt.Errorf("не указан идентификатор задачи")
	}
	result, err := database.Exec(`UPDATE scheduler SET date = ? WHERE id = ?`, date, id)
	if err != nil {
		return fmt.Errorf("обновление даты задачи: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("проверка обновлённой даты задачи: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("задача не найдена")
	}
	return nil
}

func DeleteTask(id string) error {
	if id == "" {
		return fmt.Errorf("не указан идентификатор задачи")
	}
	result, err := database.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("удаление задачи: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("проверка удаления задачи: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("задача не найдена")
	}
	return nil
}

func Tasks(limit int) ([]*Task, error) {
	return queryTasks(`SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date, id LIMIT ?`, limit)
}

func TasksByText(limit int, search string) ([]*Task, error) {
	pattern := "%" + search + "%"
	return queryTasks(
		`SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date, id LIMIT ?`,
		pattern,
		pattern,
		limit,
	)
}

func TasksByDate(limit int, date string) ([]*Task, error) {
	return queryTasks(
		`SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? ORDER BY date, id LIMIT ?`,
		date,
		limit,
	)
}

func queryTasks(query string, args ...any) ([]*Task, error) {
	rows, err := database.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("запрос списка задач: %w", err)
	}
	defer rows.Close()

	tasks := make([]*Task, 0)
	for rows.Next() {
		task := new(Task)
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, fmt.Errorf("чтение задачи из результата запроса: %w", err)
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("чтение списка задач: %w", err)
	}
	return tasks, nil
}
