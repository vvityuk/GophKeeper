package crypto

import "testing"

func TestEncryptDecryptData(t *testing.T) {
	masterPassword := "test_master_password"
	originalData := []byte("sensitive data to encrypt")

	// Шифруем
	encrypted, err := EncryptData(originalData, masterPassword)
	if err != nil {
		t.Fatalf("EncryptData failed: %v", err)
	}

	if len(encrypted) == 0 {
		t.Error("EncryptData returned empty data")
	}

	// Расшифровываем
	decrypted, err := DecryptData(encrypted, masterPassword)
	if err != nil {
		t.Fatalf("DecryptData failed: %v", err)
	}

	if string(decrypted) != string(originalData) {
		t.Errorf("Decrypted data doesn't match original: got %s, want %s", string(decrypted), string(originalData))
	}
}

func TestEncryptDecryptData_DifferentPasswords(t *testing.T) {
	masterPassword1 := "password1"
	masterPassword2 := "password2"
	originalData := []byte("sensitive data")

	encrypted, err := EncryptData(originalData, masterPassword1)
	if err != nil {
		t.Fatalf("EncryptData failed: %v", err)
	}

	// Попытка расшифровать с другим паролем должна провалиться
	_, err = DecryptData(encrypted, masterPassword2)
	if err == nil {
		t.Error("DecryptData succeeded with wrong password")
	}
}

func TestEncryptData_DifferentEncryptions(t *testing.T) {
	masterPassword := "test_password"
	originalData := []byte("data")

	encrypted1, err1 := EncryptData(originalData, masterPassword)
	encrypted2, err2 := EncryptData(originalData, masterPassword)

	if err1 != nil || err2 != nil {
		t.Fatalf("EncryptData failed: %v, %v", err1, err2)
	}

	// Зашифрованные данные должны быть разными из-за случайной соли и nonce
	if string(encrypted1) == string(encrypted2) {
		t.Error("EncryptData returned same encrypted data (should have different salt/nonce)")
	}

	// Но оба должны расшифровываться одинаково
	decrypted1, _ := DecryptData(encrypted1, masterPassword)
	decrypted2, _ := DecryptData(encrypted2, masterPassword)

	if string(decrypted1) != string(originalData) || string(decrypted2) != string(originalData) {
		t.Error("Both encrypted versions should decrypt to same original data")
	}
}

