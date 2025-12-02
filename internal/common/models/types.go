// Package models содержит общие модели данных для клиента и сервера.
package models

import "time"

// DataType представляет тип хранимых данных.
type DataType string

const (
	// DataTypeCredentials представляет пары логин/пароль.
	DataTypeCredentials DataType = "CREDENTIALS"
	// DataTypeText представляет произвольные текстовые данные.
	DataTypeText DataType = "TEXT"
	// DataTypeBinary представляет произвольные бинарные данные.
	DataTypeBinary DataType = "BINARY"
	// DataTypeCard представляет данные банковских карт.
	DataTypeCard DataType = "CARD"
)

// IsValid проверяет, является ли тип данных валидным.
func (dt DataType) IsValid() bool {
	return dt == DataTypeCredentials ||
		dt == DataTypeText ||
		dt == DataTypeBinary ||
		dt == DataTypeCard
}

// DataRecord представляет запись данных пользователя.
type DataRecord struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id,omitempty"`
	Type        DataType  `json:"type"`
	Name        string    `json:"name"`
	Metadata    string    `json:"metadata"`
	Data        string    `json:"data"` // Зашифрованные данные
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// User представляет пользователя системы.
type User struct {
	ID        string    `json:"id"`
	Login     string    `json:"login"`
	CreatedAt time.Time `json:"created_at"`
}

// CredentialsRequest представляет запрос на регистрацию или вход.
type CredentialsRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// AuthResponse представляет ответ с токенами аутентификации.
type AuthResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // В секундах
}

// ErrorResponse представляет ответ с ошибкой.
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

// CreateDataRequest представляет запрос на создание записи данных.
type CreateDataRequest struct {
	Type     DataType `json:"type"`
	Name     string   `json:"name"`
	Metadata string   `json:"metadata"`
	Data     string   `json:"data"` // Незашифрованные данные
}

// UpdateDataRequest представляет запрос на обновление записи данных.
type UpdateDataRequest struct {
	Name     string `json:"name,omitempty"`
	Metadata string `json:"metadata,omitempty"`
	Data     string `json:"data,omitempty"` // Незашифрованные данные
	Version  int    `json:"version"`       // Текущая версия для оптимистичной блокировки
}

