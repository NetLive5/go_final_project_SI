package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE scheduler (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	date CHAR(8) NOT NULL DEFAULT "",
	title VARCHAR(256) NOT NULL DEFAULT "",
	comment TEXT NOT NULL DEFAULT "",
	repeat VARCHAR(128) NOT NULL DEFAULT ""
);
CREATE INDEX scheduler_date_idx ON scheduler (date);
`

var database *sql.DB

func Close() error {
	if database == nil {
		return nil
	}
	err := database.Close()
	database = nil
	if err != nil {
		return fmt.Errorf("закрытие базы данных: %w", err)
	}
	return nil
}

func Init(dbFile string) error {
	if dbFile == "" {
		return fmt.Errorf("не задан путь к файлу базы данных")
	}

	_, err := os.Stat(dbFile)
	install := os.IsNotExist(err)
	if err != nil && !install {
		return fmt.Errorf("проверка файла базы данных %q: %w", dbFile, err)
	}

	database, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("открытие базы данных %q: %w", dbFile, err)
	}
	if err := database.Ping(); err != nil {
		_ = database.Close()
		database = nil
		return fmt.Errorf("подключение к базе данных %q: %w", dbFile, err)
	}

	if install {
		if _, err := database.Exec(schema); err != nil {
			_ = database.Close()
			database = nil
			return fmt.Errorf("инициализация базы данных %q: %w", dbFile, err)
		}
	}

	return nil
}
