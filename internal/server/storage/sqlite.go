// Package storage содержит реализацию хранилища на основе SQLite.
package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/victor/gophkeeper/internal/common/models"
	_ "modernc.org/sqlite" // SQLite драйвер
)

// SQLiteStorage реализует Storage интерфейс используя SQLite.
type SQLiteStorage struct {
	db *sql.DB
}

// NewSQLiteStorage создает новое хранилище SQLite и инициализирует схему БД.
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	storage := &SQLiteStorage{db: db}

	if err := storage.initSchema(); err != nil {
		db.Close()
		return nil, err
	}

	return storage, nil
}

// initSchema создает необходимые таблицы в БД.
func (s *SQLiteStorage) initSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			login TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			refresh_token TEXT UNIQUE NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS data_records (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			type TEXT NOT NULL,
			name TEXT NOT NULL,
			metadata TEXT,
			encrypted_data TEXT NOT NULL,
			version INTEGER NOT NULL DEFAULT 1,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_data_records_user_id ON data_records(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_refresh_token ON sessions(refresh_token)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

// CreateUser создает нового пользователя.
func (s *SQLiteStorage) CreateUser(ctx context.Context, login string, passwordHash string) (*models.User, error) {
	// Проверяем, существует ли пользователь
	_, err := s.GetUserByLogin(ctx, login)
	if err == nil {
		return nil, ErrUserExists
	}
	if err != ErrUserNotFound {
		return nil, err
	}

	id := uuid.New().String()
	now := time.Now()

	_, err = s.db.ExecContext(ctx,
		"INSERT INTO users (id, login, password_hash, created_at) VALUES (?, ?, ?, ?)",
		id, login, passwordHash, now)
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:        id,
		Login:     login,
		CreatedAt: now,
	}, nil
}

// GetUserByLogin получает пользователя по логину.
func (s *SQLiteStorage) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	var user models.User
	err := s.db.QueryRowContext(ctx,
		"SELECT id, login, created_at FROM users WHERE login = ?",
		login).Scan(&user.ID, &user.Login, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByID получает пользователя по ID.
func (s *SQLiteStorage) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	err := s.db.QueryRowContext(ctx,
		"SELECT id, login, created_at FROM users WHERE id = ?",
		userID).Scan(&user.ID, &user.Login, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetUserWithPassword получает пользователя с хешем пароля по логину.
func (s *SQLiteStorage) GetUserWithPassword(ctx context.Context, login string) (*UserWithPassword, error) {
	var user models.User
	var passwordHash string
	err := s.db.QueryRowContext(ctx,
		"SELECT id, login, password_hash, created_at FROM users WHERE login = ?",
		login).Scan(&user.ID, &user.Login, &passwordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &UserWithPassword{
		User:         &user,
		PasswordHash: passwordHash,
	}, nil
}

// CreateSession создает новую сессию пользователя.
func (s *SQLiteStorage) CreateSession(ctx context.Context, userID string, refreshToken string, expiresAt time.Time) error {
	id := uuid.New().String()
	now := time.Now()

	_, err := s.db.ExecContext(ctx,
		"INSERT INTO sessions (id, user_id, refresh_token, expires_at, created_at) VALUES (?, ?, ?, ?, ?)",
		id, userID, refreshToken, expiresAt, now)
	return err
}

// GetSession получает сессию по refresh token.
func (s *SQLiteStorage) GetSession(ctx context.Context, refreshToken string) (*Session, error) {
	var session Session
	err := s.db.QueryRowContext(ctx,
		"SELECT id, user_id, refresh_token, expires_at, created_at FROM sessions WHERE refresh_token = ?",
		refreshToken).Scan(&session.ID, &session.UserID, &session.RefreshToken, &session.ExpiresAt, &session.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	return &session, nil
}

// DeleteSession удаляет сессию по refresh token.
func (s *SQLiteStorage) DeleteSession(ctx context.Context, refreshToken string) error {
	_, err := s.db.ExecContext(ctx,
		"DELETE FROM sessions WHERE refresh_token = ?",
		refreshToken)
	return err
}

// DeleteUserSessions удаляет все сессии пользователя.
func (s *SQLiteStorage) DeleteUserSessions(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx,
		"DELETE FROM sessions WHERE user_id = ?",
		userID)
	return err
}

// CreateDataRecord создает новую запись данных.
func (s *SQLiteStorage) CreateDataRecord(ctx context.Context, userID string, record *models.DataRecord) error {
	if record.ID == "" {
		record.ID = uuid.New().String()
	}
	if record.Version == 0 {
		record.Version = 1
	}
	now := time.Now()
	if record.CreatedAt.IsZero() {
		record.CreatedAt = now
	}
	if record.UpdatedAt.IsZero() {
		record.UpdatedAt = now
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO data_records (id, user_id, type, name, metadata, encrypted_data, version, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.ID, userID, record.Type, record.Name, record.Metadata, record.Data, record.Version, record.CreatedAt, record.UpdatedAt)
	return err
}

// GetDataRecord получает запись данных по ID с проверкой принадлежности пользователю.
func (s *SQLiteStorage) GetDataRecord(ctx context.Context, recordID string, userID string) (*models.DataRecord, error) {
	var record models.DataRecord
	err := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, type, name, metadata, encrypted_data, version, created_at, updated_at
		 FROM data_records WHERE id = ? AND user_id = ?`,
		recordID, userID).Scan(
		&record.ID, &record.UserID, &record.Type, &record.Name, &record.Metadata,
		&record.Data, &record.Version, &record.CreatedAt, &record.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &record, nil
}

// GetAllDataRecords получает все записи данных пользователя.
func (s *SQLiteStorage) GetAllDataRecords(ctx context.Context, userID string) ([]*models.DataRecord, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, type, name, metadata, encrypted_data, version, created_at, updated_at
		 FROM data_records WHERE user_id = ? ORDER BY created_at DESC`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*models.DataRecord
	for rows.Next() {
		var record models.DataRecord
		if err := rows.Scan(
			&record.ID, &record.UserID, &record.Type, &record.Name, &record.Metadata,
			&record.Data, &record.Version, &record.CreatedAt, &record.UpdatedAt); err != nil {
			return nil, err
		}
		records = append(records, &record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

// UpdateDataRecord обновляет запись данных с проверкой версии (оптимистичная блокировка).
func (s *SQLiteStorage) UpdateDataRecord(ctx context.Context, recordID string, userID string, record *models.DataRecord) error {
	record.UpdatedAt = time.Now()
	record.Version++

	result, err := s.db.ExecContext(ctx,
		`UPDATE data_records 
		 SET name = ?, metadata = ?, encrypted_data = ?, version = ?, updated_at = ?
		 WHERE id = ? AND user_id = ? AND version = ?`,
		record.Name, record.Metadata, record.Data, record.Version, record.UpdatedAt,
		recordID, userID, record.Version-1)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrVersionConflict
	}

	return nil
}

// DeleteDataRecord удаляет запись данных.
func (s *SQLiteStorage) DeleteDataRecord(ctx context.Context, recordID string, userID string) error {
	result, err := s.db.ExecContext(ctx,
		"DELETE FROM data_records WHERE id = ? AND user_id = ?",
		recordID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// Close закрывает соединение с БД.
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}
