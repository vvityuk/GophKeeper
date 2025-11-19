package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/victor/gophkeeper/internal/common/models"
)

func setupTestStorage(t *testing.T) *SQLiteStorage {
	// Создаем временный файл БД
	tmpFile, err := os.CreateTemp("", "test_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFile.Close()

	storage, err := NewSQLiteStorage(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("failed to create storage: %v", err)
	}

	t.Cleanup(func() {
		storage.Close()
		os.Remove(tmpFile.Name())
	})

	return storage
}

func TestSQLiteStorage_CreateUser(t *testing.T) {
	storage := setupTestStorage(t)
	ctx := context.Background()

	user, err := storage.CreateUser(ctx, "test@example.com", "hashed_password")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if user.ID == "" {
		t.Error("CreateUser returned user without ID")
	}

	if user.Login != "test@example.com" {
		t.Errorf("CreateUser: got login %s, want test@example.com", user.Login)
	}
}

func TestSQLiteStorage_CreateUser_Duplicate(t *testing.T) {
	storage := setupTestStorage(t)
	ctx := context.Background()

	_, err := storage.CreateUser(ctx, "test@example.com", "hashed_password")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Попытка создать пользователя с тем же логином
	_, err = storage.CreateUser(ctx, "test@example.com", "another_hash")
	if err != ErrUserExists {
		t.Errorf("CreateUser: got error %v, want ErrUserExists", err)
	}
}

func TestSQLiteStorage_GetUserByLogin(t *testing.T) {
	storage := setupTestStorage(t)
	ctx := context.Background()

	created, err := storage.CreateUser(ctx, "test@example.com", "hashed_password")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	user, err := storage.GetUserByLogin(ctx, "test@example.com")
	if err != nil {
		t.Fatalf("GetUserByLogin failed: %v", err)
	}

	if user.ID != created.ID {
		t.Errorf("GetUserByLogin: got ID %s, want %s", user.ID, created.ID)
	}
}

func TestSQLiteStorage_GetUserByLogin_NotFound(t *testing.T) {
	storage := setupTestStorage(t)
	ctx := context.Background()

	_, err := storage.GetUserByLogin(ctx, "nonexistent@example.com")
	if err != ErrUserNotFound {
		t.Errorf("GetUserByLogin: got error %v, want ErrUserNotFound", err)
	}
}

func TestSQLiteStorage_CreateSession(t *testing.T) {
	storage := setupTestStorage(t)
	ctx := context.Background()

	user, err := storage.CreateUser(ctx, "test@example.com", "hashed_password")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	refreshToken := "test_refresh_token"
	expiresAt := time.Now().Add(24 * time.Hour)

	err = storage.CreateSession(ctx, user.ID, refreshToken, expiresAt)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
}

func TestSQLiteStorage_GetSession(t *testing.T) {
	storage := setupTestStorage(t)
	ctx := context.Background()

	user, err := storage.CreateUser(ctx, "test@example.com", "hashed_password")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	refreshToken := "test_refresh_token"
	expiresAt := time.Now().Add(24 * time.Hour)

	err = storage.CreateSession(ctx, user.ID, refreshToken, expiresAt)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	session, err := storage.GetSession(ctx, refreshToken)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}

	if session.UserID != user.ID {
		t.Errorf("GetSession: got userID %s, want %s", session.UserID, user.ID)
	}
}

func TestSQLiteStorage_CreateDataRecord(t *testing.T) {
	storage := setupTestStorage(t)
	ctx := context.Background()

	user, err := storage.CreateUser(ctx, "test@example.com", "hashed_password")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	record := &models.DataRecord{
		Type:     models.DataTypeCredentials,
		Name:     "Test Record",
		Metadata: "Test metadata",
		Data:     "encrypted_data",
		Version:  1,
	}

	err = storage.CreateDataRecord(ctx, user.ID, record)
	if err != nil {
		t.Fatalf("CreateDataRecord failed: %v", err)
	}

	if record.ID == "" {
		t.Error("CreateDataRecord: record ID not set")
	}
}

