// Package crypto предоставляет функции для криптографических операций на клиенте.
// Использует общие функции из internal/common/crypto для переиспользования кода.
package crypto

import (
	commoncrypto "github.com/victor/gophkeeper/internal/common/crypto"
)

// DecryptData расшифровывает данные используя AES-256-GCM.
// Перенаправляет вызов в common/crypto для переиспользования кода.
func DecryptData(encryptedData []byte, masterPassword string) ([]byte, error) {
	return commoncrypto.DecryptData(encryptedData, masterPassword)
}
