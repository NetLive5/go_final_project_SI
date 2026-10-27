package main

import (
	"log"

	"github.com/joho/godotenv"
	"go_final_project/pkg/config"
	"go_final_project/pkg/db"
	"go_final_project/pkg/server"
)

func main() {
	logger := log.New(log.Writer(), "MAIN: ", log.LstdFlags)
	_ = godotenv.Load("tests/.env")
	_ = godotenv.Load("tests\\.env")
	_ = godotenv.Load(".env")

	cfg := config.Init()
	if err := db.Init(cfg.DBFile); err != nil {
		logger.Fatal(err)
	}
	logger.Printf("Starting task scheduler on :%s", cfg.Port)
	if err := server.Run(cfg.Port, "web"); err != nil {
		logger.Fatal(err)
	}
}
