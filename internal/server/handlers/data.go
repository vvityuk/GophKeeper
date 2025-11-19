package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/victor/gophkeeper/internal/common/models"
	"github.com/victor/gophkeeper/internal/common/protocol"
	"github.com/victor/gophkeeper/internal/server/crypto"
	"github.com/victor/gophkeeper/internal/server/middleware"
	"github.com/victor/gophkeeper/internal/server/storage"
)

// DataHandler обрабатывает запросы для работы с данными.
type DataHandler struct {
	storage storage.Storage
}

// NewDataHandler создает новый DataHandler.
func NewDataHandler(storage storage.Storage) *DataHandler {
	return &DataHandler{storage: storage}
}

// GetAllData получает все записи данных пользователя.
func (h *DataHandler) GetAllData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed", protocol.ErrorCodeInvalidRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeErrorResponse(w, http.StatusUnauthorized, "user not authenticated", protocol.ErrorCodeUnauthorized)
		return
	}

	ctx := r.Context()
	records, err := h.storage.GetAllDataRecords(ctx, userID)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "failed to get data", protocol.ErrorCodeInternalError)
		return
	}

	// Расшифровываем данные (в реальном приложении нужен мастер-пароль от пользователя)
	// Для упрощения возвращаем зашифрованные данные - расшифровка на клиенте
	decryptedRecords := make([]*models.DataRecord, len(records))
	for i, record := range records {
		decryptedRecords[i] = record
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(decryptedRecords)
}

// GetData получает конкретную запись данных.
func (h *DataHandler) GetData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed", protocol.ErrorCodeInvalidRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeErrorResponse(w, http.StatusUnauthorized, "user not authenticated", protocol.ErrorCodeUnauthorized)
		return
	}

	// Извлекаем ID из пути
	recordID := r.PathValue("id")
	if recordID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "record id is required", protocol.ErrorCodeInvalidRequest)
		return
	}

	ctx := r.Context()
	record, err := h.storage.GetDataRecord(ctx, recordID, userID)
	if err != nil {
		if err == storage.ErrRecordNotFound {
			writeErrorResponse(w, http.StatusNotFound, "record not found", protocol.ErrorCodeNotFound)
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "failed to get record", protocol.ErrorCodeInternalError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(record)
}

// CreateData создает новую запись данных.
func (h *DataHandler) CreateData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed", protocol.ErrorCodeInvalidRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeErrorResponse(w, http.StatusUnauthorized, "user not authenticated", protocol.ErrorCodeUnauthorized)
		return
	}

	var req models.CreateDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "invalid request body", protocol.ErrorCodeInvalidRequest)
		return
	}

	// Валидация
	if !req.Type.IsValid() {
		writeErrorResponse(w, http.StatusBadRequest, "invalid data type", protocol.ErrorCodeInvalidRequest)
		return
	}

	if req.Name == "" {
		writeErrorResponse(w, http.StatusBadRequest, "name is required", protocol.ErrorCodeInvalidRequest)
		return
	}

	// Шифруем данные
	// В реальном приложении мастер-пароль должен передаваться отдельно или храниться на клиенте
	// Для упрощения используем фиксированный ключ (в production это должно быть иначе)
	masterPassword := "default-master-password" // TODO: получать от пользователя
	encryptedData, err := crypto.EncryptData([]byte(req.Data), masterPassword)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "failed to encrypt data", protocol.ErrorCodeInternalError)
		return
	}

	// Создаем запись
	now := time.Now()
	record := &models.DataRecord{
		ID:        uuid.New().String(),
		Type:      req.Type,
		Name:      req.Name,
		Metadata:  req.Metadata,
		Data:      string(encryptedData),
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	ctx := r.Context()
	if err := h.storage.CreateDataRecord(ctx, userID, record); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "failed to create record", protocol.ErrorCodeInternalError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(record)
}

// UpdateData обновляет запись данных.
func (h *DataHandler) UpdateData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed", protocol.ErrorCodeInvalidRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeErrorResponse(w, http.StatusUnauthorized, "user not authenticated", protocol.ErrorCodeUnauthorized)
		return
	}

	// Извлекаем ID из пути
	recordID := r.PathValue("id")
	if recordID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "record id is required", protocol.ErrorCodeInvalidRequest)
		return
	}

	var req models.UpdateDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "invalid request body", protocol.ErrorCodeInvalidRequest)
		return
	}

	// Получаем текущую запись
	ctx := r.Context()
	currentRecord, err := h.storage.GetDataRecord(ctx, recordID, userID)
	if err != nil {
		if err == storage.ErrRecordNotFound {
			writeErrorResponse(w, http.StatusNotFound, "record not found", protocol.ErrorCodeNotFound)
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "failed to get record", protocol.ErrorCodeInternalError)
		return
	}

	// Проверяем версию (оптимистичная блокировка)
	if currentRecord.Version != req.Version {
		writeErrorResponse(w, http.StatusConflict, "version conflict", protocol.ErrorCodeConflict)
		return
	}

	// Обновляем поля
	if req.Name != "" {
		currentRecord.Name = req.Name
	}
	if req.Metadata != "" {
		currentRecord.Metadata = req.Metadata
	}
	if req.Data != "" {
		// Шифруем новые данные
		masterPassword := "default-master-password" // TODO: получать от пользователя
		encryptedData, err := crypto.EncryptData([]byte(req.Data), masterPassword)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "failed to encrypt data", protocol.ErrorCodeInternalError)
			return
		}
		currentRecord.Data = string(encryptedData)
	}

	// Обновляем запись
	if err := h.storage.UpdateDataRecord(ctx, recordID, userID, currentRecord); err != nil {
		if err == storage.ErrVersionConflict {
			writeErrorResponse(w, http.StatusConflict, "version conflict", protocol.ErrorCodeConflict)
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "failed to update record", protocol.ErrorCodeInternalError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(currentRecord)
}

// DeleteData удаляет запись данных.
func (h *DataHandler) DeleteData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed", protocol.ErrorCodeInvalidRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeErrorResponse(w, http.StatusUnauthorized, "user not authenticated", protocol.ErrorCodeUnauthorized)
		return
	}

	// Извлекаем ID из пути
	recordID := r.PathValue("id")
	if recordID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "record id is required", protocol.ErrorCodeInvalidRequest)
		return
	}

	ctx := r.Context()
	if err := h.storage.DeleteDataRecord(ctx, recordID, userID); err != nil {
		if err == storage.ErrRecordNotFound {
			writeErrorResponse(w, http.StatusNotFound, "record not found", protocol.ErrorCodeNotFound)
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "failed to delete record", protocol.ErrorCodeInternalError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
