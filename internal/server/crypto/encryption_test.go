// Package crypto тестирует обертки над common/crypto.
// Основные тесты находятся в internal/common/crypto/encryption_test.go
package crypto

import (
	"testing"

	commoncrypto "github.com/victor/gophkeeper/internal/common/crypto"
)

// TestEncryptDecryptData проверяет, что обертки работают корректно.
func TestEncryptDecryptData(t *testing.T) {
	masterPassword := "test_master_password"
	originalData := []byte("sensitive data to encrypt")

	// Шифруем через обертку сервера
	encrypted, err := EncryptData(originalData, masterPassword)
	if err != nil {
		t.Fatalf("EncryptData failed: %v", err)
	}

	// Расшифровываем через обертку сервера
	decrypted, err := DecryptData(encrypted, masterPassword)
	if err != nil {
		t.Fatalf("DecryptData failed: %v", err)
	}

	if string(decrypted) != string(originalData) {
		t.Errorf("Decrypted data doesn't match original: got %s, want %s", string(decrypted), string(originalData))
	}

	// Проверяем совместимость: данные, зашифрованные через common, должны расшифровываться через обертку
	encryptedCommon, _ := commoncrypto.EncryptData(originalData, masterPassword)
	decryptedWrapper, err := DecryptData(encryptedCommon, masterPassword)
	if err != nil {
		t.Fatalf("DecryptData failed with common encrypted data: %v", err)
	}
	if string(decryptedWrapper) != string(originalData) {
		t.Error("Wrapper DecryptData failed to decrypt data encrypted by common EncryptData")
	}
}
