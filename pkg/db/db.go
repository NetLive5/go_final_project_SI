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


func Init(dbFile string) error {
	if dbFile == "" {
		return fmt.Errorf("database file must not be empty")
	}

	_, err := os.Stat(dbFile)
	install := os.IsNotExist(err)
	if err != nil && !install {
		return fmt.Errorf("check database file %q: %w", dbFile, err)
	}

	database, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("open database %q: %w", dbFile, err)
	}
	if err := database.Ping(); err != nil {
		_ = database.Close()
		database = nil
		return fmt.Errorf("connect to database %q: %w", dbFile, err)
	}

	if install {
		if _, err := database.Exec(schema); err != nil {
			_ = database.Close()
			database = nil
			return fmt.Errorf("initialize database %q: %w", dbFile, err)
		}
	}

	return nil
}