func TestSQLiteStorage_GetDataRecord(t *testing.T) {
	storage := setupTestStorage(t)
	ctx := context.Background()

	user, err := storage.CreateUser(ctx, "test@example.com", "hashed_password")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	record := &models.DataRecord{
		Type:     models.DataTypeCredentials,
		Name:     "Test Record",
		Metadata: "Test metadata",
		Data:     "encrypted_data",
		Version:  1,
	}

	err = storage.CreateDataRecord(ctx, user.ID, record)
	if err != nil {
		t.Fatalf("CreateDataRecord failed: %v", err)
	}

	retrieved, err := storage.GetDataRecord(ctx, record.ID, user.ID)
	if err != nil {
		t.Fatalf("GetDataRecord failed: %v", err)
	}

	if retrieved.Name != record.Name {
		t.Errorf("GetDataRecord: got name %s, want %s", retrieved.Name, record.Name)
	}
}

func TestSQLiteStorage_GetAllDataRecords(t *testing.T) {
	storage := setupTestStorage(t)
	ctx := context.Background()

	user, err := storage.CreateUser(ctx, "test@example.com", "hashed_password")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Создаем несколько записей
	for i := 0; i < 3; i++ {
		record := &models.DataRecord{
			Type:     models.DataTypeCredentials,
			Name:     "Test Record",
			Metadata: "Test metadata",
			Data:     "encrypted_data",
			Version:  1,
		}
		if err := storage.CreateDataRecord(ctx, user.ID, record); err != nil {
			t.Fatalf("CreateDataRecord failed: %v", err)
		}
	}

	records, err := storage.GetAllDataRecords(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetAllDataRecords failed: %v", err)
	}

	if len(records) != 3 {
		t.Errorf("GetAllDataRecords: got %d records, want 3", len(records))
	}
}

func TestSQLiteStorage_UpdateDataRecord(t *testing.T) {
	storage := setupTestStorage(t)
	ctx := context.Background()

	user, err := storage.CreateUser(ctx, "test@example.com", "hashed_password")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	record := &models.DataRecord{
		Type:     models.DataTypeCredentials,
		Name:     "Test Record",
		Metadata: "Test metadata",
		Data:     "encrypted_data",
		Version:  1,
	}

	err = storage.CreateDataRecord(ctx, user.ID, record)
	if err != nil {
		t.Fatalf("CreateDataRecord failed: %v", err)
	}

	record.Name = "Updated Name"
	err = storage.UpdateDataRecord(ctx, record.ID, user.ID, record)
	if err != nil {
		t.Fatalf("UpdateDataRecord failed: %v", err)
	}

	updated, err := storage.GetDataRecord(ctx, record.ID, user.ID)
	if err != nil {
		t.Fatalf("GetDataRecord failed: %v", err)
	}

	if updated.Name != "Updated Name" {
		t.Errorf("UpdateDataRecord: got name %s, want Updated Name", updated.Name)
	}

	if updated.Version != 2 {
		t.Errorf("UpdateDataRecord: got version %d, want 2", updated.Version)
	}
}

func TestSQLiteStorage_DeleteDataRecord(t *testing.T) {
	storage := setupTestStorage(t)
	ctx := context.Background()

	user, err := storage.CreateUser(ctx, "test@example.com", "hashed_password")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	record := &models.DataRecord{
		Type:     models.DataTypeCredentials,
		Name:     "Test Record",
		Metadata: "Test metadata",
		Data:     "encrypted_data",
		Version:  1,
	}

	err = storage.CreateDataRecord(ctx, user.ID, record)
	if err != nil {
		t.Fatalf("CreateDataRecord failed: %v", err)
	}

	err = storage.DeleteDataRecord(ctx, record.ID, user.ID)
	if err != nil {
		t.Fatalf("DeleteDataRecord failed: %v", err)
	}

	_, err = storage.GetDataRecord(ctx, record.ID, user.ID)
	if err != ErrRecordNotFound {
		t.Errorf("DeleteDataRecord: record still exists, got error %v", err)
	}
}
