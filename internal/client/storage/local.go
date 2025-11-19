// Package storage предоставляет локальное хранилище для клиента.
package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/victor/gophkeeper/internal/common/models"
	_ "modernc.org/sqlite" // SQLite драйвер
)

// LocalStorage представляет локальное хранилище клиента.
type LocalStorage struct {
	db *sql.DB
}

// NewLocalStorage создает новое локальное хранилище.
func NewLocalStorage(dbPath string) (*LocalStorage, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	storage := &LocalStorage{db: db}

	if err := storage.initSchema(); err != nil {
		db.Close()
		return nil, err
	}

	return storage, nil
}

// initSchema создает необходимые таблицы в БД.
func (s *LocalStorage) initSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			token TEXT NOT NULL,
			refresh_token TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS data_records (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			name TEXT NOT NULL,
			metadata TEXT,
			encrypted_data TEXT NOT NULL,
			version INTEGER NOT NULL DEFAULT 1,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			synced INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE INDEX IF NOT EXISTS idx_data_records_synced ON data_records(synced)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

// SaveTokens сохраняет токены аутентификации.
func (s *LocalStorage) SaveTokens(ctx context.Context, token, refreshToken string) error {
	// Удаляем старые токены
	_, err := s.db.ExecContext(ctx, "DELETE FROM tokens")
	if err != nil {
		return err
	}

	// Сохраняем новые токены
	_, err = s.db.ExecContext(ctx,
		"INSERT INTO tokens (token, refresh_token, created_at) VALUES (?, ?, ?)",
		token, refreshToken, time.Now())
	return err
}

// GetTokens получает сохраненные токены.
func (s *LocalStorage) GetTokens(ctx context.Context) (token, refreshToken string, err error) {
	err = s.db.QueryRowContext(ctx,
		"SELECT token, refresh_token FROM tokens ORDER BY created_at DESC LIMIT 1").
		Scan(&token, &refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", ErrNoTokens
		}
		return "", "", err
	}
	return token, refreshToken, nil
}

// SaveDataRecord сохраняет запись данных локально.
func (s *LocalStorage) SaveDataRecord(ctx context.Context, record *models.DataRecord) error {
	now := time.Now()
	if record.CreatedAt.IsZero() {
		record.CreatedAt = now
	}
	if record.UpdatedAt.IsZero() {
		record.UpdatedAt = now
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO data_records 
		 (id, type, name, metadata, encrypted_data, version, created_at, updated_at, synced)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1)`,
		record.ID, record.Type, record.Name, record.Metadata, record.Data,
		record.Version, record.CreatedAt, record.UpdatedAt)
	return err
}

// GetDataRecord получает запись данных по ID.
func (s *LocalStorage) GetDataRecord(ctx context.Context, recordID string) (*models.DataRecord, error) {
	var record models.DataRecord
	err := s.db.QueryRowContext(ctx,
		`SELECT id, type, name, metadata, encrypted_data, version, created_at, updated_at
		 FROM data_records WHERE id = ?`,
		recordID).Scan(
		&record.ID, &record.Type, &record.Name, &record.Metadata, &record.Data,
		&record.Version, &record.CreatedAt, &record.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &record, nil
}

// GetAllDataRecords получает все записи данных.
func (s *LocalStorage) GetAllDataRecords(ctx context.Context) ([]*models.DataRecord, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, type, name, metadata, encrypted_data, version, created_at, updated_at
		 FROM data_records ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*models.DataRecord
	for rows.Next() {
		var record models.DataRecord
		if err := rows.Scan(
			&record.ID, &record.Type, &record.Name, &record.Metadata, &record.Data,
			&record.Version, &record.CreatedAt, &record.UpdatedAt); err != nil {
			return nil, err
		}
		records = append(records, &record)
	}

	return records, nil
}

// DeleteDataRecord удаляет запись данных.
func (s *LocalStorage) DeleteDataRecord(ctx context.Context, recordID string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM data_records WHERE id = ?", recordID)
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

// MarkUnsynced помечает запись как несинхронизированную.
func (s *LocalStorage) MarkUnsynced(ctx context.Context, recordID string) error {
	_, err := s.db.ExecContext(ctx, "UPDATE data_records SET synced = 0 WHERE id = ?", recordID)
	return err
}

// Close закрывает соединение с БД.
func (s *LocalStorage) Close() error {
	return s.db.Close()
}

var (
	// ErrNoTokens возвращается, когда токены не найдены.
	ErrNoTokens = errors.New("no tokens found")
	// ErrRecordNotFound возвращается, когда запись не найдена.
	ErrRecordNotFound = errors.New("record not found")
)
