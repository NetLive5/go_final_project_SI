package config

import "os"

type Config struct {
	Port     string
	DBFile   string
	Password string
}

func Init() Config {
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}

	return Config{
		Port:     port,
		DBFile:   dbFile,
		Password: os.Getenv("TODO_PASSWORD"),
	}
}
