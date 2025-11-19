// Package crypto предоставляет функции для криптографических операций на клиенте.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"errors"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// SaltSize определяет размер соли для PBKDF2.
	SaltSize = 32
	// NonceSize определяет размер nonce для AES-GCM.
	NonceSize = 12
	// KeySize определяет размер ключа для AES-256.
	KeySize = 32
	// PBKDF2Iterations определяет количество итераций для PBKDF2.
	PBKDF2Iterations = 100000
)

// DecryptData расшифровывает данные используя AES-256-GCM.
func DecryptData(encryptedData []byte, masterPassword string) ([]byte, error) {
	if len(encryptedData) < SaltSize+NonceSize {
		return nil, errors.New("encrypted data too short")
	}

	// Извлекаем соль, nonce и ciphertext
	salt := encryptedData[:SaltSize]
	nonce := encryptedData[SaltSize : SaltSize+NonceSize]
	ciphertext := encryptedData[SaltSize+NonceSize:]

	// Генерируем ключ из мастер-пароля
	key := pbkdf2.Key([]byte(masterPassword), salt, PBKDF2Iterations, KeySize, sha256.New)

	// Создаем cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Создаем GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Расшифровываем данные
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

