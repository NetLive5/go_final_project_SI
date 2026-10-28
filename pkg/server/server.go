package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go_final_project/pkg/api"
	"go_final_project/pkg/config"
)

func Run(ctx context.Context, cfg config.Config, webDir string) error {
	if cfg.Port == "" {
		return fmt.Errorf("не задан порт сервера")
	}

	mux := http.NewServeMux()
	api.Init(mux, cfg.Password)
	mux.Handle("/", http.FileServer(http.Dir(webDir)))

	server := &http.Server{Addr: ":" + cfg.Port, Handler: mux}
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			if closeErr := server.Close(); closeErr != nil {
				return fmt.Errorf("остановка HTTP-сервера: %w; закрытие соединений: %w", err, closeErr)
			}
			return fmt.Errorf("остановка HTTP-сервера: %w", err)
		}
		if err := <-serveErr; err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("работа HTTP-сервера: %w", err)
		}
		return nil
	}
}
