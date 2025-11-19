// Package middleware предоставляет middleware для HTTP обработчиков.
package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/victor/gophkeeper/internal/common/protocol"
	"github.com/victor/gophkeeper/internal/server/crypto"
)

type contextKey string

const userIDKey contextKey = "user_id"

// UserIDKey экспортируется для использования в тестах
var UserIDKey = userIDKey

// AuthMiddleware проверяет JWT токен и добавляет user_id в контекст.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeErrorResponse(w, http.StatusUnauthorized, "missing authorization header", protocol.ErrorCodeUnauthorized)
			return
		}

		// Проверяем формат "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			writeErrorResponse(w, http.StatusUnauthorized, "invalid authorization header format", protocol.ErrorCodeUnauthorized)
			return
		}

		token := parts[1]
		claims, err := crypto.ValidateJWT(token)
		if err != nil {
			writeErrorResponse(w, http.StatusUnauthorized, "invalid token", protocol.ErrorCodeInvalidToken)
			return
		}

		// Добавляем user_id в контекст
		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID извлекает user_id из контекста.
func GetUserID(ctx context.Context) string {
	userID, ok := ctx.Value(userIDKey).(string)
	if !ok {
		return ""
	}
	return userID
}

// writeErrorResponse записывает JSON ответ с ошибкой.
func writeErrorResponse(w http.ResponseWriter, statusCode int, message string, code protocol.ErrorCode) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := map[string]interface{}{
		"error": message,
		"code":  string(code),
	}
	json.NewEncoder(w).Encode(response)
}
