// Package storage предоставляет интерфейс и реализации хранилища данных.
package storage

import (
	"context"
	"time"

	"github.com/victor/gophkeeper/internal/common/models"
)

// Storage определяет интерфейс для работы с хранилищем данных.
type Storage interface {
	// User методы
	CreateUser(ctx context.Context, login string, passwordHash string) (*models.User, error)
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
	GetUserByID(ctx context.Context, userID string) (*models.User, error)
	GetUserWithPassword(ctx context.Context, login string) (*UserWithPassword, error)

	// Session методы
	CreateSession(ctx context.Context, userID string, refreshToken string, expiresAt time.Time) error
	GetSession(ctx context.Context, refreshToken string) (*Session, error)
	DeleteSession(ctx context.Context, refreshToken string) error
	DeleteUserSessions(ctx context.Context, userID string) error

	// Data методы
	CreateDataRecord(ctx context.Context, userID string, record *models.DataRecord) error
	GetDataRecord(ctx context.Context, recordID string, userID string) (*models.DataRecord, error)
	GetAllDataRecords(ctx context.Context, userID string) ([]*models.DataRecord, error)
	UpdateDataRecord(ctx context.Context, recordID string, userID string, record *models.DataRecord) error
	DeleteDataRecord(ctx context.Context, recordID string, userID string) error

	// Закрытие соединения
	Close() error
}

// UserWithPassword представляет пользователя с хешем пароля.
type UserWithPassword struct {
	User         *models.User
	PasswordHash string
}

// Session представляет сессию пользователя.
type Session struct {
	ID           string
	UserID       string
	RefreshToken string
	ExpiresAt    time.Time
	CreatedAt    time.Time
}
