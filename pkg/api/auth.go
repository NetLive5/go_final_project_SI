package api

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const authTokenLifetime = 8 * time.Hour

type authClaims struct {
	PasswordHash string `json:"password_hash"`
	jwt.RegisteredClaims
}

func auth(password string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if password == "" {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie("token")
		if err != nil || !validToken(cookie.Value, password, time.Now()) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "требуется авторизация"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func signinHandler(password string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "метод не поддерживается"})
			return
		}
		if password == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "парольная авторизация не настроена"})
			return
		}

		var request struct {
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "неверный формат JSON"})
			return
		}
		if subtle.ConstantTimeCompare([]byte(request.Password), []byte(password)) != 1 {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "неверный пароль"})
			return
		}

		token, err := createToken(password, time.Now().Add(authTokenLifetime))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"token": token})
	})
}

func createToken(password string, expiresAt time.Time) (string, error) {
	claims := authClaims{
		PasswordHash: passwordHash(password),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(password))
	if err != nil {
		return "", fmt.Errorf("не удалось подписать токен: %w", err)
	}
	return signed, nil
}

func validToken(token, password string, now time.Time) bool {
	claims := new(authClaims)
	parsed, err := jwt.ParseWithClaims(token, claims, func(_ *jwt.Token) (any, error) {
		return []byte(password), nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(func() time.Time { return now }),
	)
	if err != nil || !parsed.Valid {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(claims.PasswordHash), []byte(passwordHash(password))) == 1
}

func passwordHash(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}
