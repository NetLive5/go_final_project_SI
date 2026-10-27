package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const authTokenLifetime = 8 * time.Hour

var jwtHeader = base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

type jwtClaims struct {
	PasswordHash string `json:"password_hash"`
	ExpiresAt    int64  `json:"exp"`
}

func auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		password := currentPassword()
		if password == "" {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie("token")
		if err != nil || !validToken(cookie.Value, password, time.Now()) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func signinHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	configuredPassword := currentPassword()
	if configuredPassword == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password authentication is not configured"})
		return
	}

	var request struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if subtle.ConstantTimeCompare([]byte(request.Password), []byte(configuredPassword)) != 1 {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "wrong password"})
		return
	}

	token, err := createToken(configuredPassword, time.Now().Add(authTokenLifetime))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func createToken(password string, expiresAt time.Time) (string, error) {
	claims := jwtClaims{
		PasswordHash: passwordHash(password),
		ExpiresAt:    expiresAt.Unix(),
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("encode token claims: %w", err)
	}
	payload := base64.RawURLEncoding.EncodeToString(claimsJSON)
	message := jwtHeader + "." + payload
	mac := hmac.New(sha256.New, []byte(password))
	_, _ = mac.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return message + "." + signature, nil
}

func validToken(token, password string, now time.Time) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != jwtHeader {
		return false
	}

	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(password))
	_, _ = mac.Write([]byte(parts[0] + "." + parts[1]))
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return false
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	var claims jwtClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return false
	}
	if claims.ExpiresAt <= now.Unix() {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(claims.PasswordHash), []byte(passwordHash(password))) == 1
}

func currentPassword() string {
	return os.Getenv("TODO_PASSWORD")
}

func passwordHash(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}
