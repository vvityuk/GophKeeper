package crypto

import "testing"

func TestHashPassword(t *testing.T) {
	password := "test_password_123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Error("HashPassword returned empty hash")
	}

	if hash == password {
		t.Error("HashPassword returned plain password")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "test_password_123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if !VerifyPassword(hash, password) {
		t.Error("VerifyPassword failed for correct password")
	}

	if VerifyPassword(hash, "wrong_password") {
		t.Error("VerifyPassword succeeded for wrong password")
	}
}

func TestHashPassword_DifferentHashes(t *testing.T) {
	password := "test_password_123"
	hash1, err1 := HashPassword(password)
	hash2, err2 := HashPassword(password)

	if err1 != nil || err2 != nil {
		t.Fatalf("HashPassword failed: %v, %v", err1, err2)
	}

	// Хеши должны быть разными из-за случайной соли
	if hash1 == hash2 {
		t.Error("HashPassword returned same hash for same password (should have different salts)")
	}
}

