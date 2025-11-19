// Package crypto предоставляет функции для криптографических операций.
package crypto

import "golang.org/x/crypto/bcrypt"

const (
	// BcryptCost определяет сложность хеширования паролей.
	BcryptCost = 10
)

// HashPassword создает bcrypt хеш пароля.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword проверяет пароль против хеша.
func VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

