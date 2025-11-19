package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/victor/gophkeeper/internal/common/models"
	"github.com/victor/gophkeeper/internal/server/storage"
)

// mockStorage реализует storage.Storage для тестирования
type mockStorage struct {
	users         map[string]*models.User
	usersPassword map[string]string
	sessions      map[string]*storage.Session
	dataRecords   map[string]*models.DataRecord
	userData      map[string][]*models.DataRecord
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		users:         make(map[string]*models.User),
		usersPassword: make(map[string]string),
		sessions:      make(map[string]*storage.Session),
		dataRecords:   make(map[string]*models.DataRecord),
		userData:      make(map[string][]*models.DataRecord),
	}
}

func (m *mockStorage) CreateUser(ctx context.Context, login string, passwordHash string) (*models.User, error) {
	for _, u := range m.users {
		if u.Login == login {
			return nil, storage.ErrUserExists
		}
	}
	user := &models.User{
		ID:        "user-" + login,
		Login:     login,
		CreatedAt: time.Now(),
	}
	m.users[user.ID] = user
	m.usersPassword[user.ID] = passwordHash
	return user, nil
}

func (m *mockStorage) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	for _, u := range m.users {
		if u.Login == login {
			return u, nil
		}
	}
	return nil, storage.ErrUserNotFound
}

func (m *mockStorage) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	user, ok := m.users[userID]
	if !ok {
		return nil, storage.ErrUserNotFound
	}
	return user, nil
}

func (m *mockStorage) GetUserWithPassword(ctx context.Context, login string) (*storage.UserWithPassword, error) {
	user, err := m.GetUserByLogin(ctx, login)
	if err != nil {
		return nil, err
	}
	return &storage.UserWithPassword{
		User:         user,
		PasswordHash: m.usersPassword[user.ID],
	}, nil
}

func (m *mockStorage) CreateSession(ctx context.Context, userID string, refreshToken string, expiresAt time.Time) error {
	session := &storage.Session{
		ID:           "session-" + refreshToken,
		UserID:       userID,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
	}
	m.sessions[refreshToken] = session
	return nil
}

func (m *mockStorage) GetSession(ctx context.Context, refreshToken string) (*storage.Session, error) {
	session, ok := m.sessions[refreshToken]
	if !ok {
		return nil, storage.ErrSessionNotFound
	}
	return session, nil
}

func (m *mockStorage) DeleteSession(ctx context.Context, refreshToken string) error {
	delete(m.sessions, refreshToken)
	return nil
}

func (m *mockStorage) DeleteUserSessions(ctx context.Context, userID string) error {
	for token, session := range m.sessions {
		if session.UserID == userID {
			delete(m.sessions, token)
		}
	}
	return nil
}

func (m *mockStorage) CreateDataRecord(ctx context.Context, userID string, record *models.DataRecord) error {
	m.dataRecords[record.ID] = record
	m.userData[userID] = append(m.userData[userID], record)
	return nil
}

func (m *mockStorage) GetDataRecord(ctx context.Context, recordID string, userID string) (*models.DataRecord, error) {
	record, ok := m.dataRecords[recordID]
	if !ok {
		return nil, storage.ErrRecordNotFound
	}
	return record, nil
}

func (m *mockStorage) GetAllDataRecords(ctx context.Context, userID string) ([]*models.DataRecord, error) {
	return m.userData[userID], nil
}

func (m *mockStorage) UpdateDataRecord(ctx context.Context, recordID string, userID string, record *models.DataRecord) error {
	existing, ok := m.dataRecords[recordID]
	if !ok {
		return storage.ErrRecordNotFound
	}
	// Симулируем поведение sqlite.go: увеличиваем версию и проверяем старую
	// record.Version содержит текущую версию (до увеличения)
	oldVersion := record.Version
	record.Version++
	// Проверка версии для оптимистичной блокировки (как в sqlite WHERE version = ?)
	// Проверяем, что существующая версия равна той, что была до увеличения
	if existing.Version != oldVersion {
		return storage.ErrVersionConflict
	}
	// Обновляем запись с новой версией
	m.dataRecords[recordID] = record
	return nil
}

func (m *mockStorage) DeleteDataRecord(ctx context.Context, recordID string, userID string) error {
	_, ok := m.dataRecords[recordID]
	if !ok {
		return storage.ErrRecordNotFound
	}
	delete(m.dataRecords, recordID)
	return nil
}

func (m *mockStorage) Close() error {
	return nil
}

