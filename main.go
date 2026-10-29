package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"go_final_project/pkg/config"
	"go_final_project/pkg/db"
	"go_final_project/pkg/server"
)

func main() {
	logger := log.New(log.Writer(), "MAIN: ", log.LstdFlags)
	if err := run(logger); err != nil {
		logger.Print(err)
		os.Exit(1)
	}
}

func run(logger *log.Logger) error {
	_ = godotenv.Load(".env")

	cfg := config.Init()
	if err := db.Init(cfg.DBFile); err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Printf("ошибка закрытия базы данных: %v", err)
		}
	}()

	logger.Printf("Планировщик задач запущен на :%s", cfg.Port)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := server.Run(ctx, cfg, "web"); err != nil {
		return err
	}
	return nil
}
