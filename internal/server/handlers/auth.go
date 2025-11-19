// Package handlers содержит HTTP обработчики для API.
package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/victor/gophkeeper/internal/common/models"
	"github.com/victor/gophkeeper/internal/common/protocol"
	"github.com/victor/gophkeeper/internal/server/crypto"
	"github.com/victor/gophkeeper/internal/server/storage"
)

// AuthHandler обрабатывает запросы аутентификации.
type AuthHandler struct {
	storage storage.Storage
}

// NewAuthHandler создает новый AuthHandler.
func NewAuthHandler(storage storage.Storage) *AuthHandler {
	return &AuthHandler{storage: storage}
}

// Register обрабатывает регистрацию нового пользователя.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed", protocol.ErrorCodeInvalidRequest)
		return
	}

	var req models.CredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "invalid request body", protocol.ErrorCodeInvalidRequest)
		return
	}

	// Валидация
	if req.Login == "" || req.Password == "" {
		writeErrorResponse(w, http.StatusBadRequest, "login and password are required", protocol.ErrorCodeInvalidRequest)
		return
	}

	if len(req.Password) < 8 {
		writeErrorResponse(w, http.StatusBadRequest, "password must be at least 8 characters", protocol.ErrorCodeInvalidRequest)
		return
	}

	// Хешируем пароль
	passwordHash, err := crypto.HashPassword(req.Password)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "failed to hash password", protocol.ErrorCodeInternalError)
		return
	}

	// Создаем пользователя
	ctx := r.Context()
	user, err := h.storage.CreateUser(ctx, req.Login, passwordHash)
	if err != nil {
		if err == storage.ErrUserExists {
			writeErrorResponse(w, http.StatusConflict, "user already exists", protocol.ErrorCodeUserExists)
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "failed to create user", protocol.ErrorCodeInternalError)
		return
	}

	// Генерируем токены
	token, err := crypto.GenerateJWT(user.ID)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "failed to generate token", protocol.ErrorCodeInternalError)
		return
	}

	refreshToken, err := crypto.GenerateRefreshToken()
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "failed to generate refresh token", protocol.ErrorCodeInternalError)
		return
	}

	// Сохраняем refresh token
	expiresAt := time.Now().Add(crypto.RefreshTokenExpirationTime)
	if err := h.storage.CreateSession(ctx, user.ID, refreshToken, expiresAt); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "failed to create session", protocol.ErrorCodeInternalError)
		return
	}

	// Возвращаем ответ
	response := models.AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    int(crypto.JWTExpirationTime.Seconds()),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Login обрабатывает вход пользователя.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed", protocol.ErrorCodeInvalidRequest)
		return
	}

	var req models.CredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "invalid request body", protocol.ErrorCodeInvalidRequest)
		return
	}

	// Получаем пользователя с паролем
	ctx := r.Context()
	userWithPassword, err := h.storage.GetUserWithPassword(ctx, req.Login)
	if err != nil {
		if err == storage.ErrUserNotFound {
			writeErrorResponse(w, http.StatusUnauthorized, "invalid credentials", protocol.ErrorCodeInvalidCredentials)
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "failed to get user", protocol.ErrorCodeInternalError)
		return
	}

	// Проверяем пароль
	if !crypto.VerifyPassword(userWithPassword.PasswordHash, req.Password) {
		writeErrorResponse(w, http.StatusUnauthorized, "invalid credentials", protocol.ErrorCodeInvalidCredentials)
		return
	}

	// Генерируем токены
	token, err := crypto.GenerateJWT(userWithPassword.User.ID)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "failed to generate token", protocol.ErrorCodeInternalError)
		return
	}

	refreshToken, err := crypto.GenerateRefreshToken()
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "failed to generate refresh token", protocol.ErrorCodeInternalError)
		return
	}

	// Сохраняем refresh token
	expiresAt := time.Now().Add(crypto.RefreshTokenExpirationTime)
	if err := h.storage.CreateSession(ctx, userWithPassword.User.ID, refreshToken, expiresAt); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "failed to create session", protocol.ErrorCodeInternalError)
		return
	}

	// Возвращаем ответ
	response := models.AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    int(crypto.JWTExpirationTime.Seconds()),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Refresh обрабатывает обновление токена.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed", protocol.ErrorCodeInvalidRequest)
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "invalid request body", protocol.ErrorCodeInvalidRequest)
		return
	}

	// Получаем сессию
	ctx := r.Context()
	session, err := h.storage.GetSession(ctx, req.RefreshToken)
	if err != nil {
		if err == storage.ErrSessionNotFound {
			writeErrorResponse(w, http.StatusUnauthorized, "invalid refresh token", protocol.ErrorCodeInvalidToken)
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "failed to get session", protocol.ErrorCodeInternalError)
		return
	}

	// Проверяем срок действия
	if time.Now().After(session.ExpiresAt) {
		writeErrorResponse(w, http.StatusUnauthorized, "refresh token expired", protocol.ErrorCodeInvalidToken)
		return
	}

	// Генерируем новый JWT
	token, err := crypto.GenerateJWT(session.UserID)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "failed to generate token", protocol.ErrorCodeInternalError)
		return
	}

	// Возвращаем ответ
	response := models.AuthResponse{
		Token:        token,
		RefreshToken: req.RefreshToken, // Возвращаем тот же refresh token
		ExpiresIn:    int(crypto.JWTExpirationTime.Seconds()),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Logout обрабатывает выход пользователя.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed", protocol.ErrorCodeInvalidRequest)
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "invalid request body", protocol.ErrorCodeInvalidRequest)
		return
	}

	// Удаляем сессию
	ctx := r.Context()
	if err := h.storage.DeleteSession(ctx, req.RefreshToken); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "failed to delete session", protocol.ErrorCodeInternalError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// writeErrorResponse записывает JSON ответ с ошибкой.
func writeErrorResponse(w http.ResponseWriter, statusCode int, message string, code protocol.ErrorCode) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := models.ErrorResponse{
		Error: message,
		Code:  string(code),
	}
	json.NewEncoder(w).Encode(response)
}