func TestAuthHandler_Register(t *testing.T) {
	mockStorage := newMockStorage()
	handler := NewAuthHandler(mockStorage)

	tests := []struct {
		name           string
		method         string
		body           interface{}
		expectedStatus int
	}{
		{
			name:           "successful registration",
			method:         http.MethodPost,
			body:           models.CredentialsRequest{Login: "test@example.com", Password: "password123"},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "wrong method",
			method:         http.MethodGet,
			body:           models.CredentialsRequest{Login: "test@example.com", Password: "password123"},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "empty login",
			method:         http.MethodPost,
			body:           models.CredentialsRequest{Login: "", Password: "password123"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "short password",
			method:         http.MethodPost,
			body:           models.CredentialsRequest{Login: "test@example.com", Password: "short"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "duplicate user",
			method:         http.MethodPost,
			body:           models.CredentialsRequest{Login: "test@example.com", Password: "password123"},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(tt.method, "/api/v1/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Register(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusCreated {
				var resp models.AuthResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Errorf("failed to decode response: %v", err)
				}
				if resp.Token == "" {
					t.Error("token is empty")
				}
				if resp.RefreshToken == "" {
					t.Error("refresh token is empty")
				}
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	mockStorage := newMockStorage()
	handler := NewAuthHandler(mockStorage)

	// Создаем пользователя для теста
	req := models.CredentialsRequest{Login: "test@example.com", Password: "password123"}
	body, _ := json.Marshal(req)
	registerReq := httptest.NewRequest(http.MethodPost, "/api/v1/register", bytes.NewBuffer(body))
	registerReq.Header.Set("Content-Type", "application/json")
	registerW := httptest.NewRecorder()
	handler.Register(registerW, registerReq)

	tests := []struct {
		name           string
		method         string
		body           models.CredentialsRequest
		expectedStatus int
	}{
		{
			name:           "successful login",
			method:         http.MethodPost,
			body:           models.CredentialsRequest{Login: "test@example.com", Password: "password123"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "wrong method",
			method:         http.MethodGet,
			body:           models.CredentialsRequest{Login: "test@example.com", Password: "password123"},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "invalid credentials",
			method:         http.MethodPost,
			body:           models.CredentialsRequest{Login: "test@example.com", Password: "wrongpassword"},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "user not found",
			method:         http.MethodPost,
			body:           models.CredentialsRequest{Login: "nonexistent@example.com", Password: "password123"},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(tt.method, "/api/v1/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Login(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var resp models.AuthResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Errorf("failed to decode response: %v", err)
				}
				if resp.Token == "" {
					t.Error("token is empty")
				}
			}
		})
	}
}

func TestAuthHandler_Refresh(t *testing.T) {
	mockStorage := newMockStorage()
	handler := NewAuthHandler(mockStorage)

	// Создаем пользователя и сессию
	user, _ := mockStorage.CreateUser(context.Background(), "test@example.com", "hashed")
	refreshToken := "test-refresh-token"
	expiresAt := time.Now().Add(24 * time.Hour)
	mockStorage.CreateSession(context.Background(), user.ID, refreshToken, expiresAt)

	tests := []struct {
		name           string
		method         string
		refreshToken   string
		expectedStatus int
	}{
		{
			name:           "successful refresh",
			method:         http.MethodPost,
			refreshToken:   refreshToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "wrong method",
			method:         http.MethodGet,
			refreshToken:   refreshToken,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "invalid token",
			method:         http.MethodPost,
			refreshToken:   "invalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "expired token",
			method:         http.MethodPost,
			refreshToken:   "expired-token",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	// Создаем expired token
	expiredToken := "expired-token"
	expiredAt := time.Now().Add(-24 * time.Hour)
	mockStorage.CreateSession(context.Background(), user.ID, expiredToken, expiredAt)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(map[string]string{"refresh_token": tt.refreshToken})
			req := httptest.NewRequest(tt.method, "/api/v1/refresh", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Refresh(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var resp models.AuthResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Errorf("failed to decode response: %v", err)
				}
				if resp.Token == "" {
					t.Error("token is empty")
				}
			}
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	mockStorage := newMockStorage()
	handler := NewAuthHandler(mockStorage)

	// Создаем сессию
	user, _ := mockStorage.CreateUser(context.Background(), "test@example.com", "hashed")
	refreshToken := "test-refresh-token"
	expiresAt := time.Now().Add(24 * time.Hour)
	mockStorage.CreateSession(context.Background(), user.ID, refreshToken, expiresAt)

	tests := []struct {
		name           string
		method         string
		refreshToken   string
		expectedStatus int
	}{
		{
			name:           "successful logout",
			method:         http.MethodPost,
			refreshToken:   refreshToken,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "wrong method",
			method:         http.MethodGet,
			refreshToken:   refreshToken,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(map[string]string{"refresh_token": tt.refreshToken})
			req := httptest.NewRequest(tt.method, "/api/v1/logout", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Logout(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
