// Package api предоставляет клиент для взаимодействия с сервером GophKeeper.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/victor/gophkeeper/internal/common/models"
	"github.com/victor/gophkeeper/internal/common/protocol"
)

// Client представляет клиент для работы с API сервера.
type Client struct {
	baseURL      string
	httpClient   *http.Client
	token        string
	refreshToken string
}

// NewClient создает новый API клиент.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetToken устанавливает JWT токен для аутентификации.
func (c *Client) SetToken(token string) {
	c.token = token
}

// SetRefreshToken устанавливает refresh токен.
func (c *Client) SetRefreshToken(refreshToken string) {
	c.refreshToken = refreshToken
}

// Register регистрирует нового пользователя.
func (c *Client) Register(ctx context.Context, login, password string) (*models.AuthResponse, error) {
	req := models.CredentialsRequest{
		Login:    login,
		Password: password,
	}

	var resp models.AuthResponse
	if err := c.doRequest(ctx, http.MethodPost, protocol.APIPrefix+"/register", req, &resp); err != nil {
		return nil, err
	}

	c.token = resp.Token
	c.refreshToken = resp.RefreshToken
	return &resp, nil
}

// Login выполняет вход пользователя.
func (c *Client) Login(ctx context.Context, login, password string) (*models.AuthResponse, error) {
	req := models.CredentialsRequest{
		Login:    login,
		Password: password,
	}

	var resp models.AuthResponse
	if err := c.doRequest(ctx, http.MethodPost, protocol.APIPrefix+"/login", req, &resp); err != nil {
		return nil, err
	}

	c.token = resp.Token
	c.refreshToken = resp.RefreshToken
	return &resp, nil
}

// Refresh обновляет JWT токен.
func (c *Client) Refresh(ctx context.Context) (*models.AuthResponse, error) {
	req := map[string]string{
		"refresh_token": c.refreshToken,
	}

	var resp models.AuthResponse
	if err := c.doRequest(ctx, http.MethodPost, protocol.APIPrefix+"/refresh", req, &resp); err != nil {
		return nil, err
	}

	c.token = resp.Token
	return &resp, nil
}

// Logout выполняет выход пользователя.
func (c *Client) Logout(ctx context.Context) error {
	req := map[string]string{
		"refresh_token": c.refreshToken,
	}

	return c.doRequest(ctx, http.MethodPost, protocol.APIPrefix+"/logout", req, nil)
}

// GetAllData получает все записи данных пользователя.
func (c *Client) GetAllData(ctx context.Context) ([]*models.DataRecord, error) {
	var records []*models.DataRecord
	if err := c.doRequest(ctx, http.MethodGet, protocol.APIPrefix+"/data", nil, &records); err != nil {
		return nil, err
	}
	return records, nil
}

// GetData получает конкретную запись данных.
func (c *Client) GetData(ctx context.Context, recordID string) (*models.DataRecord, error) {
	var record models.DataRecord
	url := fmt.Sprintf("%s/data/%s", protocol.APIPrefix, recordID)
	if err := c.doRequest(ctx, http.MethodGet, url, nil, &record); err != nil {
		return nil, err
	}
	return &record, nil
}

// CreateData создает новую запись данных.
func (c *Client) CreateData(ctx context.Context, req *models.CreateDataRequest) (*models.DataRecord, error) {
	var record models.DataRecord
	if err := c.doRequest(ctx, http.MethodPost, protocol.APIPrefix+"/data", req, &record); err != nil {
		return nil, err
	}
	return &record, nil
}

// UpdateData обновляет запись данных.
func (c *Client) UpdateData(ctx context.Context, recordID string, req *models.UpdateDataRequest) (*models.DataRecord, error) {
	var record models.DataRecord
	url := fmt.Sprintf("%s/data/%s", protocol.APIPrefix, recordID)
	if err := c.doRequest(ctx, http.MethodPut, url, req, &record); err != nil {
		return nil, err
	}
	return &record, nil
}

// DeleteData удаляет запись данных.
func (c *Client) DeleteData(ctx context.Context, recordID string) error {
	url := fmt.Sprintf("%s/data/%s", protocol.APIPrefix, recordID)
	return c.doRequest(ctx, http.MethodDelete, url, nil, nil)
}

// doRequest выполняет HTTP запрос к API.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp models.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			return fmt.Errorf("API error: %s (code: %s)", errResp.Error, errResp.Code)
		}
		return fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
