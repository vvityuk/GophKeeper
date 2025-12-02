package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// JWTSecretKey используется для подписи JWT токенов.
	// Должен быть установлен через SetJWTConfig.
	JWTSecretKey string
	// JWTExpirationTime определяет время жизни JWT токена.
	JWTExpirationTime time.Duration
	// RefreshTokenExpirationTime определяет время жизни refresh токена.
	RefreshTokenExpirationTime time.Duration
)

// SetJWTConfig устанавливает конфигурацию JWT.
func SetJWTConfig(secretKey string, expirationTime, refreshExpirationTime time.Duration) {
	JWTSecretKey = secretKey
	JWTExpirationTime = expirationTime
	RefreshTokenExpirationTime = refreshExpirationTime
}

// JWTClaims представляет claims для JWT токена.
type JWTClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateJWT генерирует JWT токен для пользователя.
func GenerateJWT(userID string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(JWTExpirationTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JWTSecretKey))
}

// ValidateJWT проверяет и парсит JWT токен.
func ValidateJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(JWTSecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GenerateRefreshToken генерирует случайный refresh token.
func GenerateRefreshToken() (string, error) {
	// Используем криптографически стойкий генератор случайных чисел
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}

	// Конвертируем в hex строку
	return hex.EncodeToString(token), nil
}
