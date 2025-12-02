package crypto

import (
	"testing"
	"time"
)

func init() {
	// Инициализируем JWT конфигурацию для тестов
	SetJWTConfig(
		"test-secret-key-for-testing-only-min-32-chars",
		time.Hour,
		7*24*time.Hour,
	)
}

func TestGenerateJWT(t *testing.T) {
	userID := "test-user-id"
	token, err := GenerateJWT(userID)
	if err != nil {
		t.Fatalf("GenerateJWT failed: %v", err)
	}

	if token == "" {
		t.Error("GenerateJWT returned empty token")
	}
}

func TestValidateJWT(t *testing.T) {
	userID := "test-user-id"
	token, err := GenerateJWT(userID)
	if err != nil {
		t.Fatalf("GenerateJWT failed: %v", err)
	}

	claims, err := ValidateJWT(token)
	if err != nil {
		t.Fatalf("ValidateJWT failed: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("ValidateJWT: got userID %s, want %s", claims.UserID, userID)
	}
}

func TestValidateJWT_InvalidToken(t *testing.T) {
	invalidToken := "invalid.token.here"
	_, err := ValidateJWT(invalidToken)
	if err == nil {
		t.Error("ValidateJWT should fail for invalid token")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	token1, err1 := GenerateRefreshToken()
	if err1 != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err1)
	}

	if token1 == "" {
		t.Error("GenerateRefreshToken returned empty token")
	}

	// Токены должны быть разными
	token2, err2 := GenerateRefreshToken()
	if err2 != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err2)
	}

	if token1 == token2 {
		t.Error("GenerateRefreshToken returned same token twice")
	}
}
