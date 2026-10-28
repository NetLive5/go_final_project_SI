package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	if err := prepareTests(); err != nil {
		fmt.Fprintln(os.Stderr, "Подготовка тестов:", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func prepareTests() error {
	if err := godotenv.Load("../.env"); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("загрузка .env: %w", err)
	}
	// Сервер запускается из корня проекта, а тесты работают из каталога tests.
	if dbFile := os.Getenv("TODO_DBFILE"); dbFile != "" && !filepath.IsAbs(dbFile) {
		absolutePath, err := filepath.Abs(filepath.Join("..", dbFile))
		if err != nil {
			return fmt.Errorf("путь к тестовой базе данных: %w", err)
		}
		if err := os.Setenv("TODO_DBFILE", absolutePath); err != nil {
			return fmt.Errorf("настройка пути к базе данных: %w", err)
		}
	}
	if Token != "" {
		return nil
	}
	var err error
	Token, err = loadToken()
	return err
}

func loadToken() (string, error) {
	if token := os.Getenv("TODO_TOKEN"); token != "" {
		return token, nil
	}
	password := os.Getenv("TODO_PASSWORD")
	if password == "" {
		return "", nil
	}

	requestBody, err := json.Marshal(map[string]string{"password": password})
	if err != nil {
		return "", fmt.Errorf("подготовка запроса авторизации: %w", err)
	}
	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Post(getURL("api/signin"), "application/json", bytes.NewReader(requestBody))
	if err != nil {
		return "", fmt.Errorf("получение токена (проверьте, что сервер запущен): %w", err)
	}
	defer response.Body.Close()

	var result struct {
		Token string `json:"token"`
		Error string `json:"error"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("чтение ответа авторизации: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("сервер отклонил авторизацию (HTTP %d): %s", response.StatusCode, result.Error)
	}
	if result.Token == "" {
		return "", fmt.Errorf("сервер не вернул токен авторизации")
	}
	return result.Token, nil
}
