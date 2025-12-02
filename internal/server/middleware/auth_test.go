package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/victor/gophkeeper/internal/server/crypto"
)

func init() {
	// Инициализируем JWT конфигурацию для тестов
	crypto.SetJWTConfig(
		"test-secret-key-for-testing-only-min-32-chars",
		time.Hour,
		7*24*time.Hour,
	)
}

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedUserID string
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer " + generateTestToken(t, "test-user-id"),
			expectedStatus: http.StatusOK,
			expectedUserID: "test-user-id",
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid format - no Bearer",
			authHeader:     "Token test-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid format - no space",
			authHeader:     "Bearertest-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "empty token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userID := GetUserID(r.Context())
				if userID != tt.expectedUserID {
					t.Errorf("expected userID %s, got %s", tt.expectedUserID, userID)
				}
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected string
	}{
		{
			name:     "userID in context",
			ctx:      context.WithValue(context.Background(), userIDKey, "test-user-id"),
			expected: "test-user-id",
		},
		{
			name:     "no userID in context",
			ctx:      context.Background(),
			expected: "",
		},
		{
			name:     "wrong type in context",
			ctx:      context.WithValue(context.Background(), userIDKey, 123),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetUserID(tt.ctx)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func generateTestToken(t *testing.T, userID string) string {
	token, err := crypto.GenerateJWT(userID)
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}
	return token
}
