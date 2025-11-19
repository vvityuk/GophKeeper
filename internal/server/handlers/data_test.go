package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/victor/gophkeeper/internal/common/models"
	"github.com/victor/gophkeeper/internal/server/middleware"
)

// setUserIDInContext устанавливает userID в контекст для тестов
// Использует тот же ключ, что и middleware
func setUserIDInContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, middleware.UserIDKey, userID)
}

func TestDataHandler_GetAllData(t *testing.T) {
	mockStorage := newMockStorage()
	handler := NewDataHandler(mockStorage)

	userID := "test-user-id"
	ctx := context.Background()

	// Создаем тестовые данные
	record1 := &models.DataRecord{
		ID:        "record-1",
		Type:      models.DataTypeCredentials,
		Name:      "Test Record 1",
		Metadata:  "metadata",
		Data:      "encrypted_data",
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	record2 := &models.DataRecord{
		ID:        "record-2",
		Type:      models.DataTypeText,
		Name:      "Test Record 2",
		Metadata:  "metadata2",
		Data:      "encrypted_data2",
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mockStorage.CreateDataRecord(ctx, userID, record1)
	mockStorage.CreateDataRecord(ctx, userID, record2)

	tests := []struct {
		name           string
		method         string
		userID         string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "successful get all",
			method:         http.MethodGet,
			userID:         userID,
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "wrong method",
			method:         http.MethodPost,
			userID:         userID,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "no user id",
			method:         http.MethodGet,
			userID:         "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/data", nil)
			if tt.userID != "" {
				req = req.WithContext(setUserIDInContext(req.Context(), tt.userID))
			}
			w := httptest.NewRecorder()

			handler.GetAllData(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var records []*models.DataRecord
				if err := json.NewDecoder(w.Body).Decode(&records); err != nil {
					t.Errorf("failed to decode response: %v", err)
				}
				if len(records) != tt.expectedCount {
					t.Errorf("expected %d records, got %d", tt.expectedCount, len(records))
				}
			}
		})
	}
}

func TestDataHandler_GetData(t *testing.T) {
	mockStorage := newMockStorage()
	handler := NewDataHandler(mockStorage)

	userID := "test-user-id"
	recordID := "record-1"
	ctx := context.Background()

	record := &models.DataRecord{
		ID:        recordID,
		Type:      models.DataTypeCredentials,
		Name:      "Test Record",
		Metadata:  "metadata",
		Data:      "encrypted_data",
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mockStorage.CreateDataRecord(ctx, userID, record)

	tests := []struct {
		name           string
		method         string
		recordID       string
		userID         string
		expectedStatus int
	}{
		{
			name:           "successful get",
			method:         http.MethodGet,
			recordID:       recordID,
			userID:         userID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "wrong method",
			method:         http.MethodPost,
			recordID:       recordID,
			userID:         userID,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "no user id",
			method:         http.MethodGet,
			recordID:       recordID,
			userID:         "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "record not found",
			method:         http.MethodGet,
			recordID:       "nonexistent",
			userID:         userID,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/data/"+tt.recordID, nil)
			req.SetPathValue("id", tt.recordID)
			if tt.userID != "" {
				req = req.WithContext(setUserIDInContext(req.Context(), tt.userID))
			}
			w := httptest.NewRecorder()

			handler.GetData(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var record models.DataRecord
				if err := json.NewDecoder(w.Body).Decode(&record); err != nil {
					t.Errorf("failed to decode response: %v", err)
				}
				if record.ID != tt.recordID {
					t.Errorf("expected record ID %s, got %s", tt.recordID, record.ID)
				}
			}
		})
	}
}

