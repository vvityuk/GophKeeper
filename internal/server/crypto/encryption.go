// Package crypto предоставляет криптографические функции для сервера.
// Использует общие функции из internal/common/crypto для шифрования данных.
package crypto

import (
	commoncrypto "github.com/victor/gophkeeper/internal/common/crypto"
)

// EncryptData шифрует данные используя AES-256-GCM.
// Перенаправляет вызов в common/crypto для переиспользования кода.
func EncryptData(data []byte, masterPassword string) ([]byte, error) {
	return commoncrypto.EncryptData(data, masterPassword)
}

// DecryptData расшифровывает данные используя AES-256-GCM.
// Перенаправляет вызов в common/crypto для переиспользования кода.
func DecryptData(encryptedData []byte, masterPassword string) ([]byte, error) {
	return commoncrypto.DecryptData(encryptedData, masterPassword)
}