func TestDataHandler_CreateData(t *testing.T) {
	mockStorage := newMockStorage()
	handler := NewDataHandler(mockStorage)

	userID := "test-user-id"

	tests := []struct {
		name           string
		method         string
		body           models.CreateDataRequest
		userID         string
		expectedStatus int
	}{
		{
			name:   "successful create",
			method: http.MethodPost,
			body: models.CreateDataRequest{
				Type:     models.DataTypeCredentials,
				Name:     "Test Record",
				Metadata: "metadata",
				Data:     "test data",
			},
			userID:         userID,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "wrong method",
			method:         http.MethodGet,
			body:           models.CreateDataRequest{},
			userID:         userID,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "no user id",
			method:         http.MethodPost,
			body:           models.CreateDataRequest{},
			userID:         "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "invalid data type",
			method: http.MethodPost,
			body: models.CreateDataRequest{
				Type: models.DataType("INVALID"),
				Name: "Test",
				Data: "data",
			},
			userID:         userID,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "empty name",
			method: http.MethodPost,
			body: models.CreateDataRequest{
				Type: models.DataTypeCredentials,
				Name: "",
				Data: "data",
			},
			userID:         userID,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(tt.method, "/api/v1/data", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			if tt.userID != "" {
				req = req.WithContext(setUserIDInContext(req.Context(), tt.userID))
			}
			w := httptest.NewRecorder()

			handler.CreateData(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusCreated {
				var record models.DataRecord
				if err := json.NewDecoder(w.Body).Decode(&record); err != nil {
					t.Errorf("failed to decode response: %v", err)
				}
				if record.ID == "" {
					t.Error("record ID is empty")
				}
			}
		})
	}
}

func TestDataHandler_UpdateData(t *testing.T) {
	mockStorage := newMockStorage()
	handler := NewDataHandler(mockStorage)

	userID := "test-user-id"
	recordID := "record-1"
	ctx := context.Background()

	record := &models.DataRecord{
		ID:        recordID,
		Type:      models.DataTypeCredentials,
		Name:      "Test Record",
		Metadata:  "metadata",
		Data:      "encrypted_data",
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mockStorage.CreateDataRecord(ctx, userID, record)

	tests := []struct {
		name           string
		method         string
		recordID       string
		body           models.UpdateDataRequest
		userID         string
		expectedStatus int
	}{
		{
			name:     "successful update",
			method:   http.MethodPut,
			recordID: recordID,
			body: models.UpdateDataRequest{
				Name:    "Updated Name",
				Data:    "updated data",
				Version: 1,
			},
			userID:         userID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "wrong method",
			method:         http.MethodGet,
			recordID:       recordID,
			body:           models.UpdateDataRequest{},
			userID:         userID,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "record not found",
			method:         http.MethodPut,
			recordID:       "nonexistent",
			body:           models.UpdateDataRequest{Version: 1},
			userID:         userID,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:     "version conflict",
			method:   http.MethodPut,
			recordID: recordID,
			body: models.UpdateDataRequest{
				Version: 999, // Неправильная версия
			},
			userID:         userID,
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(tt.method, "/api/v1/data/"+tt.recordID, bytes.NewBuffer(body))
			req.SetPathValue("id", tt.recordID)
			req.Header.Set("Content-Type", "application/json")
			if tt.userID != "" {
				req = req.WithContext(setUserIDInContext(req.Context(), tt.userID))
			}
			w := httptest.NewRecorder()

			handler.UpdateData(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestDataHandler_DeleteData(t *testing.T) {
	mockStorage := newMockStorage()
	handler := NewDataHandler(mockStorage)

	userID := "test-user-id"
	recordID := "record-1"
	ctx := context.Background()

	record := &models.DataRecord{
		ID:        recordID,
		Type:      models.DataTypeCredentials,
		Name:      "Test Record",
		Metadata:  "metadata",
		Data:      "encrypted_data",
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mockStorage.CreateDataRecord(ctx, userID, record)

	tests := []struct {
		name           string
		method         string
		recordID       string
		userID         string
		expectedStatus int
	}{
		{
			name:           "successful delete",
			method:         http.MethodDelete,
			recordID:       recordID,
			userID:         userID,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "wrong method",
			method:         http.MethodGet,
			recordID:       recordID,
			userID:         userID,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "record not found",
			method:         http.MethodDelete,
			recordID:       "nonexistent",
			userID:         userID,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/data/"+tt.recordID, nil)
			req.SetPathValue("id", tt.recordID)
			if tt.userID != "" {
				req = req.WithContext(setUserIDInContext(req.Context(), tt.userID))
			}
			w := httptest.NewRecorder()

			handler.DeleteData(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
